package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Loads configuration or panics.
func Load() {
	viper.SetConfigName("garden")
	viper.SetConfigType("ini")

	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		logrus.Fatalf("[app] unable to load configuration: %s", err)
	}
}

// Gets the configuration value with the provided key.
func GetString(key string) string {
	return viper.GetString(key)
}
