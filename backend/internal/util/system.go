package util

import (
	"regexp"
	"time"
)

// Structs can be generated from JSON strings with https://mholt.github.io/json-to-go/
// TODO: track alerts per system. possible alerts: flash size mismatch, temperature/humidity out of bounds, etc.

// Regular expression that all incoming system identifiers must match.
var SystemIdentifierRegex regexp.Regexp = *regexp.MustCompile("^[a-fA-F0-9]{12}$")

type GardenSystem struct {
	Identifier string `gorm:"primaryKey;notNull"`
	Name string
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
	IsMesh bool

	// The channel that the Wi-Fi station uses. Only valid if this is the controller.
	// For the mesh to be reliable, it must use the same channel as Wi-Fi.
	Channel       int
	RestartReason string

	CoreVersion string
	SdkVersion  string

	FilesystemUsedSize  int
	FilesystemTotalSize int

	Sensors []Sensor
}

type Reading struct {
	CreatedAt time.Time

	// Parent garden system that generated this reading.
	GardenSystemID string
	Error          bool
	Temperature    float32
	Humidity       float32
}
