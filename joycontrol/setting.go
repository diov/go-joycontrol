package joycontrol

import (
	"bufio"
	"errors"
	"os"
	"strings"

	"dio.wtf/joycontrol/joycontrol/log"
	"github.com/muka/go-bluetooth/hw/linux/cmd"
)

const (
	servicePath  = "/lib/systemd/system/bluetooth.service"
	overrideDir  = "/run/systemd/system/bluetooth.service.d"
	overridePath = overrideDir + "/nxbt.conf"
)

func toggleCleanBluez(flag bool) {
	ret, err := cmd.Exec("ps", "--no-headers", "-o", "comm", "1")
	if nil != err || strings.TrimSpace(ret) != "systemd" {
		return
	}

	if flag {
		if _, err := os.Stat(overridePath); nil != err && !errors.Is(err, os.ErrNotExist) {
			// Override exist, no need to restart bluetooth
			return
		}

		file, _ := os.Open(servicePath)
		defer file.Close()
		scanner := bufio.NewScanner(file)
		execStart := ""
		for scanner.Scan() {
			if strings.HasPrefix(scanner.Text(), "ExecStart=") {
				execStart = strings.Trim(scanner.Text(), " ") + " --compat --noplugin=*"
			}
			// TODO: Throw error
		}
		override := "[Service]\nExecStart=\n" + execStart

		os.MkdirAll(overrideDir, os.ModePerm)
		if err = os.WriteFile(overridePath, []byte(override), 0644); nil != err {
			log.Error(err)
		}
		log.Debug("Override conf")
	} else {
		os.Remove(overridePath)
		log.Debug("Remove conf")
	}

	cmd.Exec("systemctl", "daemon-reload")
	cmd.Exec("systemctl", "restart", "bluetooth")
	log.Debug("systemd found and bluetooth reloaded")
}

func crc8Checksum(bytes []byte) byte {
	polynomial := byte(0x07)
	accumulator := byte(0)

	for j := 0; j < len(bytes)-1; j++ {
		b := bytes[j]
		accumulator = accumulator ^ b
		for i := 0; i < 8; i++ {
			if (accumulator & 0x80) != 0x00 {
				accumulator = (accumulator << 1) ^ polynomial
			} else {
				accumulator = accumulator << 1
			}
		}
	}
	return accumulator
}
