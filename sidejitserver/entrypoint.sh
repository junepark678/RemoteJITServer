#!/bin/sh

pymobiledevice3 remote tunneld --host 0.0.0.0 &

SideJITServer --port 8080 -n
