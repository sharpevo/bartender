package main

import (
	"automation/cmd/extractor/app"
	_ "automation/internal/pkg/alog"
	//"github.com/sirupsen/logrus"
)

func main() {
	app.NewExtractCommand()
}
