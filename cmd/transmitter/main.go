package main

import (
	"automation/cmd/transmitter/app"
	_ "automation/internal/pkg/alog"
	"log"
	"os"
)

func main() {
	cmd := app.NewTransmitterCommand()
	if err := cmd.Validate(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	if err := cmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
