package joycontrol

import (
	_ "embed"
	"net"

	"dio.wtf/joycontrol/joycontrol/log"
	"github.com/godbus/dbus/v5"
	"github.com/google/uuid"
	"golang.org/x/exp/slices"
	"golang.org/x/sys/unix"
)

//go:embed sdp/controller.xml
var sdpRecord string

const (
	GAMEPAD_CLASS = "0x002508"
	HID_PATH      = "/joysticker/controller"

	ALIAS = "Pro Controller"
)

type Server struct {
	device    *Device
	protocol  *Protocol
	needWatch bool
}

func NewServer() *Server {
	device, _ := NewDevice()
	protocol := NewProtocol()
	return &Server{
		device:   device,
		protocol: protocol,
	}
}

func (s *Server) Run() {
	toggleCleanBluez(true)
	addr, _ := s.device.GetAddress()
	mac, _ := net.ParseMAC(addr)

	s.Setup()
	itr, ctrl := s.Connect()
	s.protocol.Setup(itr, ctrl, mac)
}

func (s *Server) Setup() (err error) {
	if err = s.device.SetPowered(true); nil != err {
		log.Error(err)
	}
	if err = s.device.SetPairable(true); nil != err {
		log.Error(err)
	}
	if err = s.device.SetPairableTimeout(0); nil != err {
		log.Error(err)
	}
	if err = s.device.SetDiscoverableTimeout(180); nil != err {
		log.Error(err)
	}
	if err = s.device.SetAlias(ALIAS); nil != err {
		log.Error(err)
	} else {
		log.Debug("setting device name to Pro Controller...")
	}

	options := map[string]interface{}{
		"ServiceRecord":         sdpRecord,
		"Role":                  "server",
		"RequireAuthentication": false,
		"RequireAuthorization":  false,
		"AutoConnect":           true,
	}
	sdpUuid := uuid.NewString()
	err = s.device.RegisterProfile(HID_PATH, sdpUuid, options)
	return
}

func (s *Server) Connect() (int, int) {
	addr, _ := s.device.GetAddress()
	log.DebugF("MAC: %s", addr)

	ctrlSock, err := SetupSocket(addr, 17)
	if nil != err {
		log.Error(err)
	}
	itrSock, err := SetupSocket(addr, 19)
	if nil != err {
		log.Error(err)
	}
	s.device.SetDiscoverable(true)
	s.device.SetClass(GAMEPAD_CLASS)

	s.needWatch = true
	go s.watchConnReset()

	itr, itrAddr, _ := unix.Accept(itrSock)
	itrL2Addr := itrAddr.(*unix.SockaddrL2)
	log.DebugF("Accept interrupt %d from %v %d", itr, itrAddr, itrL2Addr.PSM)
	ctrl, ctrlAddr, _ := unix.Accept(ctrlSock)
	ctrlL2Addr := ctrlAddr.(*unix.SockaddrL2)
	log.DebugF("Accept control %d from %v %d", ctrl, ctrlAddr, ctrlL2Addr.PSM)
	s.needWatch = false

	// stop advertising
	s.device.SetDiscoverable(false)
	s.device.SetPairable(false)

	return itr, ctrl
}

func (s *Server) watchConnReset() {
	connectedDevice := make(map[string]struct{})
	disconnectRecord := make(map[string]int)
	for s.needWatch {
		discoverable, _ := s.device.GetDiscoverable()
		if !discoverable {
			log.Debug("Resetup device")
			s.device.SetPowered(true)
			s.device.SetPairable(true)
			s.device.SetPairableTimeout(0)
			s.device.SetDiscoverable(true)
			s.device.SetClass(GAMEPAD_CLASS)
		}
		paths, _ := s.device.FindConnectedAdapter()

		for i := range paths {
			path := paths[i]
			connectedDevice[path] = struct{}{}
		}

		disconnected := make([]string, 0)
		for k := range connectedDevice {
			if !slices.Contains(paths, k) {
				disconnected = append(disconnected, k)
			}
		}
		if len(disconnected) > 0 {
			for _, k := range disconnected {
				if _, ok := disconnectRecord[k]; ok {
					disconnectRecord[k] = disconnectRecord[k] + 1
				} else {
					disconnectRecord[k] = 1
				}
				delete(connectedDevice, k)
			}
		}

		// Delete Switches that connect/disconnect twice.
		// This behaviour is characteristic of connection issues and is corrected
		// by removing the Switch's connection to the system.
		if len(disconnectRecord) > 0 {
			for k, v := range disconnectRecord {
				if v >= 2 {
					log.DebugF("A Nintendo Switch disconnected. Resetting Connection...Removing %s", k)
					if err := s.device.RemoveDevice(dbus.ObjectPath(k)); nil != err {
						log.DebugF("Remove device failed: %v", err)
					}
					disconnectRecord[k] = 0
				}
			}
		}
	}
}
