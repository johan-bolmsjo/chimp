package chimp

import (
	"fmt"
	"time"
)

// Event is implemented by all event types.
type Event interface {
	// Time is the time when the event was generated.
	Time() time.Time

	// String returns a textual representation of the event.
	String() string
}

// EventPositionPen is generated for movement of pen on a typical 2D tablet.
type EventPositionPen struct {
	Timestamp time.Time // Time when event was generated.
	Coord     Coord2D   // Pen position on tablet, axis are in range [0, 1], origo in upper left corner.
	Distance  float32   // Distance of for example pen to tablet in range [0, 1], 0 is on tablet.
}

func (e *EventPositionPen) Time() time.Time {
	return e.Timestamp
}

func (e *EventPositionPen) String() string {
	return fmt.Sprintf(fmtEventPositionPen, e.Timestamp, e.Coord.X, e.Coord.Y, e.Distance)
}

// EventPositionFinger is generated for movement of finger on tablet or similar.
type EventPositionFinger struct {
	Timestamp time.Time // Time when event was generated.
	Coord     Coord2D   // Finger position, axis are in range [0, 1], origo in upper left corner
}

func (e *EventPositionFinger) Time() time.Time {
	return e.Timestamp
}

func (e *EventPositionFinger) String() string {
	return fmt.Sprintf(fmtEventPositionFinger, e.Timestamp, e.Coord.X, e.Coord.Y)
}

// PositionDevice is an enumeration of different position device types.
type PositionDevice uint32

//go:generate stringer -type=PositionDevice -trimprefix=PositionDevice

const (
	PositionDevicePen PositionDevice = iota
	PositionDeviceFinger
)

// EventButton is generated for everything that can be modeled as a digital or
// analogue button.
type EventButton struct {
	Timestamp time.Time // Time when event was generated.
	Button    Button    // Button code.
	Pressure  float32   // Pressure on button in range [0, 1], 0 or 1 for digital buttons.
}

func (e *EventButton) Time() time.Time {
	return e.Timestamp
}

func (e *EventButton) String() string {
	return fmt.Sprintf(fmtEventButton, e.Timestamp, e.Button, e.Pressure)
}

// Button is an enumeration of different buttons.
type Button uint32

//go:generate stringer -type=Button -trimprefix=Button

const (
	ButtonPenTip    Button = iota // Tip of pen
	ButtonPenEraser               // Eraser on pen
	ButtonPen1                    // Button 1 on pen
	ButtonPen2                    // Button 2 on pen
	ButtonPen3                    // Button 3 on pen
	ButtonLeft
	ButtonRight
	ButtonForward
	ButtonBack
	ButtonTouch // Single-touch event such as finger on touchpad
)

const fmtEventPositionPen = `EventPositionPen: {
    Time:     %s
    X:        %f
    Y:        %f
    Distance: %f
}`

const fmtEventPositionFinger = `EventPositionFinger: {
    Time:     %s
    X:        %f
    Y:        %f
}`

const fmtEventButton = `EventButton: {
    Time:     %s
    Name:     %s
    Pressure: %f
}`

// Internal event indicating input error.
// Event is never exposed to user of API but unboxed to proper error.
type eventError struct {
	timestamp time.Time
	err       error
}

func newEventError(err error) *eventError {
	return &eventError{
		timestamp: time.Now(),
		err:       err,
	}
}

func (e *eventError) Time() time.Time {
	return e.timestamp
}

func (e *eventError) String() string {
	return e.err.Error()
}
