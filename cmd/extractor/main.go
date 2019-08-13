package main

import (
	"automation/parser/cmd/extractor/app"
	_ "automation/parser/internal/pkg/alog"
	//"github.com/sirupsen/logrus"
)

func main() {
	app.NewExtractCommand()
}
