package main

import (
	"LAN_Transfer/service"
)

func main() {
	service.InitWidget()
	service.Log("Init Widget Success")
	service.Receiver.InitSetting()
	service.Sender.InitSetting()
	service.Sender.RunIpSearcher()
	service.MainWindow.ShowAndRun()
}
