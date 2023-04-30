package report

type OutputReportId uint8

const (
	RumbleAndSubcommand OutputReportId = 0x01
	UpdateNfcPacket     OutputReportId = 0x03
	RumbleOnly          OutputReportId = 0x10
	RequestNfcData      OutputReportId = 0x11
	UnknownOutputType   OutputReportId = 0x12
)

func (o OutputReportId) String() string {
	switch o {
	case 0x01:
		return "RumbleAndSubcommand"
	case 0x03:
		return "UpdateNfcPacket"
	case 0x10:
		return "RumbleOnly"
	case 0x11:
		return "RequestNfcData"
	case 0x12:
		return "UnknownOutputType"
	default:
		return "UNKNOWN"
	}
}

type InputReportId uint8

const (
	SimpleHidId        InputReportId = 0x3F
	SubcommandReplies  InputReportId = 0x21
	StandardFullModeId InputReportId = 0x30
	NfcMcuModeId       InputReportId = 0x31
	// UpdateNfcReport    InputReportId = 0x23
	// UnknownInputType   InputReportId = 0x32 | 0x33
)

type InputReportMode uint8

const (
	StandFullMode InputReportMode = 0x30
	NfcMode       InputReportMode = 0x31
	SimpleHidMode InputReportMode = 0x3F
)

// https://github.com/dekuNukem/Nintendo_Switch_Reverse_Engineering/blob/master/bluetooth_hid_subcommands_notes.md
type Subcommand uint8

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

type McuCommand uint8

const (
	SetMcuMode           McuCommand = 0x21
	RequestMcuStatus     McuCommand = 0x01
	RequestNfcDataReport McuCommand = 0x02
	RequestIrDataReport  McuCommand = 0x03
)

func (m McuCommand) String() string {
	switch m {
	case 0x21:
		return "SetMcuMode"
	case 0x01:
		return "RequestMcuStatus"
	case 0x02:
		return "RequestNfcDataReport"
	case 0x03:
		return "RequestIrDataReport"
	default:
		return "UNKNOWN"
	}
}

func replaceSlice(slice []byte, start, end int, replacement byte) {
	for i := start; i < end; i++ {
		slice[i] = replacement
	}
}
