package chimp

// DeviceInfo contains some brief device information that can be inspected
// before deciding to open a device.
type DeviceInfo struct {
	Name string                 // Name of device.
	Type DeviceType             // Device type.
	Open func() (Device, error) // Function that opens the device.
}

// DeviceType is an enumeration of basic device types such as "Mouse", "Tablet" etc.
type DeviceType int

//go:generate stringer -type=DeviceType -trimprefix=DeviceType

const (
	DeviceTypeTablet DeviceType = iota
	DeviceTypeMouse
)

// Device is any opened input device.
type Device interface {
	// Properties returns properties of device.
	Properties() Properties

	// Capabilities returns capabilities of device.
	Capabilities() *Capabilities

	// Read event from device.
	// TODO: Define how to handle errors, reopening device?
	Read() (Event, error)

	// Close device.
	Close()
}

// ListDevices lists available input devices that may be opened and used.
func ListDevices() ([]DeviceInfo, error) {
	return listDevices()
}
