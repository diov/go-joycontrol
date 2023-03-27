package joycontrol

// https://github.com/dekuNukem/Nintendo_Switch_Reverse_Engineering/blob/master/bluetooth_hid_notes.md

type ControllerState struct {
	dirty bool
	bs    *ButtonState
}

func NewControllerState() *ControllerState {
	return &ControllerState{
		bs: &ButtonState{
			data: [3]byte{},
		},
	}
}

func (c *ControllerState) press(buttons ...string) {
	c.dirty = true
	c.bs.press(buttons...)
}

func (c *ControllerState) release(buttons ...string) {
	c.dirty = true
	c.bs.release(buttons...)
}

func (c *ControllerState) dump() []byte {
	c.dirty = false
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
	"Down":         {2, 0},
	"Up":           {2, 1},
	"Right":        {2, 2},
	"Left":         {2, 3},
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
