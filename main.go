package main

import "dio.wtf/joysticker/joysticker"

func main() {
	server := joysticker.NewServer()
	server.Run()
}
