package controller

import (
	R "dio.wtf/joycontrol/joycontrol/report"
)

// https://github.com/dekuNukem/Nintendo_Switch_Reverse_Engineering/blob/master/bluetooth_hid_notes.md

type Controller struct {
	Mode R.InputReportMode

	DeviceInfoRequired bool
	ImuEnabled         bool
	VibrationEnabled   bool
	PlayerNumber       bool

	Dirty bool
	bs    *ButtonState
	mcu   *MicroControllerUnit
}

func NewController() *Controller {
	return &Controller{
		bs: &ButtonState{
			data: [3]byte{},
		},
		mcu: &MicroControllerUnit{
			mode:       McuStandby,
			powerState: McuSuspend,
		},
	}
}

func (c *Controller) Press(buttons ...string) {
	c.Dirty = true
	c.bs.press(buttons...)
}

func (c *Controller) Release(buttons ...string) {
	c.Dirty = true
	c.bs.release(buttons...)
}

func (c *Controller) SetMcuState(state McuMode) {
	c.mcu.SetState(state)
}

func (c *Controller) ToggleMcuPower(on bool) {
	c.mcu.TogglePowerState(on)
}

func (c *Controller) McuState() []byte {
	return c.mcu.StateData()
}

func (c *Controller) Dump() []byte {
	c.Dirty = false
	return c.bs.data[:]
}

// | Byte       | x01 | x02 | x04    | x08    | x10 | x20    | x40 | x80         |
// |:----------:|:---:|:---:|:------:|:------:|:---:|:------:|:---:|:-----------:|
// | 3 (Right)  | Y   | X   | B      | A      | SR  | SL     | R   | ZR          |
// | 4 (Shared) | -   | +   | R Stick| L Stick| Home| Capture| --  |Charging Grip|
// | 5 (Left)   | Down| Up  | Right  | Left   | SR  | SL     | L   | ZL          |
var buttonMap = map[string]struct {
	index int
	bit   int
}{
	"Y": {0, 0},
	"X": {0, 1},
	"B": {0, 2},
	"A": {0, 3},
	// "SR": {0, 4},
	// "SL": {0, 5},
	"R":  {0, 6},
	"ZR": {0, 7},
	"+":  {1, 0},
	"-":  {1, 1},
	// "RStick":       {1, 2},
	// "LStick":       {1, 3},
	"Home":         {1, 4},
	"Capture":      {1, 5},
	"ChargingGrip": {1, 7},
	"DOWN":         {2, 0},
	"UP":           {2, 1},
	"RIGHT":        {2, 2},
	"LEFT":         {2, 3},
	// "SR":    {2, 4},
	// "SL":    {2, 5},
	"L":  {2, 6},
	"ZL": {2, 7},
}

type ButtonState struct {
	data [3]byte
}

func (b *ButtonState) press(buttons ...string) {
	for i := range buttons {
		button := buttons[i]
		if info, ok := buttonMap[button]; ok {
			// Check if button is already pressed
			if (b.data[info.index]>>info.bit)&1 != 1 {
				b.data[info.index] ^= 1 << info.bit
			}
		}
	}
}

func (b *ButtonState) release(buttons ...string) {
	for i := range buttons {
		button := buttons[i]
		if info, ok := buttonMap[button]; ok {
			// Check if button is pressed
			if (b.data[info.index]>>info.bit)&1 == 1 {
				b.data[info.index] ^= 1 << info.bit
			}
		}
	}
}
