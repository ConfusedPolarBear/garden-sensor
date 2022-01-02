package db

import (
	"encoding/base64"

	"github.com/ConfusedPolarBear/garden/internal/util"
	"github.com/sirupsen/logrus"
)

func initializeConfig() {
	var config util.Configuration
	db.Limit(1).Find(&config)

	if config.MeshKey != "" {
		return
	}

	logrus.Print("[db] first run detected, initializing configuration")

	config.MeshKey = base64.RawStdEncoding.EncodeToString(util.SecureRandom(48))

	UpdateConfiguration(config)
}

func GetConfiguration() (util.Configuration, error) {
	var config util.Configuration

	err := db.First(&config).Error

	if err == nil {
		config.ChaChaKey = util.DeriveKey("chacha-symmetric-key", config.MeshKey)
	}

	return config, err
}

func UpdateConfiguration(config util.Configuration) error {
	err := db.Save(&config).Error

	if err != nil {
		logrus.Errorf("[db] unable to save configuration: %s", err)
	}

	return err
}
