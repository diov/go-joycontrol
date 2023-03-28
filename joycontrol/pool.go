package joycontrol

import (
	"sync"

	R "dio.wtf/joycontrol/joycontrol/report"
)

const standardSize = 50
const nfcSize = 363

var (
	standardPool = sync.Pool{
		New: func() any {
			report := R.InputReport(make([]byte, standardSize))
			return &report
		},
	}
	nfcPool = sync.Pool{
		New: func() any {
			report := R.InputReport(make([]byte, nfcSize))
			return &report
		},
	}
	emptyInputReport = [nfcSize]byte{0xA1}
)

func AllocStandardReport() *R.InputReport {
	report := standardPool.Get().(*R.InputReport)
	copy((*report)[:], emptyInputReport[:])
	return report
}

func AllocNfcReport() *R.InputReport {
	report := nfcPool.Get().(*R.InputReport)
	copy((*report)[:], emptyInputReport[:])
	return report
}

func FreeReport(report *R.InputReport) {
	switch cap(*report) {
	case standardSize:
		standardPool.Put(report)
	case nfcSize:
		nfcPool.Put(report)
	}
}
