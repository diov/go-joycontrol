package joysticker

import (
	"strings"

	"dio.wtf/joysticker/joysticker/log"
	"github.com/godbus/dbus/v5"
	"github.com/muka/go-bluetooth/bluez"
	"github.com/muka/go-bluetooth/bluez/profile/adapter"
	"github.com/muka/go-bluetooth/bluez/profile/device"
	"github.com/muka/go-bluetooth/bluez/profile/profile"
	"github.com/muka/go-bluetooth/hw/linux/cmd"
)

type Device struct {
	*adapter.Adapter1
	devicePath string
	deviceId   string
}

func NewDevice() (d *Device, err error) {
	objects, err := getManagedObjects()
	if nil != err {
		return
	}

	var adapter1 *adapter.Adapter1
	var objectPath string
	for path, ifaces := range objects {
		if _, ok := ifaces[adapter.Adapter1Interface]; ok {
			dev, err := adapter.NewAdapter1(path)
			if nil != err {
				return nil, err
			}
			adapter1 = dev
			objectPath = string(path)
			break
		}
	}

	s := strings.Split(string(objectPath), "/")
	deviceId := s[len(s)-1]
	log.DebugF("Using adapter under object path: %s", objectPath)
	return &Device{
		Adapter1:   adapter1,
		devicePath: objectPath,
		deviceId:   deviceId,
	}, nil
}

func (d *Device) PrepairedSwitches() (paths []dbus.ObjectPath, err error) {
	objects, err := getManagedObjects()
	if nil != err {
		return
	}

	for path, ifaces := range objects {
		if iface, ok := ifaces[device.Device1Interface]; ok {
			prop := new(device.Device1Properties)
			prop, err = prop.FromDBusMap(iface)
			if nil != err {
				return
			}
			if prop.Name == "Nintendo Switch" {
				paths = append(paths, path)
			}
		}
	}
	return
}

func (d *Device) SetClass(cls string) error {
	_, err := cmd.Exec("hciconfig", d.deviceId, "class", cls)
	return err
}

func (d *Device) Reset() error {
	_, err := cmd.Exec("hciconfig", d.deviceId, "reset")
	return err
}

func (d *Device) RegisterProfile(profilePath, uuid string, options map[string]interface{}) error {
	mgr, err := profile.NewProfileManager1()
	if nil != err {
		return err
	}

	return mgr.RegisterProfile(dbus.ObjectPath(profilePath), uuid, options)
}

func (d *Device) FindConnectedAdapter() (paths []string, err error) {
	objects, err := getManagedObjects()
	if nil != err {
		return
	}

	for path, ifaces := range objects {
		if iface, ok := ifaces[device.Device1Interface]; ok {
			prop := new(device.Device1Properties)
			prop, err = prop.FromDBusMap(iface)
			if nil != err {
				return
			}
			if prop.Connected && (prop.Name == "Nintendo Switch" || prop.Alias == "Nintendo Switch") {
				paths = append(paths, string(path))
			}
		}
	}
	return
}

func getManagedObjects() (map[dbus.ObjectPath]map[string]map[string]dbus.Variant, error) {
	om, err := bluez.GetObjectManager()
	if nil != err {
		return nil, err
	}
	objects, err := om.GetManagedObjects()
	if nil != err {
		return nil, err
	}
	return objects, nil
}
