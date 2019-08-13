package main

import (
	"automation/cmd/transmitter/app"
	_ "automation/internal/pkg/alog"
	//"github.com/sirupsen/logrus"
)

func main() {
	app.NewTransmitterCommand()
}
