package main

import (
	"automation/parser/cmd/transmitter/app"
	_ "automation/parser/internal/pkg/alog"
	//"github.com/sirupsen/logrus"
)

func main() {
	app.NewTransmitterCommand()
}
