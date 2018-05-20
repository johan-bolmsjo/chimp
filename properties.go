package chimp

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

// Properties of device.
type Properties map[Property]PropertyValue

func (props Properties) String() (result string) {
	var sb strings.Builder
	tw := tabwriter.NewWriter(&sb, 4, 0, 1, ' ', 0)

	twOk := func(err error) bool {
		if err != nil {
			result = fmt.Sprintf("tabwriter-error: %q", err)
		}
		return err == nil
	}

	twPrintf := func(format string, a ...interface{}) bool {
		_, err := fmt.Fprintf(tw, format, a...)
		return twOk(err)
	}

	if !twPrintf("Properties: {\n") {
		return
	}

	kvList := props.sortedKvList()
	for _, kv := range kvList {
		if !twPrintf("    %s:\t%s\n", kv.k, kv.v) {
			return
		}
	}

	if !twPrintf("}") || !twOk(tw.Flush()) {
		return
	}

	return sb.String()
}

type kv struct {
	k, v string
}

// Get sorted key/value from map holding properties.
func (props Properties) sortedKvList() []kv {
	var s []kv
	for k, v := range props {
		s = append(s, kv{k: string(k), v: v.String()})
	}
	sort.Slice(s, func(i, j int) bool { return s[i].k < s[j].k })
	return s
}

// Property of device
type Property string

const (
	PropertyDeviceName           Property = "device-name"
	PropertyDeviceType           Property = "device-type"
	PropertyPadWidthMillimeters  Property = "pad-width-millimeters"  // Width of the pad along the X-axis.
	PropertyPadHeightMillimeters Property = "pad-height-millimeters" // Width of the pad along the Y-axis.
	PropertyPadWidthHeightRatio  Property = "pad-width-height-ratio"
)

// PropertyValue represents property values of different types.
type PropertyValue interface {
	// Type returns the property value type as a string, e.g. "string" or "number".
	Type() string

	// String returns the property value as a string.
	// All types can be converted to a string.
	String() string

	// Number returns the property value as a float.
	// Zero is returned when this is not applicable.
	Number() float64
}

// PropertyValueString represent string property values.
type PropertyValueString string

func (val PropertyValueString) Type() string {
	return "string"
}

func (val PropertyValueString) String() string {
	return string(val)
}

func (val PropertyValueString) Number() float64 {
	return 0
}

// PropertyValueNumber represent numeric property values.
type PropertyValueNumber float64

func (val PropertyValueNumber) Type() string {
	return "number"
}

func (val PropertyValueNumber) String() string {
	return strconv.FormatFloat(float64(val), 'g', -1, 64)
}

func (val PropertyValueNumber) Number() float64 {
	return float64(val)
}
