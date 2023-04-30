package amiibo

import "sync"

var once sync.Once

func init() {
	once.Do(func() {
		mcu = &MicroControllerUnit{}
	})
}

var mcu *MicroControllerUnit

type MicroControllerUnit struct {
}
