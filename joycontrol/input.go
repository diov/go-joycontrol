package joycontrol

import (
	"fmt"
	"strings"
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

// InputReport represents report sent from the Controller to the Switch.
type InputReport []byte

func (i InputReport) setReportId(id InputReportId) {
	i[1] = byte(id)
}

func (i InputReport) setImuData(enabled bool) {
	if !enabled {
		return
	}

	data := []byte{
		0x75, 0xFD, 0xFD, 0xFF, 0x09, 0x10, 0x21, 0x00, 0xD5, 0xFF,
		0xE0, 0xFF, 0x72, 0xFD, 0xF9, 0xFF, 0x0A, 0x10, 0x22, 0x00,
		0xD5, 0xFF, 0xE0, 0xFF, 0x76, 0xFD, 0xFC, 0xFF, 0x09, 0x10,
		0x23, 0x00, 0xD5, 0xFF, 0xE0, 0xFF}
	copy(i[14:14+len(data)], data)
}

func (i InputReport) fillStandardData(elapsed int64, queryDeviceIno bool) {
	i[2] = byte(elapsed)

	if queryDeviceIno {
		i[3] = 0x90 + 0x00 // Battery level + Connection info(Pro Controller)

		i[4] = 0x00 // Button state
		i[5] = 0x00
		i[6] = 0x00

		i[7] = 0x6F // Left Stick state
		i[8] = 0xC8
		i[9] = 0x77

		i[10] = 0x16 // Right Stick state
		i[11] = 0xD8
		i[12] = 0x7D

		i[13] = 0x80 // Vibrator
	}
}

func (i InputReport) setButtonState(data []byte) {
	copy(i[4:6], data)
}

func (i InputReport) ackSetInputReportMode() {
	i[14] = 0x80                     // ACK without data
	i[15] = byte(SetInputReportMode) // Subcommand Reply
}

func (i InputReport) ackDeviceInfo(mac []byte) {
	i[14] = 0x82                    // ACK with data
	i[15] = byte(RequestDeviceInfo) // Subcommand Reply

	i[16] = 0x03 // Firmware version
	i[17] = 0x8B

	i[18] = 0x03 // Pro Controller

	i[19] = 0x02 // Unknown Byte, always 2

	copy(i[20:26], mac)

	i[26] = 0x01 // Unknown byte, always 1
	i[27] = 0x01 // Controller colours location
}

func (i InputReport) ackTriggerButtonsElapsedTime() {
	i[14] = 0x83                            // ACK
	i[15] = byte(TriggerButtonsElapsedTime) // Subcommand Reply
}

func (i InputReport) ackSetShipmentLowPowerState() {
	i[14] = 0x80                           // ACK
	i[15] = byte(SetShipmentLowPowerState) // Subcommand Reply
}

// https://github.com/dekuNukem/Nintendo_Switch_Reverse_Engineering/blob/master/spi_flash_notes.md
func (i InputReport) ackSpiFlashRead(data []byte) {
	lowEnd := data[0]
	highEnd := data[1]
	sectionRange := data[4]

	i[14] = 0x90               // ACK
	i[15] = byte(SpiFlashRead) // Subcommand Reply

	i[16] = lowEnd       // Low byte in Little-Endian address
	i[17] = highEnd      // High byte in Little-Endian address
	i[20] = sectionRange // Section range

	if lowEnd == 0x00 && highEnd == 0x60 {
		// Serial number
		replaceSlice(i[:], 21, 21+int(sectionRange), 0xFF)
	} else if lowEnd == 0x50 && highEnd == 0x60 {
		// Body #RGB color
		replaceSlice(i[:], 21, 21+int(sectionRange), 0xFF)
	} else if lowEnd == 0x80 && highEnd == 0x60 {
		// Factory Sensor and Stick device parameters
		// TODO: Copy NXBT
		replaceSlice(i[:], 21, 21+int(sectionRange), 0xFF)
	} else if lowEnd == 0x98 && highEnd == 0x60 {
		// Factory Stick device parameters 2
		// TODO: Copy NXBT
		replaceSlice(i[:], 21, 21+int(sectionRange), 0xFF)
	} else if lowEnd == 0x10 && highEnd == 0x80 {
		// User Analog sticks calibration
		replaceSlice(i[:], 21, 21+int(sectionRange), 0xFF)
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
		copy(i[21:30], leftCalibration)
		copy(i[30:39], rightCalibration)
		replaceSlice(i[:], 39, 39+int(sectionRange)-len(leftCalibration)-len(rightCalibration), 0xFF)
	} else if lowEnd == 0x20 && highEnd == 0x60 {
		// Factory configuration & calibration 1
		replaceSlice(i[:], 21, 21+int(sectionRange), 0xFF)
	}
}

func (i InputReport) ackSetNfcMcuConfig() {
	i[14] = 0xA0                  // ACK
	i[15] = byte(SetNfcMcuConfig) // Subcommand Reply

	data := []byte{
		0x01, 0x00, 0xFF, 0x00, 0x08, 0x00,
		0x1B, 0x01, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0xC8}
	copy(i[16:16+len(data)], data)
}

func (i InputReport) ackSetNfcMcuState() {
	i[14] = 0x80                 // ACK
	i[15] = byte(SetNfcMcuState) // Subcommand Reply
}

func (i InputReport) ackSetPlayerLights() {
	i[14] = 0x80                  // ACK
	i[15] = byte(SetPlayerLights) // Subcommand Reply
}

func (i InputReport) ackEnableImu() {
	// TODO: Toggle IMU
	i[14] = 0x80            // ACK
	i[15] = byte(EnableImu) // Subcommand Reply
}

func (i InputReport) ackEnableVibration() {
	i[14] = 0x82                  // ACK
	i[15] = byte(EnableVibration) // Subcommand Reply
}

func (i InputReport) String() string {
	var builder strings.Builder

	id := InputReportId(i[1])
	if id == SubcommandReplies {
		builder.WriteString(fmt.Sprintf("--- %s Msg ---", Subcommand(i[15]).String()))
	}
	builder.WriteString("\nPayload:    ")
	for _, p := range i[:14] {
		builder.WriteString(fmt.Sprintf("0x%02X ", p))
	}
	builder.WriteString("\nSubcommand: ")
	for _, p := range i[14:] {
		builder.WriteString(fmt.Sprintf("0x%02X ", p))
	}
	return builder.String()
}

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

func (s Subcommand) String() string {
	switch s {
	case 0x02:
		return "RequestDeviceInfo"
	case 0x03:
		return "SetInputReportMode"
	case 0x04:
		return "TriggerButtonsElapsedTime"
	case 0x08:
		return "SetShipmentLowPowerState"
	case 0x10:
		return "SpiFlashRead"
	case 0x21:
		return "SetNfcMcuConfig"
	case 0x22:
		return "SetNfcMcuState"
	case 0x30:
		return "SetPlayerLights"
	case 0x40:
		return "EnableImu"
	case 0x48:
		return "EnableVibration"
	default:
		return "UNKNOWN"
	}
}
