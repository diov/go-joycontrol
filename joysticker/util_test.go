package joysticker

import (
	"fmt"
	"net"
	"strings"
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

func TestHexString(t *testing.T) {
	data := []byte{1, 0, 255, 0, 8, 0, 27, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 200}

	var hex []string
	for _, d := range data {
		hex = append(hex, fmt.Sprintf("0x%02X ", d))
	}
	t.Log(strings.Join(hex, ","))
}

func TestUpdateTimer(t *testing.T) {
	lastTime := time.Now()
	time.Sleep(4 * time.Second)
	now := time.Now()
	duration := now.Sub(lastTime)
	elapsed := int64(duration*4) & 0xFF
	t.Log(elapsed)
}
