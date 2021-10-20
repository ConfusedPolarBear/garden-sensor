package util

import (
	"time"
	"gorm.io/gorm"
)

// Structs can be generated from JSON strings with https://mholt.github.io/json-to-go/

type GardenSystem struct {
	gorm.Model
	Identifier   string `json:"Identifier"`
	Announcement GardenSystemInfo`gorm:"foreignKey:Identifier"`
	Readings []Reading`gorm:"foreignKey:Identifier"`
	LastSeen    time.Time
}

type GardenSystemInfo struct {
	gorm.Model
	Identifier   string  `json:"Identifier"`
	RestartReason string `json:"RestartReason"`
	CoreVersion   string `json:"CoreVersion"`
	SdkVersion    string `json:"SdkVersion"`
	FlashSize     int    `json:"FlashSize"`
	RealFlashSize int    `json:"RealFlashSize"`
	//Sensors []string `json:"Sensors"` TODO add this to the model
}

type Reading struct {
	gorm.Model
	Identifier   string `json:"Identifier"`
	Temperature float32 `json:"Temperature"`
	Humidity    float32	`json:"Humidity"`
	Time time.Time
}

//Table "Readings":
// PK identifier
// Temperature
// Humidity
// Time of reading

//Table "GardenSystem"
// Everything in a garden system, identifier as primary key

