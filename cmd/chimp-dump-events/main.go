package main

import (
	"fmt"
	"os"

	"github.com/johan-bolmsjo/chimp"
)

func main() {
	devices, err := chimp.ListDevices()
	if err != nil {
		fatalf("Failed to list devices, error: %s\n")
	}

	if len(devices) == 0 {
		fmt.Println("No identified supported devices.")
		return
	}

	deviceInfo := devices[0]
	fmt.Printf("Opening %s %q\n", deviceInfo.Type, deviceInfo.Name)

	device, err := deviceInfo.Open()
	if err != nil {
		fatalf("Failed to open device %q, error: %s\n", deviceInfo.Name, err)
	}

	fmt.Println(device.Properties())
	fmt.Println(device.Capabilities())

	for {
		event, err := device.Read()
		if err != nil {
			fatalf("Failed to read event from device %q, error: %s\n", deviceInfo.Name, err)
		}
		fmt.Println(event)
	}
}

func fatalf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}
