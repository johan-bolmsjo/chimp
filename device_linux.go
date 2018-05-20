package chimp

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"time"

	"github.com/johan-bolmsjo/chimp/internal/channel"
	"github.com/johan-bolmsjo/golang-evdev"
)

func listDevices() ([]DeviceInfo, error) {
	devices, err := evdev.ListInputDevices()
	if err != nil {
		return nil, err
	}

	deviceMatchers := []deviceMatcher{
		newDeviceMatcherWacomBamboo16FG6x8(),
	}

	logicalDevices := map[string]logicalDevice{}

	for _, dev := range devices {
		// Close opened file to avoid leaking file descriptors. The
		// parameters will be matched against saved parameters when a
		// selected device is opened later on to make sure that device
		// names has not been renumbered when replugging devices.
		dev.File.Close()

		devInfo := linuxDeviceInfo{
			dev:  dev.Fn,
			name: dev.Name,
			phys: dev.Phys,
		}
		for _, matcher := range deviceMatchers {
			if match, logicalID := matcher.match(devInfo); match {
				logicalDevice := logicalDevices[logicalID]
				if logicalDevice == nil {
					logicalDevice = matcher.newLogicalDevice()
					logicalDevices[logicalID] = logicalDevice
				}
				logicalDevice.addLinuxDevice(devInfo)
			}
		}
	}

	type tmpLogicalDevice struct {
		logicalID     string
		logicalDevice logicalDevice
	}

	// Logical devices sorted by logical ID. Map iteration order is not
	// deterministic so sort them to ensure the same list order given the
	// same set of devices.
	var sortedLogicalDevices []tmpLogicalDevice
	for k, v := range logicalDevices {
		sortedLogicalDevices = append(sortedLogicalDevices, tmpLogicalDevice{k, v})
	}
	s := sortedLogicalDevices
	sort.Slice(s, func(i, j int) bool { return s[i].logicalID < s[j].logicalID })

	var deviceInfo []DeviceInfo
	for _, v := range sortedLogicalDevices {
		deviceInfo = append(deviceInfo, v.logicalDevice.deviceInfo())
	}

	return deviceInfo, nil
}

var (
	reUSBDeviceID = regexp.MustCompile(`^(usb-[0-9a-z:.-]+)/input\d+$`)
)

type deviceMatcher interface {
	// match checks if the Linux device information is known to the device
	// matcher and if so a unique logical device ID is generated from device
	// name and physical name. The logical device ID is used to group
	// separate physical devices into one logical device.
	match(devInfo linuxDeviceInfo) (match bool, logicalID string)

	// newLogicalDevice creates a new logical device to which physical devices can be added.
	newLogicalDevice() logicalDevice
}

type logicalDevice interface {
	// addLinuxDevice adds linux device information to the logical device.
	// It's used group several physical devices as one logical device.
	addLinuxDevice(devInfo linuxDeviceInfo)

	// deviceInfo creates device info though which the logical device can be
	// identified and opened.
	deviceInfo() DeviceInfo
}

type linuxDeviceInfo struct {
	dev, name, phys string
}

// openDevice opens an input device and verifies that the device name and
// physical location remains unchanged from when the information was collected.
func (info *linuxDeviceInfo) openDevice() (dev *evdev.InputDevice, err error) {
	if dev, err = evdev.Open(info.dev); err == nil {
		if dev.Name != info.name || dev.Phys != info.phys {
			dev.File.Close()
			return nil, fmt.Errorf("opened input device {%s, %s} does match saved parameters {%s, %s}", dev.Name, dev.Phys, info.dev, info.phys)
		}
		if err = dev.Grab(); err != nil {
			dev.File.Close()
		}
	}
	return
}

type eventMux struct {
	events       <-chan Event
	sync         *channel.ConsSync
	inputDevices []*evdev.InputDevice
	prod         eventMuxProd
}

type eventMuxProd struct {
	events chan<- Event
	sync   channel.ProdSync
}

func newEventMux() eventMux {
	events := make(chan Event, 100)
	sync := channel.NewConsSync()
	return eventMux{
		events: events,
		sync:   sync,
		prod: eventMuxProd{
			events: events,
			sync:   sync.ProdSync(),
		},
	}
}

// inputEventFunc processes Linux input events and produce events exposed by this package.
type inputEventFunc func(inputEvents []evdev.InputEvent) []Event

// addEventSource adds input device to mux and starts a goroutine to read and
// process events from it. The input event function should process the Linux input
// events and emit events of type Event to the mux using send() or trySend().
func (mux *eventMux) addEventSource(inputDevice *evdev.InputDevice, inputEventFunc inputEventFunc) {
	mux.inputDevices = append(mux.inputDevices, inputDevice)
	muxProd := &mux.prod

	mux.sync.Add(1)
	go func() {
	out:
		for {
			inputEvents, err := inputDevice.Read()
			if err != nil {
				muxProd.send(newEventError(err))
				break out
			}

			events := inputEventFunc(inputEvents)
			for _, event := range events {
				shutdown := false
				switch v := event.(type) {
				case *EventPositionPen, *EventPositionFinger:
					// Can always drop position events without harm
					shutdown = muxProd.trySend(event)
				case *EventButton:
					if v.Pressure == 0 {
						// Always emit button release events
						shutdown = muxProd.send(event)
					} else {
						shutdown = muxProd.trySend(event)
					}
				case *eventError:
					muxProd.send(event)
					shutdown = true
				default:
					shutdown = muxProd.send(event)
				}

				if shutdown {
					break out
				}
			}
		}
		muxProd.sync.Done()
	}()
}

var errorEventSourceClosed = errors.New("input event source closed")

// Read consumes an event.
func (mux *eventMux) Read() (Event, error) {
	event, ok := <-mux.events
	if !ok {
		return nil, errorEventSourceClosed
	}

	if event, ok := event.(*eventError); ok {
		err := event.err
		if !mux.close() {
			// This was not the first reported error or caused by a call to
			// *eventMux.Close(). Convert spurious errors caused by shutting down
			// producers to a generic error message.
			err = errorEventSourceClosed
		}
		return nil, err
	}
	return event, nil
}

func (mux *eventMux) close() (shutdown bool) {
	if wait := mux.sync.Shutdown(); wait != nil {
		// Close all input sources so that producers wake up if stuck on read.
		for _, v := range mux.inputDevices {
			v.File.Close()
		}

		wait()

		// All producers are gone so it's safe to close the event channel to wake up
		// any consumer stuck on reading from it. Multiple consumers are not
		// advisable as events can be processed out of order but make it work.
		close(mux.prod.events)
		return true
	}
	return false
}

// Close the event source.
func (mux *eventMux) Close() {
	mux.close()
}

// Send event to channel, returns true if shutdown is in progress.
func (muxProd *eventMuxProd) send(event Event) (shutdown bool) {
	select {
	case muxProd.events <- event:
		return false
	case <-muxProd.sync.SignalChan:
		return true
	}
}

// Try to send event but drop it if channel queue is full, returns true if shutdown is in progress.
func (muxProd *eventMuxProd) trySend(event Event) (shutdown bool) {
	select {
	case muxProd.events <- event:
		return false
	case <-muxProd.sync.SignalChan:
		return true
	default:
		return false
	}
}

// Translates from Linux button codes to package exported button codes.
var buttonCodeTrans = map[uint16]Button{
	evdev.BTN_STYLUS:  ButtonPen1,
	evdev.BTN_STYLUS2: ButtonPen2,
	evdev.BTN_LEFT:    ButtonLeft,
	evdev.BTN_RIGHT:   ButtonRight,
	evdev.BTN_FORWARD: ButtonForward,
	evdev.BTN_BACK:    ButtonBack,
}

var digitalButtonInterval = f32cival{b: 1}

func normalizeDigitalButtonValue(v int32) float32 {
	return digitalButtonInterval.normalize(float32(v))
}

// Convert struct timeval like time in input event to Go time type.
func inputEventTime(event *evdev.InputEvent) time.Time {
	return time.Unix(event.Time.Sec, event.Time.Usec*1000)
}

type inputEventFlag uint8

const (
	inputEventFlagPosition inputEventFlag = 1 << iota
	inputEventFlagButton
	inputEventFlagPressure
)

// Add one or more flags to set.
func (set *inputEventFlag) set(flag inputEventFlag) {
	*set |= flag
}

// Check if one or more flags are in set.
func (set *inputEventFlag) has(flag inputEventFlag) bool {
	return *set&flag == flag
}
