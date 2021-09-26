package main

import (
	"os"

	"github.com/ConfusedPolarBear/garden/internal/api"
	"github.com/ConfusedPolarBear/garden/internal/config"
	"github.com/ConfusedPolarBear/garden/internal/mqtt"

	"github.com/sirupsen/logrus"
)

func main() {
	setupLogrus()

	config.Load()
	mqtt.Setup(true)
	api.StartServer()
}

func setupLogrus() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	if os.Getenv("GARDEN_DEBUG") != "" {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("[app] enabled debug logging")
	}
}
