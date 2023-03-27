package main

import (
	"os"
	"os/signal"
	"syscall"

	"dio.wtf/joycontrol/joycontrol"
)

func main() {
	server := joycontrol.NewServer()
	server.Start()
	defer server.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
