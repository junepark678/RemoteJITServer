package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/exec"
	"text/template"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Config struct {
	PrivateKey string
	Address    string
	PublicKey  string
	Endpoint   string
}

func GenerateRandomIPv6(prefix string) (net.IP, error) {
	// Parse the /48 prefix
	ip, ipNet, err := net.ParseCIDR(prefix)
	if err != nil {
		return nil, err
	}

	// Ensure the prefix is 48 bits (6 bytes)
	if ones, _ := ipNet.Mask.Size(); ones != 48 {
		return nil, fmt.Errorf("not a /48 subnet")
	}

	// Copy the 48-bit prefix into a 16-byte slice for the IPv6 address
	randomIP := make(net.IP, len(ip))
	copy(randomIP, ip)

	// Randomly generate the remaining 80 bits
	randomBytes := make([]byte, 10) // 80 bits = 10 bytes
	_, err = rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	// Fill the last 80 bits of the address with random values
	copy(randomIP[6:], randomBytes)

	return randomIP, nil
}

func main() {
	// try reading from ./interfaceKey
	// if it exists, use it as the private key
	// if it doesn't exist, generate a new private key
	// write the private key to ./interfaceKey

	go func() {
		err := exec.Command("ip", "link", "add", "dev", "wg0", "type", "wireguard").Run()
		if err != nil {
			fmt.Println("Failed to create wireguard interface")
			os.Exit(1)
		}

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
		if err != nil {
			fmt.Println("Failed to read device configuration")
			os.Exit(1)
		}
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
					if exec.Command("ip", "addr", "add", ip.String(), "dev", "wg0").Run() != nil {
						fmt.Println("Failed to add IP address to interface")
						os.Exit(1)
					}
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

		if exec.Command("ip", "addr", "add", "10.12.0.69/32", "dev", "wg0").Run() != nil {
			fmt.Println("Failed to add IP address to interface")
			os.Exit(1)
		}

		if exec.Command("ip", "link", "set", "up", "dev", "wg0").Run() != nil {
			fmt.Println("Failed to bring up interface")
			os.Exit(1)
		}
	}()

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

	control, err := wgctrl.New()
	if err != nil {
		fmt.Println("Failed to create wgctrl client")
		os.Exit(1)
	}

	listenPort := 51820

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 418
		w.WriteHeader(418)
		w.Write([]byte("I'm a teapot"))
	})
	// Handle /config
	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		// load config template from ./template.conf

		templateFile := "./template.conf"
		t, err := template.ParseFiles(templateFile)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Internal server error"))
			return
		}

		// execute template with config values
		privateKey, err := wgtypes.GeneratePrivateKey()
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Internal server error"))
			return
		}

		a, err := rand.Int(rand.Reader, big.NewInt(254))
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Internal server error"))
			return
		}
		b, err := rand.Int(rand.Reader, big.NewInt(254))
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Internal server error"))
			return
		}
		address := net.IPv4(10, 12, byte(a.Int64()+1), byte(b.Int64()+1))

		config := Config{
			PrivateKey: privateKey.String(),
			Address:    address.String(),
			PublicKey:  interfacePrivateKey.PublicKey().String(),
			Endpoint:   fmt.Sprintf("%s:51820", os.Getenv("HOSTNAME")),
		}
		// writer to string
		err = t.Execute(w, config)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Internal server error"))
			return
		}
		control.ConfigureDevice("wg0", wgtypes.Config{
			PrivateKey:   &interfacePrivateKey,
			ListenPort:   &listenPort,
			ReplacePeers: false,
			Peers: []wgtypes.PeerConfig{
				{
					PublicKey: privateKey.PublicKey(),
					AllowedIPs: []net.IPNet{{
						IP:   address,
						Mask: net.CIDRMask(32, 32),
					}},
				},
			}})

		device, err := control.Device("wg0")
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Internal server error"))
			return
		}

		data, err := json.Marshal(device)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Internal server error"))
			return
		}

		os.WriteFile("./device.json", data, 0644)

		peers := make([]wgtypes.PeerConfig, len(device.Peers)+1)
		deviceConfig, err := control.Device("wg0")
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Internal server error"))
			return
		}
		for i, peer := range deviceConfig.Peers {
			peers[i] = wgtypes.PeerConfig{
				PublicKey:  peer.PublicKey,
				AllowedIPs: peer.AllowedIPs,
			}
		}
		peers[len(peers)-1] = wgtypes.PeerConfig{
			PublicKey: privateKey.PublicKey(),
			AllowedIPs: []net.IPNet{{
				IP:   address,
				Mask: net.CIDRMask(32, 32),
			}},
		}
		control.ConfigureDevice("wg0", wgtypes.Config{
			PrivateKey:   &interfacePrivateKey,
			ListenPort:   &listenPort,
			ReplacePeers: true,
			Peers:        peers,
		})

		exec.Command("ip", "route", "add", address.String()+"/32", "dev", "wg0").Run()

		return
	})

	http.ListenAndServe(":6969", nil)
}
