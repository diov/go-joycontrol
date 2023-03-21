package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"dio.wtf/joysticker/joysticker"
	"golang.org/x/sys/unix"
)

func main() {
	server := joysticker.NewServer()
	server.Run()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}

func main2() {
	// 创建一个非阻塞的 Unix 套接字
	fd, err := unix.Socket(unix.AF_UNIX, unix.SOCK_STREAM|unix.SOCK_NONBLOCK, 0)
	if err != nil {
		panic(err)
	}

	// 创建 epoll 实例
	epfd, err := unix.EpollCreate1(0)
	if err != nil {
		panic(err)
	}

	// 将套接字添加到 epoll 实例中
	event := unix.EpollEvent{Events: unix.EPOLLIN | unix.EPOLLET, Fd: int32(fd)}
	if err := unix.EpollCtl(epfd, unix.EPOLL_CTL_ADD, fd, &event); err != nil {
		panic(err)
	}

	// 等待事件发生
	events := make([]unix.EpollEvent, 10)
	for {
		n, err := unix.EpollWait(epfd, events, -1)
		if err != nil {
			panic(err)
		}

		// 处理事件
		for i := 0; i < n; i++ {
			if events[i].Fd == int32(fd) {
				// 读取数据
				buf := make([]byte, 1024)
				n, err := unix.Read(fd, buf)
				if err != nil {
					panic(err)
				}
				fmt.Printf("Received %d bytes: %s\n", n, string(buf[:n]))
			}
		}
	}
}
