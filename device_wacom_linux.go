package chimp

import (
	"fmt"
	"regexp"

	evdev "github.com/johan-bolmsjo/golang-evdev"
)

type deviceMatcherWacomBamboo16FG6x8 struct {
	reName *regexp.Regexp
}

func newDeviceMatcherWacomBamboo16FG6x8() *deviceMatcherWacomBamboo16FG6x8 {
	return &deviceMatcherWacomBamboo16FG6x8{
		reName: regexp.MustCompile(`^(Wacom Bamboo 16FG 6x8) (Pen|Finger|Pad)$`),
	}
}

func (matcher *deviceMatcherWacomBamboo16FG6x8) match(devInfo linuxDeviceInfo) (match bool, logicalID string) {
	reMatchName := matcher.reName.FindStringSubmatch(devInfo.name)
	reMatchPhys := reUSBDeviceID.FindStringSubmatch(devInfo.phys)

	if reMatchName != nil && reMatchPhys != nil {
		match = true
		logicalID = reMatchName[1] + " " + reMatchPhys[1]
	}

	return
}

func (matcher *deviceMatcherWacomBamboo16FG6x8) newLogicalDevice() logicalDevice {
	return &logicalDeviceWacomBamboo16FG6x8{
		reName: matcher.reName,
	}
}

type logicalDeviceWacomBamboo16FG6x8 struct {
	reName       *regexp.Regexp
	linuxDevices [wacomLinuxDeviceTypes]linuxDeviceInfo
}

func (logicalDevice *logicalDeviceWacomBamboo16FG6x8) addLinuxDevice(devInfo linuxDeviceInfo) {
	reMatchName := logicalDevice.reName.FindStringSubmatch(devInfo.name)
	if reMatchName != nil {
		logicalDevice.linuxDevices[wacomLinuxDeviceTypeFromString(reMatchName[2])] = devInfo
	}
}

func (logicalDevice *logicalDeviceWacomBamboo16FG6x8) deviceInfo() DeviceInfo {
	return DeviceInfo{
		Name: wacomBamboo16FG6x8Properties[PropertyDeviceName].String(),
		Type: DeviceTypeTablet,
		Open: logicalDevice.Open,
	}
}

func (logicalDevice *logicalDeviceWacomBamboo16FG6x8) Open() (Device, error) {
	var err error
	var inputDevices [wacomLinuxDeviceTypes]*evdev.InputDevice

	closeInputDevices := func() {
		for _, v := range inputDevices {
			if v != nil {
				v.File.Close()
			}
		}
	}

	for i, v := range logicalDevice.linuxDevices {
		if v.dev != "" {
			if inputDevices[i], err = v.openDevice(); err != nil {
				closeInputDevices()
				return nil, err
			}
		}
	}

	return newWacomDevice(inputDevices, wacomBamboo16FG6x8Properties,
		wacomBamboo16FG6x8Capabilities, wacomBamboo16FG6x8DeviceParams), nil
}

type wacomDevice struct {
	eventMux
	properties   Properties
	capabilities Capabilities
	params       wacomDeviceParams

	// Recorded state that is used to produce an event when SYN_REPORT is observed.
	state struct {
		penCoord           Coord2D
		penTool            Button // ButtonPenTip or ButtonPenEraser
		penToolSelected    bool
		penInputEventFlags inputEventFlag // Flags about content of one event group
		penDistance        float32
		penPressure        float32

		// Generate button events after any positioning events.
		// Keep them in a side structure for this purpose.
		penButtonEvents []Event

		fingerCoord           Coord2D
		fingerTouchPressure   float32
		fingerInputEventFlags inputEventFlag // Flags about content of one event group
	}
}

func (dev *wacomDevice) Properties() Properties {
	return dev.properties
}

func (dev *wacomDevice) Capabilities() *Capabilities {
	return &dev.capabilities
}

func newWacomDevice(inputDevices [wacomLinuxDeviceTypes]*evdev.InputDevice, properties Properties,
	capabilities Capabilities, params wacomDeviceParams) *wacomDevice {

	dev := &wacomDevice{
		eventMux:     newEventMux(),
		properties:   properties,
		capabilities: capabilities,
		params:       params,
	}

	funs := [wacomLinuxDeviceTypes]inputEventFunc{
		// matches order of wacomLinuxDeviceType
		dev.inputEventPen,
		dev.inputEventFinger,
		dev.inputEventPad,
	}

	for i, v := range inputDevices {
		if v != nil {
			dev.addEventSource(v, funs[i])
		}
	}
	return dev
}

// Can be used when adding support for a device to see what Linux input events are available.
func dumpInputEvents(inputEvents []evdev.InputEvent) {
	for i, v := range inputEvents {
		fmt.Printf("%02d: %s\n", i, &v)
	}
}

func (dev *wacomDevice) inputEventPen(inputEvents []evdev.InputEvent) (events []Event) {
	for _, v := range inputEvents {
		switch v.Type {
		case evdev.EV_SYN:
			switch v.Code {
			case evdev.SYN_REPORT:
				emitPressureEvent := dev.state.penInputEventFlags.has(inputEventFlagPressure)
				if emitPressureEvent && dev.state.penPressure > 0 {
					// The Linux device driver seems to be able to generate a
					// positive pressure event and a positive distance in the
					// same event group. This is in my view not logically
					// possible. Clear the distance if a positive pressure was
					// in the event group.
					dev.state.penDistance = 0
				}

				// Emit pressure (button) event before position movement.
				// If events are consumed in order and position movement triggers
				// some draw operation or similar it may be better to have adjusted
				// the pressure beforehand. Note that all events form the same event
				// group have the same timestamp so ordering is not really important
				// if it is used by the application.
				if emitPressureEvent {
					events = append(events, &EventButton{
						Timestamp: inputEventTime(&v),
						Button:    dev.state.penTool,
						Pressure:  dev.state.penPressure,
					})
				}

				// The position event is not generated if no pen tool is selected.
				// It seems the Linux driver generates a {X: 0, Y: 0, Distance: 0}
				// event when the tool leaves the detectable range of the pad. We
				// don't want this.
				if dev.state.penInputEventFlags.has(inputEventFlagPosition) &&
					dev.state.penToolSelected {

					events = append(events, &EventPositionPen{
						Timestamp: inputEventTime(&v),
						Coord:     dev.state.penCoord,
						Distance:  dev.state.penDistance,
					})
				}
				for _, event := range dev.state.penButtonEvents {
					events = append(events, event)
				}

				dev.state.penButtonEvents = dev.state.penButtonEvents[:0]
				dev.state.penInputEventFlags = 0
			case evdev.SYN_DROPPED:
				// TODO(jb): See comment in inputEventPad about this condition.
				return []Event{newEventError(eventBufferOverrunError)}
			}
		case evdev.EV_ABS:
			switch v.Code {
			case evdev.ABS_X:
				dev.state.penCoord.X = dev.params.penXInterval.normalize(float32(v.Value))
				dev.state.penInputEventFlags.set(inputEventFlagPosition)
			case evdev.ABS_Y:
				dev.state.penCoord.Y = dev.params.penYInterval.normalize(float32(v.Value))
				dev.state.penInputEventFlags.set(inputEventFlagPosition)
			case evdev.ABS_DISTANCE:
				dev.state.penDistance = dev.params.penDistanceInterval.normalize(float32(v.Value))
				dev.state.penInputEventFlags.set(inputEventFlagPosition)
			case evdev.ABS_PRESSURE:
				dev.state.penPressure = dev.params.penPressureInterval.normalize(float32(v.Value))
				// Pressure is emitted as a synthesized button event.
				dev.state.penInputEventFlags.set(inputEventFlagPressure)
			}

		case evdev.EV_KEY:
			switch v.Code {
			case evdev.BTN_TOOL_PEN:
				dev.state.penToolSelected = v.Value == 1
				if v.Value == 1 {
					dev.state.penTool = ButtonPenTip
				}
			case evdev.BTN_TOOL_RUBBER:
				dev.state.penToolSelected = v.Value == 1
				if v.Value == 1 {
					dev.state.penTool = ButtonPenEraser
				}
			case evdev.BTN_TOUCH:
				// The touch event is not needed since it can be dervied from the
				// pressure event.
			default:
				if button, ok := buttonCodeTrans[v.Code]; ok {
					s := &dev.state.penButtonEvents
					*s = append(*s, &EventButton{
						Timestamp: inputEventTime(&v),
						Button:    button,
						Pressure:  normalizeDigitalButtonValue(v.Value),
					})
				}
			}
		}
	}
	return
}

func (dev *wacomDevice) inputEventFinger(inputEvents []evdev.InputEvent) (events []Event) {
	// Don't care about multi touch events for now.
	// Just generate position events from the absolute X and Y positions and
	// button events for finger touching and leaving the pad.

	for _, v := range inputEvents {
		switch v.Type {
		case evdev.EV_SYN:
			switch v.Code {
			case evdev.SYN_REPORT:
				if dev.state.fingerInputEventFlags.has(inputEventFlagPosition) {
					events = append(events, &EventPositionFinger{
						Timestamp: inputEventTime(&v),
						Coord:     dev.state.fingerCoord,
					})
				}
				if dev.state.fingerInputEventFlags.has(inputEventFlagButton) {
					events = append(events, &EventButton{
						Timestamp: inputEventTime(&v),
						Button:    ButtonTouch,
						Pressure:  dev.state.fingerTouchPressure,
					})
				}
				dev.state.fingerInputEventFlags = 0
			case evdev.SYN_DROPPED:
				// TODO(jb): See comment in inputEventPad about this condition.
				return []Event{newEventError(eventBufferOverrunError)}
			}
		case evdev.EV_ABS:
			switch v.Code {
			case evdev.ABS_X:
				dev.state.fingerCoord.X = dev.params.fingerXInterval.normalize(float32(v.Value))
				dev.state.fingerInputEventFlags.set(inputEventFlagPosition)
			case evdev.ABS_Y:
				dev.state.fingerCoord.Y = dev.params.fingerYInterval.normalize(float32(v.Value))
				dev.state.fingerInputEventFlags.set(inputEventFlagPosition)
			}

		case evdev.EV_KEY:
			// The tool is always "finger" so we don't have to check it.
			if v.Code == evdev.BTN_TOUCH {
				dev.state.fingerTouchPressure = normalizeDigitalButtonValue(v.Value)
				dev.state.fingerInputEventFlags.set(inputEventFlagButton)
			}
		}
	}
	return
}

func (dev *wacomDevice) inputEventPad(inputEvents []evdev.InputEvent) (events []Event) {
	// The pad only generates button events so don't bother synching with SYN_REPORT.
	for _, v := range inputEvents {
		switch v.Type {
		case evdev.EV_SYN:
			if v.Code == evdev.SYN_DROPPED {
				// Buffer overrun in the evdev client's event queue.
				// Client should ignore all events up to and including next
				// SYN_REPORT event and query the device.
				//
				// TODO(jb): Querying device requires keeping state around to be able to
				//           synthesize events. Since we don't do that just raise an error and
				//           see how often this happens in practice. There is code in eventMux
				//           to drop some events so hopefully we can keep up.
				return []Event{newEventError(eventBufferOverrunError)}
			}
		case evdev.EV_KEY:
			if button, ok := buttonCodeTrans[v.Code]; ok {
				events = append(events, &EventButton{
					Timestamp: inputEventTime(&v),
					Button:    button,
					Pressure:  normalizeDigitalButtonValue(v.Value),
				})
			}
		}
	}
	return
}

type wacomLinuxDeviceType int

const (
	wacomLinuxDeviceTypePen wacomLinuxDeviceType = iota
	wacomLinuxDeviceTypeFinger
	wacomLinuxDeviceTypePad
	wacomLinuxDeviceTypes
)

func wacomLinuxDeviceTypeFromString(s string) wacomLinuxDeviceType {
	switch s {
	case "Pen":
		return wacomLinuxDeviceTypePen
	case "Finger":
		return wacomLinuxDeviceTypeFinger
	case "Pad":
		return wacomLinuxDeviceTypePad
	}
	return wacomLinuxDeviceTypePen
}
