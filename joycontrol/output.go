package joycontrol

import (
	"errors"
	"fmt"
	"strings"
)

type OutputReportId uint8

const (
	RumbleAndSubcommand OutputReportId = 0x01
	UpdateNFCPacket     OutputReportId = 0x03
	RumbleOnly          OutputReportId = 0x10
	RequestNFCData      OutputReportId = 0x11
	// UnknownOutputType   OutputReportId = 0x12

	OutputReportHeader byte = 0xA2
	OutputReportLength int  = 50
)

var (
	errBadLengthData     = errors.New("receive bad length data")
	errMalformedData     = errors.New("receive malformed data")
	errUnknownOutputId   = errors.New("receive unknown output report id")
	errUnknownSubcommand = errors.New("receive unknown subcommand")
)

// OutputReport represents report sent from the Switch to the Controller.
type OutputReport []byte

func (o OutputReport) validate() error {
	if len(o) != OutputReportLength {
		return errBadLengthData
	}
	if o[0] != OutputReportHeader {
		return errMalformedData
	}
	id := o.getId()
	if id != RumbleAndSubcommand &&
		id != RumbleOnly &&
		id != RequestNFCData &&
		id != UpdateNFCPacket {
		return errUnknownOutputId

	}
	if id == RumbleAndSubcommand {
		subcommand := o.getSubcommand()
		if subcommand != RequestDeviceInfo &&
			subcommand != SetInputReportMode &&
			subcommand != TriggerButtonsElapsedTime &&
			subcommand != SetShipmentLowPowerState &&
			subcommand != SpiFlashRead &&
			subcommand != SetNfcMcuConfig &&
			subcommand != SetNfcMcuState &&
			subcommand != SetPlayerLights &&
			subcommand != EnableImu &&
			subcommand != EnableVibration {
			return errUnknownSubcommand
		}
	}

	return nil
}

func (o OutputReport) getId() OutputReportId {
	return OutputReportId(o[1])
}

func (o OutputReport) getSubcommand() Subcommand {
	b := o[11]
	return Subcommand(b)
}

func (o OutputReport) getSubcommandData() []byte {
	return o[12:]
}

func (o OutputReport) String() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("--- %s Msg ---", o.getSubcommand().String()))
	builder.WriteString("\nPayload:    ")
	for _, p := range o[:11] {
		builder.WriteString(fmt.Sprintf("0x%02X ", p))
	}
	builder.WriteString("\nSubcommand: ")
	for _, p := range o[11:] {
		builder.WriteString(fmt.Sprintf("0x%02X ", p))
	}
	return builder.String()
}
