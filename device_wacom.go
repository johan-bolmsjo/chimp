package chimp

const (
	wacomBamboo16FG6x8PadWidthMillimeters  = 216.0
	wacomBamboo16FG6x8PadHeightMillimeters = 137.0
	wacomBamboo16FG6x8PadWidthHeightRatio  = wacomBamboo16FG6x8PadWidthMillimeters / wacomBamboo16FG6x8PadHeightMillimeters
	wacomBamboo16FG6x8PenXMax              = 21648
	wacomBamboo16FG6x8PenYMax              = 13700
	wacomBamboo16FG6x8PenPressureMax       = 1023
	wacomBamboo16FG6x8PenDistanceMax       = 30
	wacomBamboo16FG6x8FingerXMax           = 4095
	wacomBamboo16FG6x8FingerYMax           = 4095
)

var wacomBamboo16FG6x8Properties = Properties{
	PropertyDeviceName:           PropertyValueString("Wacom Bamboo 16FG 6x8"),
	PropertyDeviceType:           PropertyValueString(DeviceTypeTablet.String()),
	PropertyPadWidthMillimeters:  PropertyValueNumber(wacomBamboo16FG6x8PadWidthMillimeters),
	PropertyPadHeightMillimeters: PropertyValueNumber(wacomBamboo16FG6x8PadHeightMillimeters),
	PropertyPadWidthHeightRatio:  PropertyValueNumber(wacomBamboo16FG6x8PadWidthHeightRatio),
}

var wacomBamboo16FG6x8Capabilities = Capabilities{
	PositionDevices: []PositionDevice{PositionDevicePen, PositionDeviceFinger},
	Buttons: []Button{
		ButtonPenTip,
		ButtonPenEraser,
		ButtonPen1,
		ButtonPen2,
		ButtonLeft,
		ButtonRight,
		ButtonForward,
		ButtonBack,
		ButtonTouch,
	},
}

var wacomBamboo16FG6x8DeviceParams = wacomDeviceParams{
	penXInterval:        f32cival{b: wacomBamboo16FG6x8PenXMax},
	penYInterval:        f32cival{b: wacomBamboo16FG6x8PenYMax},
	penPressureInterval: f32cival{b: wacomBamboo16FG6x8PenPressureMax},
	penDistanceInterval: f32cival{b: wacomBamboo16FG6x8PenDistanceMax},
	fingerXInterval:     f32cival{b: wacomBamboo16FG6x8FingerXMax},
	fingerYInterval:     f32cival{b: wacomBamboo16FG6x8FingerYMax},
}

// Like properties but internal.
type wacomDeviceParams struct {
	penXInterval        f32cival
	penYInterval        f32cival
	penPressureInterval f32cival
	penDistanceInterval f32cival
	fingerXInterval     f32cival
	fingerYInterval     f32cival
}
