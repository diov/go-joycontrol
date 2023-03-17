package joysticker

import (
	"errors"
	"fmt"
	"net"

	"golang.org/x/sys/unix"
)

func SetupSocket(addr string, channel uint16) (fd int, err error) {
	fd, err = unix.Socket(unix.AF_BLUETOOTH, unix.SOCK_SEQPACKET, unix.BTPROTO_L2CAP)
	if nil != err {
		err = fmt.Errorf("unix.Socket %s", err)
		return
	}
	// if err = unix.SetNonblock(fd, true); nil != err {
	// 	err = fmt.Errorf("unix.SetNonblock %s", err)
	// 	return
	// }
	if err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1); nil != err {
		err = fmt.Errorf("unix.SetsockoptInt %s", err)
		return
	}

	sa, _ := ParseBluetoothSockaddr(addr, channel)
	if err = unix.Bind(fd, sa); nil != err {
		err = fmt.Errorf("unix.Bind %s", err)
		return
	}
	if err = unix.Listen(fd, 1); nil != err {
		err = fmt.Errorf("unix.Listen %s", err)
		return
	}
	return
}

var errInvalidMAC = errors.New("bluetooth: Bad MAC address")

func ParseBluetoothSockaddr(addr string, channel uint16) (unix.Sockaddr, error) {
	hwAddr, _ := net.ParseMAC(addr)
	var d [6]byte
	if len(hwAddr) != 6 {
		return nil, errInvalidMAC
	}
	copy(d[:], d[:6])
	sa := &unix.SockaddrL2{
		PSM:      channel,
		Addr:     d,
		AddrType: unix.BDADDR_BREDR,
	}
	return sa, nil
}
