package main

import (
	"log"
	"os"

	"github.com/sharpevo/bartender/cmd/transmitter/app"
	_ "github.com/sharpevo/bartender/internal/pkg/alog"
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
