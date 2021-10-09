package main

import (
	"os"
	"log"
	"gorm.io/gorm"
	
	"github.com/ConfusedPolarBear/garden/internal/api"
	"github.com/ConfusedPolarBear/garden/internal/config"
	"github.com/ConfusedPolarBear/garden/internal/mqtt"
	"github.com/ConfusedPolarBear/garden/internal/dbconn"
	"github.com/ConfusedPolarBear/garden/internal/util"
	"github.com/sirupsen/logrus"

)

func main() {
	log.Println("Connecting to database...")
	db, err := dbconn.Open()
	if err != nil {
		panic(err)
	}
	migrateTables(db)

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

func migrateTables(db *gorm.DB) {
	db.AutoMigrate(&util.Reading{})
	db.AutoMigrate(&util.GardenSystemInfo{})
}
