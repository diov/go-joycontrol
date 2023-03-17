package joysticker

import (
	"dio.wtf/joysticker/joysticker/log"
	"github.com/godbus/dbus/v5"
	"golang.org/x/exp/slices"
	"golang.org/x/sys/unix"
)

const responseDataLength = 50

type Server struct {
	device     *Device
	controller *Controller
	needWatch  bool
}

func NewServer() *Server {
	device, _ := NewDevice()
	controller := &Controller{Device: device}
	return &Server{
		device:     device,
		controller: controller,
	}
}

func (s *Server) Run() {
	toggleCleanBluez(true)

	s.controller.Setup()
	s.Connect()
}

func (s *Server) Connect() {
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
	s.device.SetClass("0x002508")

	s.needWatch = true
	go s.watchConnReset()

	_, itrAddr, err := unix.Accept(itrSock)
	log.DebugF("Accept interupt from %v %v", itrAddr, err)
	_, ctrlAddr, err := unix.Accept(ctrlSock)
	log.DebugF("Accept control from %v %v", ctrlAddr, err)
	s.needWatch = false
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
			s.device.SetClass("0x02508")
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
					log.DebugF("A Nintendo Switch disconnected. Resetting Connection %s...", k)
					if err := s.device.RemoveDevice(dbus.ObjectPath(k)); nil != err {
						log.DebugF("Remove device failed: %v", err)
					}
					disconnectRecord[k] = 0
				}
			}
		}
	}
}
