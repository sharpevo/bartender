package main

import (
	"automation/cmd/extractor/app"
	_ "automation/internal/pkg/alog"
	"log"
	"os"
)

func main() {
	cmd := app.NewExtractCommand()
	if err := cmd.Validate(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	if err := cmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
