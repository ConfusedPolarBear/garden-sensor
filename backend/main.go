package main

import (
	"os"

	"github.com/ConfusedPolarBear/garden/internal/api"
	"github.com/ConfusedPolarBear/garden/internal/mqtt"

	"github.com/sirupsen/logrus"
)

func main() {
	setupLogrus()

	mqtt.Setup()
	api.StartServer()
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
