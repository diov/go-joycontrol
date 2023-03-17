package joysticker

import (
	"testing"

	"github.com/godbus/dbus/v5"
)

func TestDbusProperties(t *testing.T) {
	bus, _ := dbus.SystemBus()

	obj := bus.Object("org.bluez", "/org/bluez/hci0")
	v, err := obj.GetProperty("org.bluez.Adapter1.Pairable")
	if nil != err {
		t.Error(err)
		return
	}
	t.Log(v)
}
