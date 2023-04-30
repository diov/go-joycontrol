package joycontrol

import (
	"net"
	"time"

	C "dio.wtf/joycontrol/joycontrol/controller"
	"dio.wtf/joycontrol/joycontrol/log"
	R "dio.wtf/joycontrol/joycontrol/report"
)

type Protocol struct {
	lastTime time.Time
	elapsed  int64

	mac []byte
}

func NewProtocol(mac net.HardwareAddr) *Protocol {
	return &Protocol{
		mac: mac,
	}
}

func (p Protocol) generateStandardReport(ctrl *C.Controller) (input *R.InputReport) {
	p.updateTimer()

	input = AllocStandardReport()
	input.SetReportId(R.StandardFullModeId)
	input.FillStandardData(p.elapsed, ctrl.DeviceInfoRequired)
	input.SetImuData(ctrl.ImuEnabled)
	return
}

func (p *Protocol) processSubcommandReport(ctrl *C.Controller, output R.OutputReport) (input *R.InputReport) {
	p.updateTimer()

	subcommand := output.Subcommand()
	switch subcommand {
	case R.RequestDeviceInfo:
		input = p.answerDeviceInfo(ctrl)
	case R.SetInputReportMode:
		input = p.answerSetMode(ctrl, output)
	case R.TriggerButtonsElapsedTime:
		input = p.anwserTriggerButtonsElapsedTime(ctrl)
	case R.SetShipmentLowPowerState:
		input = p.answerSetShipmentState(ctrl)
	case R.SpiFlashRead:
		input = p.answerSpiRead(ctrl, output)
	case R.SetNfcMcuConfig:
		input = p.answerSetNfcMcuConfig(ctrl, output)
	case R.SetNfcMcuState:
		input = p.answerSetNfcMcuState(ctrl, output)
	case R.SetPlayerLights:
		input = p.answerSetPlayerLights(ctrl)
	case R.EnableImu:
		input = p.answerEnableImu(ctrl, output)
	case R.EnableVibration:
		input = p.answerEnableVibration(ctrl)
	default:
		// Currently set so that the controller ignores any unknown
		// subcommands. This is better than sending a NACK response
		// since we'd just get stuck in an infinite loop arguing
		// with the Switch.
		input = p.generateStandardReport(ctrl)
	}
	return
}

func (p *Protocol) answerSetMode(ctrl *C.Controller, output R.OutputReport) (input *R.InputReport) {
	ctrl.Mode = R.InputReportMode(output.SubcommandArgs()[0])

	input = AllocStandardReport()
	input.SetReportId(R.SubcommandReplies)
	input.FillStandardData(p.elapsed, ctrl.DeviceInfoRequired)
	input.AckSetInputReportMode()
	return
}

func (p *Protocol) anwserTriggerButtonsElapsedTime(ctrl *C.Controller) (input *R.InputReport) {
	input = AllocStandardReport()
	input.SetReportId(R.SubcommandReplies)
	input.FillStandardData(p.elapsed, ctrl.DeviceInfoRequired)
	input.AckTriggerButtonsElapsedTime()
	return
}

func (p *Protocol) answerDeviceInfo(ctrl *C.Controller) (input *R.InputReport) {
	ctrl.DeviceInfoRequired = true

	input = AllocStandardReport()
	input.SetReportId(R.SubcommandReplies)
	input.FillStandardData(p.elapsed, ctrl.DeviceInfoRequired)
	input.AckDeviceInfo(p.mac)
	return
}

func (p *Protocol) answerSetShipmentState(ctrl *C.Controller) (input *R.InputReport) {
	input = AllocStandardReport()
	input.SetReportId(R.SubcommandReplies)
	input.FillStandardData(p.elapsed, ctrl.DeviceInfoRequired)
	input.AckSetShipmentLowPowerState()
	return
}

func (p *Protocol) answerSpiRead(ctrl *C.Controller, output R.OutputReport) (input *R.InputReport) {
	args := output.SubcommandArgs()

	input = AllocStandardReport()
	input.SetReportId(R.SubcommandReplies)
	input.FillStandardData(p.elapsed, ctrl.DeviceInfoRequired)
	input.AckSpiFlashRead(args)
	return
}

func (p *Protocol) answerSetNfcMcuConfig(ctrl *C.Controller, output R.OutputReport) (input *R.InputReport) {
	state := ctrl.McuState()

	switch output.McuCommand() {
	case R.SetMcuMode:
		args := output.McuCommandArgs()
		subcmd := args[0]
		if subcmd == 0x00 {
			switch args[1] {
			case 0x00:
				ctrl.SetMcuState(C.McuStandby)
			case 0x04:
				ctrl.SetMcuState(C.McuNfc)
			default:
				log.DebugF("Unknown NFC MCU mode: %02x", args[1])
			}
		} else {
			log.DebugF("Unknown NFC MCU subcommand: %02x", subcmd)
		}
	default:
		log.DebugF("Unknown NFC MCU command: %02x", output.McuCommand())
	}

	input = AllocStandardReport()
	input.SetReportId(R.SubcommandReplies)
	input.FillStandardData(p.elapsed, ctrl.DeviceInfoRequired)
	input.AckSetNfcMcuConfig(state)
	input.UpdateChecksum(crc8Checksum((*input)[16:]))
	return
}

func (p *Protocol) answerSetNfcMcuState(ctrl *C.Controller, output R.OutputReport) (input *R.InputReport) {
	args := output.SubcommandArgs()
	ctrl.ToggleMcuPower(args[0] == 0x01)

	input = AllocStandardReport()
	input.SetReportId(R.SubcommandReplies)
	input.FillStandardData(p.elapsed, ctrl.DeviceInfoRequired)
	input.AckSetNfcMcuState()
	return
}

func (p *Protocol) answerSetPlayerLights(ctrl *C.Controller) (input *R.InputReport) {
	ctrl.PlayerNumber = true

	input = AllocStandardReport()
	input.SetReportId(R.SubcommandReplies)
	input.FillStandardData(p.elapsed, ctrl.DeviceInfoRequired)
	input.AckSetPlayerLights()
	return
}

func (p *Protocol) answerEnableImu(ctrl *C.Controller, output R.OutputReport) (input *R.InputReport) {
	args := output.SubcommandArgs()
	ctrl.ImuEnabled = args[0] == 0x01

	input = AllocStandardReport()
	input.SetReportId(R.SubcommandReplies)
	input.FillStandardData(p.elapsed, ctrl.DeviceInfoRequired)
	input.AckEnableImu()
	return
}

func (p *Protocol) answerEnableVibration(ctrl *C.Controller) (input *R.InputReport) {
	ctrl.VibrationEnabled = true

	input = AllocStandardReport()
	input.SetReportId(R.SubcommandReplies)
	input.FillStandardData(p.elapsed, ctrl.DeviceInfoRequired)
	input.AckEnableVibration()
	return
}

func (p *Protocol) processNfcDataReport(ctrl *C.Controller, output R.OutputReport) {
	// TODO: Handle Request NFC Data Report
}

func (p *Protocol) updateTimer() {
	duration := time.Since(p.lastTime)

	p.elapsed = (p.elapsed + (duration.Microseconds() * 4)) & 0xFF
	p.lastTime = time.Now()
}
