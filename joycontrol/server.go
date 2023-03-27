package joycontrol

import (
	_ "embed"
	"errors"
	"net"
	"sync"
	"syscall"
	"time"

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
	device   *Device
	protocol *Protocol
	state    *ControllerState

	needWatch      bool
	reportReceived bool
	stateUpdated   bool
	mux            sync.RWMutex

	freqChan chan time.Duration

	output OutputReport
}

func NewServer() *Server {
	device, _ := NewDevice()
	protocol := NewProtocol()
	state := NewControllerState()
	return &Server{
		device:   device,
		protocol: protocol,
		state:    state,
		output:   make([]byte, OutputReportLength),
	}
}

func (s *Server) Start() {
	toggleCleanBluez(true)
	addr, _ := s.device.GetAddress()
	mac, _ := net.ParseMAC(addr)

	if err := s.Setup(); nil != err {
		log.Error(err)
		return
	}
	s.protocol.Setup(mac)
	itr, ctrl := s.Connect()
	go s.Run(itr, ctrl)

	// TODO: remove this test case
	go func() {
		time.Sleep(time.Second * 5)
		s.state.press("B")
		time.Sleep(time.Second / 10)
		s.state.release("B")
		time.Sleep(time.Second * 3)
		s.state.press("B")
		time.Sleep(time.Second / 10)
		s.state.release("B")
		time.Sleep(time.Second * 5)
		s.state.press("A")
		time.Sleep(time.Second / 10)
		s.state.release("A")
		time.Sleep(time.Second * 3)
		s.state.press("A")
		time.Sleep(time.Second / 10)
		s.state.release("A")
	}()
}

func (s *Server) Setup() (err error) {
	if err = s.device.SetPowered(true); nil != err {
		return
	}
	if err = s.device.SetPairable(true); nil != err {
		return
	}
	if err = s.device.SetPairableTimeout(0); nil != err {
		return
	}
	if err = s.device.SetDiscoverableTimeout(180); nil != err {
		return
	}
	if err = s.device.SetAlias(ALIAS); nil != err {
		return
	}
	log.Debug("setting device name to Pro Controller...")

	options := map[string]interface{}{
		"ServiceRecord":         sdpRecord,
		"Role":                  "server",
		"RequireAuthentication": false,
		"RequireAuthorization":  false,
		"AutoConnect":           true,
	}
	return s.device.RegisterProfile(HID_PATH, uuid.NewString(), options)
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
	log.DebugF("Accept interrupt %d from %v", itr, itrAddr)
	ctrl, ctrlAddr, _ := unix.Accept(ctrlSock)
	log.DebugF("Accept control %d from %v", ctrl, ctrlAddr)
	s.needWatch = false

	// stop advertising
	s.device.SetDiscoverable(false)
	s.device.SetPairable(false)

	if err := unix.SetNonblock(itr, true); nil != err {
		log.Error(err)
	}

	// Send an empty input report to the Switch to prompt a reply
	input := s.protocol.generateStandardReport()
	s.unixWrite(itr, input)

	timer := time.NewTimer(time.Second * 1)
	for range timer.C {
		// Switch responds to packets slower during pairing
		// Pairing cycle responds optimally on a 15Hz loop
		if !s.reportReceived {
			timer.Reset(time.Second * 1)
		} else {
			timer.Reset(time.Second / 15)
		}

		var err error
		var input *InputReport
		if _, err = s.unixRead(itr, s.output); err == nil {
			err = s.output.validate()
		}
		if err != nil {
			switch {
			case errors.Is(err, syscall.EAGAIN),
				errors.Is(err, errBadLengthData),
				errors.Is(err, errMalformedData),
				errors.Is(err, errUnknownOutputId),
				errors.Is(err, errUnknownSubcommand):
				input = s.protocol.generateStandardReport()
			default:
				log.ErrorF("error reading output report: %v", err)
				continue
			}
		} else {
			s.reportReceived = true
			switch s.output.getId() {
			case RumbleAndSubcommand:
				input = s.protocol.processSubcommandReport(s.output)
			default:
				log.Debug(s.output.getId())
				input = s.protocol.generateStandardReport()
			}
		}
		s.unixWrite(itr, input)

		if s.protocol.vibrationEnabled && s.protocol.playerNumber {
			log.Debug("Switch connected")
			break
		}
	}

	return itr, ctrl
}

func (s *Server) Run(itr, ctrl int) {
	tick := 0
	freq := time.Second / 66
	timer := time.NewTimer(freq)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			tick++
			timer.Reset(freq)

			var err error
			var input *InputReport
			if _, err = s.unixRead(itr, s.output); err == nil {
				err = s.output.validate()
			}
			if err != nil {
				input = s.protocol.generateStandardReport()
			} else {
				switch s.output.getId() {
				case RumbleAndSubcommand:
					input = s.protocol.processSubcommandReport(s.output)
					s.stateUpdated = true
				case RumbleOnly, UpdateNFCPacket, RequestNFCData:
					input = s.protocol.generateStandardReport()
				}
			}
			if s.state.dirty {
				input.setButtonState(s.state.dump())
				s.stateUpdated = true
			}
			if s.stateUpdated {
				_, err := s.unixWrite(itr, input)
				log.DebugF("MainLoop Update %s %v", input, err)
				s.stateUpdated = false
			} else if tick >= 132 {
				_, err := s.unixWrite(itr, input)
				log.DebugF("MainLoop Blank %s %v", input, err)
				tick = 0
			}
		case duration := <-s.freqChan:
			freq = duration
		}
	}
}

func (s *Server) unixRead(fd int, output OutputReport) (int, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return unix.Read(fd, output)
}

func (s *Server) unixWrite(fd int, input *InputReport) (int, error) {
	s.mux.Lock()
	defer func() {
		s.mux.Unlock()
		FreeReport(input)
	}()
	return unix.Write(fd, *input)
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

func (s *Server) Stop() {
	log.Debug("Gracefully shutting down server")
	toggleCleanBluez(false)
}
