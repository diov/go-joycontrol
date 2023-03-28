package report

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

func replaceSlice(slice []byte, start, end int, replacement byte) {
	for i := start; i < end; i++ {
		slice[i] = replacement
	}
}
