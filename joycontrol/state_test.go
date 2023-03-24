package joycontrol

import "testing"

func TestFlipBit(t *testing.T) {
	var b byte = 0b00000000
	bit := 1
	t.Logf("binary: %08b, decimal: %d\n", b, b)
	// Press
	if (b>>bit)&1 != 1 {
		b ^= 1 << bit
		t.Logf("binary: %08b, decimal: %d\n", b, b)
	}
	// Release
	if (b>>bit)&1 == 1 {
		b ^= 1 << bit
		t.Logf("binary: %08b, decimal: %d\n", b, b)
	}
}
