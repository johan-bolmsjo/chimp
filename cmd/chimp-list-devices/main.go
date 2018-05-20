package main

import (
	"fmt"
	"os"

	"github.com/johan-bolmsjo/chimp"
)

func main() {
	devices, err := chimp.ListDevices()
	if err != nil {
		fatalf("Failed to list devices, %s\n")
	}

	if len(devices) == 0 {
		fmt.Println("No identified supported devices.")
		return
	}

	fmt.Printf("Identified supported devices:\n\nID Type       Name\n")
	for i, device := range devices {
		fmt.Printf("%2d %-10s %q\n", i, device.Type, device.Name)
	}
}

func fatalf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}
