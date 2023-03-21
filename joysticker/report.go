package joysticker

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/exp/slices"
	"golang.org/x/sys/unix"
)

// https://github.com/dekuNukem/Nintendo_Switch_Reverse_Engineering/blob/master/bluetooth_hid_notes.md

type InputReportId uint8

const (
	ButtonAction      InputReportId = 0x3F
	SubcommandReplies InputReportId = 0x21
	UpdateNFCReport   InputReportId = 0x23
	StandardFullMode  InputReportId = 0x30
	NFCMode           InputReportId = 0x31
	// UnknownInputType  InputReportId = 0x32 | 0x33

	InputReportHeader byte = 0xA1
	// InputReportLength int  = 363 // header + 362 Standard input report
	InputReportLength int = 50
)

var emptyInputReport = [InputReportLength]byte{0xA1}

// InputReport represents report sent from the Controller to the Switch.
type InputReport struct {
	data [InputReportLength]byte
}

func (i *InputReport) reset() {
	copy(i.data[:], emptyInputReport[:])
}

func (i *InputReport) setReportId(id InputReportId) {
	i.data[1] = byte(id)
}

func (i *InputReport) fillStandardData(elapsed int64, queryDeviceIno bool) {
	i.data[2] = byte(elapsed)

	if queryDeviceIno {
		i.data[3] = 0x90 + 0x00 // Battery level + Connection info

		i.data[4] = 0x00 // Button state
		i.data[5] = 0x00
		i.data[6] = 0x00

		i.data[7] = 0x00 // Left Stick state
		i.data[8] = 0x00
		i.data[9] = 0x00

		i.data[10] = 0x00 // Right Stick state
		i.data[11] = 0x00
		i.data[12] = 0x00

		i.data[13] = 0x80 // Vibrator
	}
}

func (i *InputReport) ackSetInputReportMode() {
	i.data[14] = 0x80                     // ACK without data
	i.data[15] = byte(SetInputReportMode) // Subcommand Reply
}

func (i *InputReport) ackDeviceInfo(mac []byte) {
	i.data[14] = 0x82                    // ACK with data
	i.data[15] = byte(RequestDeviceInfo) // Subcommand Reply

	i.data[16] = 0x03 // Firmware version
	i.data[17] = 0x8B

	i.data[18] = 0x03 // Pro Controller

	i.data[19] = 0x02 // Unknown Byte, always 2

	copy(i.data[20:26], mac)

	i.data[26] = 0x01 // Unknown byte, always 1
	i.data[27] = 0x01 // Controller colours location
}

func (i *InputReport) ackTriggerButtonsElapsedTime() {
	i.data[14] = 0x83                            // ACK
	i.data[15] = byte(TriggerButtonsElapsedTime) // Subcommand Reply
}

func (i *InputReport) ackSetShipmentLowPowerState() {
	i.data[14] = 0x80                           // ACK
	i.data[15] = byte(SetShipmentLowPowerState) // Subcommand Reply
}

// https://github.com/dekuNukem/Nintendo_Switch_Reverse_Engineering/blob/master/spi_flash_notes.md#x6000-factory-configuration-and-calibration
func (i *InputReport) ackSpiFlashRead(data []byte) {
	lowEnd := data[0]
	highEnd := data[1]
	sectionRange := data[4]

	i.data[14] = 0x90               // ACK
	i.data[15] = byte(SpiFlashRead) // Subcommand Reply

	i.data[16] = lowEnd       // Low byte in Little-Endian address
	i.data[17] = highEnd      // High byte in Little-Endian address
	i.data[20] = sectionRange // Section range

	if lowEnd == 0x00 && highEnd == 0x60 {
		// Serial number
		replaceSlice(i.data[:], 21, 21+int(sectionRange), 0xFF)
	} else if lowEnd == 0x50 && highEnd == 0x60 {
		// Body #RGB color
		replaceSlice(i.data[:], 21, 21+int(sectionRange), 0xFF)
	} else if lowEnd == 0x80 && highEnd == 0x60 {
		// Factory Sensor and Stick device parameters
		replaceSlice(i.data[:], 21, 21+int(sectionRange), 0xFF)
	} else if lowEnd == 0x98 && highEnd == 0x60 {
		// Factory Stick device parameters 2
		replaceSlice(i.data[:], 21, 21+int(sectionRange), 0xFF)
	} else if lowEnd == 0x10 && highEnd == 0x80 {
		// User Analog sticks calibration
		replaceSlice(i.data[:], 21, 21+int(sectionRange), 0xFF)
	} else if lowEnd == 0x3D && highEnd == 0x60 {
		// Factory configuration & calibration 2
		leftCalibration := []byte{
			0xBA, 0xF5, 0x62,
			0x6F, 0xC8, 0x77,
			0xED, 0x95, 0x5B}
		rightCalibration := []byte{
			0x16, 0xD8, 0x7D,
			0xF2, 0xB5, 0x5F,
			0x86, 0x65, 0x5E}
		copy(i.data[21:30], leftCalibration)
		copy(i.data[30:39], rightCalibration)
		replaceSlice(i.data[:], 39, 39+int(sectionRange)-len(leftCalibration)-len(rightCalibration), 0xFF)
	} else if lowEnd == 0x20 && highEnd == 0x60 {
		// Factory configuration & calibration 1
		replaceSlice(i.data[:], 21, 21+int(sectionRange), 0xFF)
	}
}

func (i *InputReport) ackSetNfcMcuConfig() {
	i.data[14] = 0xA0                  // ACK
	i.data[15] = byte(SetNfcMcuConfig) // Subcommand Reply

	data := []byte{
		0x01, 0x00, 0xFF, 0x00, 0x08, 0x00,
		0x1B, 0x01, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0xC8}
	copy(i.data[16:16+len(data)], data)
}

func (i *InputReport) ackSetNfcMcuState() {
	i.data[14] = 0x80                 // ACK
	i.data[15] = byte(SetNfcMcuState) // Subcommand Reply
}

func (i *InputReport) ackSetPlayerLights() {
	i.data[14] = 0x80                  // ACK
	i.data[15] = byte(SetPlayerLights) // Subcommand Reply
}

func (i *InputReport) ackEnableImu() {
	// TODO: Toggle IMU
	i.data[14] = 0x80            // ACK
	i.data[15] = byte(EnableImu) // Subcommand Reply
}

func (i *InputReport) ackEnableVibration() {
	i.data[14] = 0x82                  // ACK
	i.data[15] = byte(EnableVibration) // Subcommand Reply
}

func (i *InputReport) String() string {
	var builder strings.Builder
	builder.WriteString("\nPayload:    ")
	for _, p := range i.data[:14] {
		builder.WriteString(fmt.Sprintf("0x%02X ", p))
	}
	builder.WriteString("\nSubcommand: ")
	for _, p := range i.data[14:] {
		builder.WriteString(fmt.Sprintf("0x%02X ", p))
	}
	return builder.String()
}

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
	errEmptyData         = errors.New("receive empty data")
	errBadLengthData     = errors.New("receive bad length data")
	errMalformedData     = errors.New("receive malformed data")
	errUnknownSubcommand = errors.New("receive unknown subcommand")
)

type Subcommand uint8

// https://github.com/dekuNukem/Nintendo_Switch_Reverse_Engineering/blob/master/bluetooth_hid_subcommands_notes.md
const (
	RequestDeviceInfo         Subcommand = 0x02
	SetInputReportMode        Subcommand = 0x03
	TriggerButtonsElapsedTime Subcommand = 0x04
	SetShipmentLowPowerState  Subcommand = 0x08
	SpiFlashRead              Subcommand = 0x10
	SetNfcMcuConfig           Subcommand = 0x21
	SetNfcMcuState            Subcommand = 0x22
	SetPlayerLights           Subcommand = 0x30
	EnableImu                 Subcommand = 0x40
	EnableVibration           Subcommand = 0x48
)

// OutputReport represents report sent from the Switch to the Controller.
type OutputReport struct {
	id   OutputReportId
	data [OutputReportLength]byte
}

func (o *OutputReport) load(fd int) error {
	n, err := unix.Read(fd, o.data[:])
	if err != nil {
		return err
	}
	if n != OutputReportLength {
		return errBadLengthData
	}
	if o.data[0] != OutputReportHeader {
		return errMalformedData
	}

	typeSlice := []OutputReportId{RumbleAndSubcommand, UpdateNFCPacket, RumbleOnly, RequestNFCData}
	id := OutputReportId(o.data[1])
	if !slices.Contains(typeSlice, id) {
		return errUnknownSubcommand
	}
	o.id = id
	return nil
}

func (o *OutputReport) getSubcommand() Subcommand {
	b := o.data[11]
	return Subcommand(b)
}

func (o *OutputReport) getSubcommandData() []byte {
	return o.data[12:]
}

func (o *OutputReport) String() string {
	var builder strings.Builder
	builder.WriteString("\nPayload:    ")
	for _, p := range o.data[:11] {
		builder.WriteString(fmt.Sprintf("0x%02X ", p))
	}
	builder.WriteString("\nSubcommand: ")
	for _, p := range o.data[11:] {
		builder.WriteString(fmt.Sprintf("0x%02X ", p))
	}
	return builder.String()
}
