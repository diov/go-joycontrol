package joysticker

import (
	"errors"
	"syscall"
	"time"

	"dio.wtf/joysticker/joysticker/log"
	"golang.org/x/sys/unix"
)

type Protocol struct {
	lastTime           time.Time
	elapsed            int64
	reportReceived     bool
	deviceInfoRequired bool

	queue  chan *InputReport
	output *OutputReport

	itr, ctrl int
	macAddr   []byte
}

func NewProtocol() *Protocol {
	return &Protocol{
		queue:  make(chan *InputReport, 5),
		output: &OutputReport{},
	}
}

func (p *Protocol) Setup(itr, ctrl int, macAddr []byte) {
	p.itr = itr
	p.ctrl = ctrl
	p.macAddr = macAddr

	if err := unix.SetNonblock(p.itr, true); nil != err {
		log.Error(err)
		return
	}

	go p.sendEmptyReport()
	go p.processInputQueue()
	go p.readOutputReport()
}

func (p *Protocol) sendEmptyReport() {
	ticker := time.NewTicker(time.Second)

	<-ticker.C
	input := &InputReport{}
	input.reset()
	input.setReportId(StandardFullMode)
	p.queue <- input
	ticker.Stop()
}

func (p *Protocol) processInputQueue() {
	for {
		select {
		case input := <-p.queue:
			if _, err := unix.Write(p.itr, input.data[:]); nil != err {
				log.ErrorF("error writing input report: %v", err)
			} else {
				log.DebugF("input report written %s", input)
			}
			// default:
			// log.Debug("no input report to write")
		}
	}
}

func (p *Protocol) readOutputReport() {
	// TODO: use EPOLL to improve performance
	for {
		err := p.output.load(p.itr)
		if err != nil {
			switch {
			case errors.Is(err, syscall.EAGAIN):
				continue
			case errors.Is(err, errEmptyData), errors.Is(err, errBadLengthData), errors.Is(err, errMalformedData):
				// TODO: Setting Report ID to full standard input report ID
				input := &InputReport{}
				input.reset()
				input.setReportId(StandardFullMode)
				p.queue <- input
				return
			default:
				log.ErrorF("error reading output report: %v", err)
				return
			}
		}

		p.reportReceived = true
		log.DebugF("output report read %s", p.output)
		switch p.output.id {
		case RumbleAndSubcommand:
			p.processSubcommandReport(p.output)
		case UpdateNFCPacket:
		case RumbleOnly:
		case RequestNFCData:
		}
	}
}

func (p *Protocol) processSubcommandReport(report *OutputReport) {
	p.updateTimer()

	subcommand := report.getSubcommand()
	switch subcommand {
	case RequestDeviceInfo:
		p.deviceInfoRequired = true
		p.answerDeviceInfo()
	case SetInputReportMode:
		p.answerSetMode(report.getSubcommandData())
	case TriggerButtonsElapsedTime:
		p.anwserTriggerButtonsElapsedTime()
	case SetShipmentLowPowerState:
		p.answerSetShipmentState()
	case SpiFlashRead:
		p.answerSpiRead(report.getSubcommandData())
	case SetNFCMCUConfiguration:
	case SetNFCMCUState:
	case SetPlayerLights:
	case EnableIMU:
	case EnableVibration:
		p.answerEnableVibration()
	}
}

func (p *Protocol) answerSetMode(data []byte) {
	// TODO: Update input report mode
	report := &InputReport{}
	report.reset()
	report.setReportId(SubcommandReplies)
	report.fillStandardData(p.elapsed, p.deviceInfoRequired)
	report.ackSetInputReportMode()
	p.queue <- report
}

func (p *Protocol) anwserTriggerButtonsElapsedTime() {
	report := &InputReport{}
	report.reset()
	report.setReportId(SubcommandReplies)
	report.fillStandardData(p.elapsed, p.deviceInfoRequired)
	report.ackTriggerButtonsElapsedTime()
	p.queue <- report
}

func (p *Protocol) answerDeviceInfo() {
	report := &InputReport{}
	report.reset()
	report.setReportId(SubcommandReplies)
	report.fillStandardData(p.elapsed, p.deviceInfoRequired)
	report.ackDeviceInfo(p.macAddr)
	p.queue <- report
}

func (p *Protocol) answerSetShipmentState() {
	report := &InputReport{}
	report.reset()
	report.setReportId(SubcommandReplies)
	report.fillStandardData(p.elapsed, p.deviceInfoRequired)
	report.ackSetShipmentLowPowerState()
	p.queue <- report
}

func (p *Protocol) answerSpiRead(data []byte) {
	report := &InputReport{}
	report.reset()
	report.setReportId(SubcommandReplies)
	report.fillStandardData(p.elapsed, p.deviceInfoRequired)
	report.ackSpiFlashRead(data)
	p.queue <- report
}

func (p *Protocol) answerEnableVibration() {
	report := &InputReport{}
	report.reset()
	report.setReportId(SubcommandReplies)
	report.fillStandardData(p.elapsed, p.deviceInfoRequired)
	report.ackEnableVibration()
	p.queue <- report
}

func (p *Protocol) updateTimer() {
	now := time.Now()
	duration := now.Sub(p.lastTime)

	p.elapsed = int64(duration/4) & 0xFF
	p.elapsed = 0xFF
	p.lastTime = now
}
