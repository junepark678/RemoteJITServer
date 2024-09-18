package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"

	"golang.org/x/sys/unix"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/ipc"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	ENV_WG_TUN_FD             = "WG_TUN_FD"
	ENV_WG_UAPI_FD            = "WG_UAPI_FD"
	ENV_WG_PROCESS_FOREGROUND = "WG_PROCESS_FOREGROUND"
)

func RunWireGuard() {
	interfaceName := "wg0"
	tdev, err := func() (tun.Device, error) {
		tunFdStr := os.Getenv(ENV_WG_TUN_FD)
		if tunFdStr == "" {
			return tun.CreateTUN(interfaceName, device.DefaultMTU)
		}

		// construct tun device from supplied fd

		fd, err := strconv.ParseUint(tunFdStr, 10, 32)
		if err != nil {
			return nil, err
		}

		err = unix.SetNonblock(int(fd), true)
		if err != nil {
			return nil, err
		}

		file := os.NewFile(uintptr(fd), "")
		return tun.CreateTUNFromFile(file, device.DefaultMTU)
	}()

	if err == nil {
		realInterfaceName, err2 := tdev.Name()
		if err2 == nil {
			interfaceName = realInterfaceName
		}
	}
	logger := device.NewLogger(
		device.LogLevelVerbose,
		fmt.Sprintf("(%s) ", interfaceName),
	)

	logger.Verbosef("Starting wireguard-go version %s", "0.0.20230223")

	if err != nil {
		logger.Errorf("Failed to create TUN device: %v", err)
		os.Exit(1)
	}

	// open UAPI file (or use supplied fd)

	fileUAPI, err := func() (*os.File, error) {
		uapiFdStr := os.Getenv(ENV_WG_UAPI_FD)
		if uapiFdStr == "" {
			return ipc.UAPIOpen(interfaceName)
		}

		// use supplied fd

		fd, err := strconv.ParseUint(uapiFdStr, 10, 32)
		if err != nil {
			return nil, err
		}

		return os.NewFile(uintptr(fd), ""), nil
	}()
	if err != nil {
		logger.Errorf("UAPI listen error: %v", err)
		os.Exit(1)
		return
	}

	device := device.NewDevice(tdev, conn.NewDefaultBind(), logger)

	logger.Verbosef("Device started")

	errs := make(chan error)
	term := make(chan os.Signal, 1)

	uapi, err := ipc.UAPIListen(interfaceName, fileUAPI)
	if err != nil {
		logger.Errorf("Failed to listen on uapi socket: %v", err)
		os.Exit(1)
	}

	go func() {
		for {
			conn, err := uapi.Accept()
			if err != nil {
				errs <- err
				return
			}
			go device.IpcHandle(conn)
		}
	}()

	logger.Verbosef("UAPI listener started")

	// wait for program to terminate

	signal.Notify(term, unix.SIGTERM)
	signal.Notify(term, os.Interrupt)

	interfacePrivate, err := os.ReadFile("./interfaceKey")
	if err != nil {
		privateKey, err := wgtypes.GeneratePrivateKey()
		if err != nil {
			fmt.Println("Failed to generate private key")
			os.Exit(1)
		}
		interfacePrivate = []byte(privateKey.String())
		err = os.WriteFile("./interfaceKey", interfacePrivate, 0644)
		if err != nil {
			fmt.Println("Failed to write private key to file")
			os.Exit(1)
		}
	}

	interfacePrivateKey, err := wgtypes.ParseKey(string(interfacePrivate))
	if err != nil {
		fmt.Println("Failed to parse private key")
		os.Exit(1)
	}

	// read from ./device.json
	// if it exists, use it to configure the device
	// if it doesn't exist, configure the device with the default values
	// write the device configuration to ./device.json

	var deviceConfig wgtypes.Device
	data, err := os.ReadFile("./device.json")
	json.Unmarshal(data, &deviceConfig)

	control, err := wgctrl.New()
	if err != nil {
		fmt.Println("Failed to create wgctrl client")
		os.Exit(1)
	}

	listenPort := 51820

	if err == nil {
		control.ConfigureDevice("wg0", wgtypes.Config{
			PrivateKey:   &interfacePrivateKey,
			ListenPort:   &listenPort,
			ReplacePeers: true,
		})
	} else {
		peers := make([]wgtypes.PeerConfig, len(deviceConfig.Peers))
		for i, peer := range deviceConfig.Peers {
			peers[i] = wgtypes.PeerConfig{
				PublicKey:  peer.PublicKey,
				AllowedIPs: peer.AllowedIPs,
			}
			for _, ip := range peer.AllowedIPs {
				exec.Command("ip", "addr", "add", ip.String(), "dev", "wg0").Run()
			}
		}
		control.ConfigureDevice("wg0", wgtypes.Config{
			PrivateKey:   &interfacePrivateKey,
			ListenPort:   &listenPort,
			ReplacePeers: true,
			Peers:        peers,
		})
	}

	fmt.Println(os.Getwd())

	exec.Command("ip", "addr", "add", "10.12.0.69/32", "dev", "wg0").Run()

	select {
	case <-term:
	case <-errs:
	case <-device.Wait():
	}

	// clean up

	uapi.Close()
	device.Close()

	logger.Verbosef("Shutting down")
}
