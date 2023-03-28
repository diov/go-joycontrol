package report

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
	ErrBadLengthData     = errors.New("receive bad length data")
	ErrMalformedData     = errors.New("receive malformed data")
	ErrUnknownOutputId   = errors.New("receive unknown output report id")
	ErrUnknownSubcommand = errors.New("receive unknown subcommand")
)

// OutputReport represents report sent from the Switch to the Controller.
type OutputReport []byte

func (o OutputReport) Validate() error {
	if len(o) != OutputReportLength {
		return ErrBadLengthData
	}
	if o[0] != OutputReportHeader {
		return ErrMalformedData
	}
	id := o.Id()
	if id != RumbleAndSubcommand &&
		id != RumbleOnly &&
		id != RequestNFCData &&
		id != UpdateNFCPacket {
		return ErrUnknownOutputId

	}
	if id == RumbleAndSubcommand {
		subcommand := o.Subcommand()
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
			return ErrUnknownSubcommand
		}
	}

	return nil
}

func (o OutputReport) Id() OutputReportId {
	return OutputReportId(o[1])
}

func (o OutputReport) Subcommand() Subcommand {
	b := o[11]
	return Subcommand(b)
}

func (o OutputReport) SubcommandData() []byte {
	return o[12:]
}

func (o OutputReport) String() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("--- %s Msg ---", o.Subcommand().String()))
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
