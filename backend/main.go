package main

import (
	"os"

	"github.com/sirupsen/logrus"
)

func main() {
	setupLogrus()

	SetupMQTT()
	SetupAPIServer()
}

func setupLogrus() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	if os.Getenv("GARDEN_DEBUG") != "" {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("[app] Enabled debug logging")
	}
}
