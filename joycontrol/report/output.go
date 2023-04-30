package report

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/exp/slices"
)

const (
	OutputReportHeader byte = 0xA2
	OutputReportLength int  = 50
)

var outputReportIds = []OutputReportId{RumbleAndSubcommand, UpdateNfcPacket, RumbleOnly, RequestNfcData, UnknownOutputType}

var (
	ErrBadLengthData   = errors.New("receive bad length data")
	ErrMalformedData   = errors.New("receive malformed data")
	ErrUnknownOutputId = errors.New("receive unknown output report id")
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
	if !slices.Contains(outputReportIds, id) {
		return ErrUnknownOutputId
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

func (o OutputReport) SubcommandArgs() []byte {
	return o[12:]
}

func (o OutputReport) McuCommand() McuCommand {
	// OUTPUT 0x01(RumbleAndSubcommand) && Subcommand 0x21(SetNfcMcuConfig)
	// OUTPUT 0x11(RequestNfcData)
	if o.Id() == RumbleAndSubcommand && o.Subcommand() == SetNfcMcuConfig {
		return McuCommand(o[12])
	}
	return McuCommand(o[11])
}

func (o OutputReport) McuCommandArgs() []byte {
	if o.Id() == RumbleAndSubcommand && o.Subcommand() == SetNfcMcuConfig {
		return o[13:]
	}
	return o[12:]
}

func (o OutputReport) String() string {
	var builder strings.Builder
	if o.Id() == RumbleAndSubcommand {
		if o.Subcommand() == SetNfcMcuConfig {
			builder.WriteString(fmt.Sprintf("--- %s(%s) Msg ---", o.Subcommand().String(), o.McuCommand().String()))
		} else {
			builder.WriteString(fmt.Sprintf("--- %s Msg ---", o.Subcommand().String()))
		}
	} else {
		builder.WriteString(fmt.Sprintf("--- %s Msg ---", o.McuCommand().String()))
	}
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
