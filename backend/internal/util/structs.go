package util

import (
	"time"
	"gorm.io/gorm"
)

// Structs can be generated from JSON strings with https://mholt.github.io/json-to-go/

type GardenSystem struct {
	gorm.Model
	Identifier   string
	Announcement GardenSystemInfo

	// TODO: track alerts per system. possible alerts: flash size mismatch, temperature/humidity out of bounds, etc.

	LastReading Reading
	LastSeen    time.Time
}

type GardenSystemInfo struct {
	gorm.Model
	RestartReason string `json:"RestartReason"`
	CoreVersion   string `json:"CoreVersion"`
	SdkVersion    string `json:"SdkVersion"`
	FlashSize     int    `json:"FlashSize"`
	RealFlashSize int    `json:"RealFlashSize"`
	//Sensors []string `json:"Sensors"` TODO add this to the model
}

type Reading struct {
	gorm.Model
	Temperature float32
	Humidity    float32
	Time time.Time
}

//Table "Readings":
// PK identifier
// Temperature
// Humidity
// Time of reading

//Table "GardenSystem"
// Everything in a garden system, identifier as primary key

