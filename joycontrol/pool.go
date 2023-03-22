package joycontrol

import "sync"

const standardSize = 50
const nfcSize = 363

var (
	standardPool = sync.Pool{
		New: func() any {
			report := InputReport(make([]byte, standardSize))
			return &report
		},
	}
	nfcPool = sync.Pool{
		New: func() any {
			report := InputReport(make([]byte, nfcSize))
			return &report
		},
	}
	emptyInputReport = [nfcSize]byte{0xA1}
)

func AllocStandardReport() *InputReport {
	report := standardPool.Get().(*InputReport)
	copy((*report)[:], emptyInputReport[:])
	return report
}

func AllocNfcReport() *InputReport {
	report := nfcPool.Get().(*InputReport)
	copy((*report)[:], emptyInputReport[:])
	return report
}

func FreeReport(report *InputReport) {
	switch cap(*report) {
	case standardSize:
		standardPool.Put(report)
	case nfcSize:
		nfcPool.Put(report)
	}
}
