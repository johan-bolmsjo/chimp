package chimp

import (
	"fmt"
	"strings"
)

// Capabilities of device.
type Capabilities struct {
	PositionDevices []PositionDevice
	Buttons         []Button
}

// HasButton checks if device has position device.
func (cap *Capabilities) HasPositionDevice(device PositionDevice) bool {
	for _, v := range cap.PositionDevices {
		if v == device {
			return true
		}
	}
	return false
}

// HasButton checks if device has button.
func (cap *Capabilities) HasButton(button Button) bool {
	for _, v := range cap.Buttons {
		if v == button {
			return true
		}
	}
	return false
}

func (cap *Capabilities) String() string {
	var positionDeviceNames, buttonNames []string

	for _, v := range cap.PositionDevices {
		positionDeviceNames = append(positionDeviceNames, v.String())
	}
	for _, v := range cap.Buttons {
		buttonNames = append(buttonNames, v.String())
	}
	return fmt.Sprintf(fmtCapabilities, strings.Join(positionDeviceNames, " "), strings.Join(buttonNames, " "))
}

const fmtCapabilities = `Capabilities: {
    PositionDevices: [%s]
    Buttons:         [%s]
}`
