package main

import (
	"os"
	"github.com/ConfusedPolarBear/garden/internal/api"
	"github.com/ConfusedPolarBear/garden/internal/config"
	"github.com/ConfusedPolarBear/garden/internal/db"
	"github.com/ConfusedPolarBear/garden/internal/mqtt"

	"github.com/sirupsen/logrus"
)



func main() {
	setupLogrus()

	config.Load()
	db.InitializeDatabase()
	//db.PopulateDBForTesting()
	db.WriteReadingsToCSV()
	mqtt.Setup(true)
	api.StartServer()
}

func setupLogrus() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	if level := os.Getenv("GARDEN_DEBUG"); level != "" {
		if level == "trace" {
			logrus.SetLevel(logrus.TraceLevel)
			logrus.Trace("[app] enabled trace logging")
			return
		}

		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("[app] enabled debug logging")
	}
}
