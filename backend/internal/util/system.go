package util

import (
	"time"
)

// Structs can be generated from JSON strings with https://mholt.github.io/json-to-go/
// TODO: track alerts per system. possible alerts: flash size mismatch, temperature/humidity out of bounds, etc.

type GardenSystem struct {
	Identifier string `gorm:"primaryKey;notNull"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  time.Time

	Announcement GardenSystemInfo

	LastReading Reading `gorm:"-"`
	Readings    []Reading
}

type GardenSystemInfo struct {
	// Parent garden system that generated this announcement.
	GardenSystemID string `gorm:"primaryKey"`

	// If this garden system is an actually an emulator. This field should not be sent by non-virtual systems.
	IsEmulator bool

	// If this system is connected through the mesh or MQTT.
	IsMesh        bool
	RestartReason string

	CoreVersion string
	SdkVersion  string

	FilesystemUsedSize  int
	FilesystemTotalSize int

	Sensors []Sensor
}

type Reading struct {
	// Internal identifier for this reading.
	ID        uint `gorm:"autoIncrement"`
	CreatedAt time.Time

	// Parent garden system that generated this reading.
	GardenSystemID string
	Error          bool
	Temperature    float32
	Humidity       float32
}
