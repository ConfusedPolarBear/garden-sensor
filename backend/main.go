package main

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type GardenSystem struct {
	Identifier  string
	LastReading Reading
	LastSeen    time.Time
}

type Reading struct {
	Temperature float32
	Humidity    float32
}

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
