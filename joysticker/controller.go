package joysticker

import (
	_ "embed"

	"dio.wtf/joysticker/joysticker/log"
)

//go:embed sdp/controller.xml
var sdpRecord []byte

const (
	GAMEPAD_CLASS = "0x002508"
	SDP_UUID      = "00001000-0000-1000-8000-00805f9b34fb"
	HID_PATH      = "/joysticker/controller"

	ALIAS = "Pro Controller"
)

type Controller struct {
	*Device
}

func (c *Controller) Setup() (err error) {
	if err = c.SetPowered(true); nil != err {
		log.Error(err)
	}
	if err = c.SetPairable(true); nil != err {
		log.Error(err)
	}
	if err = c.SetPairableTimeout(0); nil != err {
		log.Error(err)
	}
	if err = c.SetDiscoverableTimeout(180); nil != err {
		log.Error(err)
	}
	if err = c.SetAlias(ALIAS); nil != err {
		log.Error(err)
	} else {
		log.Debug("setting device name to Pro Controller...")
	}

	options := map[string]interface{}{
		"ServiceRecord":         string(sdpRecord),
		"Role":                  "server",
		"RequireAuthentication": false,
		"RequireAuthorization":  false,
		"AutoConnect":           true,
	}
	err = c.RegisterProfile(HID_PATH, SDP_UUID, options)
	return
}
