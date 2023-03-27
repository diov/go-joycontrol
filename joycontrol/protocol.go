package joycontrol

import (
	"time"
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

func (p Protocol) generateStandardReport() (input *InputReport) {
	p.updateTimer()

	input = AllocStandardReport()
	input.setReportId(StandardFullMode)
	input.fillStandardData(p.elapsed, p.deviceInfoRequired)
	input.setImuData(p.imuEnabled)
	return
}

func (p *Protocol) processSubcommandReport(output OutputReport) (input *InputReport) {
	p.updateTimer()

	subcommand := output.getSubcommand()
	switch subcommand {
	case RequestDeviceInfo:
		input = p.answerDeviceInfo()
	case SetInputReportMode:
		input = p.answerSetMode(output.getSubcommandData())
	case TriggerButtonsElapsedTime:
		input = p.anwserTriggerButtonsElapsedTime()
	case SetShipmentLowPowerState:
		input = p.answerSetShipmentState()
	case SpiFlashRead:
		input = p.answerSpiRead(output.getSubcommandData())
	case SetNfcMcuConfig:
		input = p.answerSetNfcMcuConfig(output.getSubcommandData())
	case SetNfcMcuState:
		input = p.answerSetNfcMcuState(output.getSubcommandData())
	case SetPlayerLights:
		input = p.answerSetPlayerLights()
	case EnableImu:
		input = p.answerEnableImu(output.getSubcommandData())
	case EnableVibration:
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

func (p *Protocol) answerSetMode(data []byte) (input *InputReport) {
	// TODO: Update input input mode
	input = AllocStandardReport()
	input.setReportId(SubcommandReplies)
	input.fillStandardData(p.elapsed, p.deviceInfoRequired)
	input.ackSetInputReportMode()
	return
}

func (p *Protocol) anwserTriggerButtonsElapsedTime() (input *InputReport) {
	input = AllocStandardReport()
	input.setReportId(SubcommandReplies)
	input.fillStandardData(p.elapsed, p.deviceInfoRequired)
	input.ackTriggerButtonsElapsedTime()
	return
}

func (p *Protocol) answerDeviceInfo() (input *InputReport) {
	p.deviceInfoRequired = true

	input = AllocStandardReport()
	input.setReportId(SubcommandReplies)
	input.fillStandardData(p.elapsed, p.deviceInfoRequired)
	input.ackDeviceInfo(p.macAddr)
	return
}

func (p *Protocol) answerSetShipmentState() (input *InputReport) {
	input = AllocStandardReport()
	input.setReportId(SubcommandReplies)
	input.fillStandardData(p.elapsed, p.deviceInfoRequired)
	input.ackSetShipmentLowPowerState()
	return
}

func (p *Protocol) answerSpiRead(data []byte) (input *InputReport) {
	input = AllocStandardReport()
	input.setReportId(SubcommandReplies)
	input.fillStandardData(p.elapsed, p.deviceInfoRequired)
	input.ackSpiFlashRead(data)
	return
}

func (p *Protocol) answerSetNfcMcuConfig(data []byte) (input *InputReport) {
	// TODO: Update NFC MCU config
	input = AllocStandardReport()
	input.setReportId(SubcommandReplies)
	input.fillStandardData(p.elapsed, p.deviceInfoRequired)
	input.ackSetNfcMcuConfig()
	return
}

func (p *Protocol) answerSetNfcMcuState(data []byte) (input *InputReport) {
	// TODO: Update NFC MCU State
	input = AllocStandardReport()
	input.setReportId(SubcommandReplies)
	input.fillStandardData(p.elapsed, p.deviceInfoRequired)
	input.ackSetNfcMcuState()
	return
}

func (p *Protocol) answerSetPlayerLights() (input *InputReport) {
	p.playerNumber = true
	input = AllocStandardReport()
	input.setReportId(SubcommandReplies)
	input.fillStandardData(p.elapsed, p.deviceInfoRequired)
	input.ackSetPlayerLights()
	return
}

func (p *Protocol) answerEnableImu(data []byte) (input *InputReport) {
	if data[0] == 0x01 {
		p.imuEnabled = true
	}

	input = AllocStandardReport()
	input.setReportId(SubcommandReplies)
	input.fillStandardData(p.elapsed, p.deviceInfoRequired)
	input.ackEnableImu()
	return
}

func (p *Protocol) answerEnableVibration() (input *InputReport) {
	p.vibrationEnabled = true
	input = AllocStandardReport()
	input.setReportId(SubcommandReplies)
	input.fillStandardData(p.elapsed, p.deviceInfoRequired)
	input.ackEnableVibration()
	return
}

func (p *Protocol) updateTimer() {
	duration := time.Since(p.lastTime)

	p.elapsed = (p.elapsed + (duration.Microseconds() * 4)) & 0xFF
	p.lastTime = time.Now()
}
