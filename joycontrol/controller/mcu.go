package controller

type McuMode uint8

const (
	McuStandby McuMode = 0x01
	McuNfc     McuMode = 0x04
	McuBusy    McuMode = 0x06
)

type McuPowerState uint8

const (
	McuSuspend McuPowerState = 0x00
	McuResume  McuPowerState = 0x01
)

type MicroControllerUnit struct {
	mode       McuMode
	powerState McuPowerState
}

func (m *MicroControllerUnit) SetState(state McuMode) {
	m.mode = state
}

func (m *MicroControllerUnit) TogglePowerState(on bool) {
	if on {
		m.powerState = McuResume
	} else {
		m.powerState = McuSuspend
	}
}

func (m *MicroControllerUnit) StateData() []byte {
	data := make([]byte, 8)
	data[0] = 0x01                // mcu input report id
	data[1] = 0x00                // Unknown
	data[2] = 0x00                // Unknown
	data[3], data[4] = 0x00, 0x08 // Major Firmware
	data[5], data[6] = 0x00, 0x1B // Minor Firmware
	data[7] = byte(m.mode)        // MCU State
	return data
}
