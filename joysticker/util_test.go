package joysticker

import (
	"net"
	"testing"
	"time"
)

func TestParseBluetoothSockaddr(t *testing.T) {
	// 0xDC 0xA6 0x32 0xC4 0xDC 0x93
	// bytes := []byte{0xDC, 0xA6, 0x32, 0xC4, 0xDC, 0x93}
	addr := "DC:A6:32:C4:DC:93"
	hwAddr, _ := net.ParseMAC(addr)

	t.Log(hwAddr)
	t.Log([]byte(hwAddr))
}

func TestSliceReplace(t *testing.T) {
	slice := []int{1, 1, 1, 1, 1, 1}
	new := []int{2, 2, 2}
	copy(slice[2:4], new)
	t.Log(slice)
}

func TestUpdateTimer(t *testing.T) {
	lastTime := time.Now()
	time.Sleep(4 * time.Second)
	now := time.Now()
	duration := now.Sub(lastTime)
	elapsed := int64(duration*4) & 0xFF
	t.Log(elapsed)
}
