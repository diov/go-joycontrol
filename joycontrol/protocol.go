package joycontrol

import (
	"time"

	R "dio.wtf/joycontrol/joycontrol/report"
)

type Protocol struct {
	lastTime           time.Time
	elapsed            int64
	deviceInfoRequired bool
	imuEnabled         bool
	vibrationEnabled   bool
	playerNumber       bool

	macAddr []byte
}

func NewProtocol() *Protocol {
	return &Protocol{}
}

func (p *Protocol) Setup(macAddr []byte) {
	p.macAddr = macAddr
}

func (p Protocol) generateStandardReport() (input *R.InputReport) {
	p.updateTimer()

	input = AllocStandardReport()
	input.SetReportId(R.StandardFullMode)
	input.FillStandardData(p.elapsed, p.deviceInfoRequired)
	input.SetImuData(p.imuEnabled)
	return
}

func (p *Protocol) processSubcommandReport(output R.OutputReport) (input *R.InputReport) {
	p.updateTimer()

	subcommand := output.Subcommand()
	switch subcommand {
	case R.RequestDeviceInfo:
		input = p.answerDeviceInfo()
	case R.SetInputReportMode:
		input = p.answerSetMode(output.SubcommandData())
	case R.TriggerButtonsElapsedTime:
		input = p.anwserTriggerButtonsElapsedTime()
	case R.SetShipmentLowPowerState:
		input = p.answerSetShipmentState()
	case R.SpiFlashRead:
		input = p.answerSpiRead(output.SubcommandData())
	case R.SetNfcMcuConfig:
		input = p.answerSetNfcMcuConfig(output.SubcommandData())
	case R.SetNfcMcuState:
		input = p.answerSetNfcMcuState(output.SubcommandData())
	case R.SetPlayerLights:
		input = p.answerSetPlayerLights()
	case R.EnableImu:
		input = p.answerEnableImu(output.SubcommandData())
	case R.EnableVibration:
		input = p.answerEnableVibration()
	default:
		// Currently set so that the controller ignores any unknown
		// subcommands. This is better than sending a NACK response
		// since we'd just get stuck in an infinite loop arguing
		// with the Switch.
		input = p.generateStandardReport()
	}
	return
}

func (p *Protocol) answerSetMode(data []byte) (input *R.InputReport) {
	// TODO: Update input input mode
	input = AllocStandardReport()
	input.SetReportId(R.SubcommandReplies)
	input.FillStandardData(p.elapsed, p.deviceInfoRequired)
	input.AckSetInputReportMode()
	return
}

func (p *Protocol) anwserTriggerButtonsElapsedTime() (input *R.InputReport) {
	input = AllocStandardReport()
	input.SetReportId(R.SubcommandReplies)
	input.FillStandardData(p.elapsed, p.deviceInfoRequired)
	input.AckTriggerButtonsElapsedTime()
	return
}

func (p *Protocol) answerDeviceInfo() (input *R.InputReport) {
	p.deviceInfoRequired = true

	input = AllocStandardReport()
	input.SetReportId(R.SubcommandReplies)
	input.FillStandardData(p.elapsed, p.deviceInfoRequired)
	input.AckDeviceInfo(p.macAddr)
	return
}

func (p *Protocol) answerSetShipmentState() (input *R.InputReport) {
	input = AllocStandardReport()
	input.SetReportId(R.SubcommandReplies)
	input.FillStandardData(p.elapsed, p.deviceInfoRequired)
	input.AckSetShipmentLowPowerState()
	return
}

func (p *Protocol) answerSpiRead(data []byte) (input *R.InputReport) {
	input = AllocStandardReport()
	input.SetReportId(R.SubcommandReplies)
	input.FillStandardData(p.elapsed, p.deviceInfoRequired)
	input.AckSpiFlashRead(data)
	return
}

func (p *Protocol) answerSetNfcMcuConfig(data []byte) (input *R.InputReport) {
	// TODO: Update NFC MCU config
	input = AllocStandardReport()
	input.SetReportId(R.SubcommandReplies)
	input.FillStandardData(p.elapsed, p.deviceInfoRequired)
	input.AckSetNfcMcuConfig()
	return
}

func (p *Protocol) answerSetNfcMcuState(data []byte) (input *R.InputReport) {
	// TODO: Update NFC MCU State
	input = AllocStandardReport()
	input.SetReportId(R.SubcommandReplies)
	input.FillStandardData(p.elapsed, p.deviceInfoRequired)
	input.AckSetNfcMcuState()
	return
}

func (p *Protocol) answerSetPlayerLights() (input *R.InputReport) {
	p.playerNumber = true
	input = AllocStandardReport()
	input.SetReportId(R.SubcommandReplies)
	input.FillStandardData(p.elapsed, p.deviceInfoRequired)
	input.AckSetPlayerLights()
	return
}

func (p *Protocol) answerEnableImu(data []byte) (input *R.InputReport) {
	if data[0] == 0x01 {
		p.imuEnabled = true
	}

	input = AllocStandardReport()
	input.SetReportId(R.SubcommandReplies)
	input.FillStandardData(p.elapsed, p.deviceInfoRequired)
	input.AckEnableImu()
	return
}

func (p *Protocol) answerEnableVibration() (input *R.InputReport) {
	p.vibrationEnabled = true
	input = AllocStandardReport()
	input.SetReportId(R.SubcommandReplies)
	input.FillStandardData(p.elapsed, p.deviceInfoRequired)
	input.AckEnableVibration()
	return
}

func (p *Protocol) updateTimer() {
	duration := time.Since(p.lastTime)

	p.elapsed = (p.elapsed + (duration.Microseconds() * 4)) & 0xFF
	p.lastTime = time.Now()
}
