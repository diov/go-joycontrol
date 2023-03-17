package joysticker

import (
	"testing"
)

func TestParseBluetoothSockaddr(t *testing.T) {
	// CC:BB:AA:33:22:11
	// [6]byte{0x11, 0x22, 0x33, 0xaa, 0xbb, 0xcc}
	// addr := "CC:BB:AA:33:22:11"
	// bleAddr := ParseBluetoothSockaddr(addr, 16)

	// if bleAddr.Addr != [6]byte{0x11, 0x22, 0x33, 0xaa, 0xbb, 0xcc} {
	// 	t.Errorf("Invalidate address")
	// }
}
