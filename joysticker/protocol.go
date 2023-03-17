package joysticker

import "errors"

type Subcommand uint8

// https://github.com/dekuNukem/Nintendo_Switch_Reverse_Engineering/blob/master/bluetooth_hid_subcommands_notes.md
const (
	RequestDeviceInfo         Subcommand = 0x02
	SetInputReportMode                   = 0x03
	TriggerButtonsElapsedTime            = 0x04
	SetShipmentLowPowerState             = 0x08
	SPIFlashRead                         = 0x10
	SetNFCMCUConfiguration               = 0x21
	SetNFCMCUState                       = 0x22
	SetPlayerLights                      = 0x30
	EnableIMU                            = 0x40
	EnableVibration                      = 0x48
)

type Protocol struct {
}

var (
	errEmptyData         = errors.New("receive empty data")
	errBadLengthData     = errors.New("receive bad length data")
	errMalformedData     = errors.New("receive malformed data")
	errUnknownSubcommand = errors.New("receive unknown subcommand")
)

type SwitchResponse struct {
	cmd  Subcommand
	data []byte
}

func NewSwitchResponse(msg []byte) (*SwitchResponse, error) {
	if len(msg) < responseDataLength {
		return nil, errBadLengthData
	}
	if msg[0] != 0xA2 {
		return nil, errMalformedData
	}

	cmd := Subcommand(msg[11])
	switch cmd {
	case RequestDeviceInfo, SetInputReportMode, TriggerButtonsElapsedTime,
		SetShipmentLowPowerState, SPIFlashRead, SetNFCMCUConfiguration,
		SetNFCMCUState, SetPlayerLights, EnableIMU, EnableVibration:
		return &SwitchResponse{
			cmd:  cmd,
			data: msg[11:],
		}, nil
	default:
		return nil, errUnknownSubcommand
	}
}
