package util

import (
	"time"
)

// Structs can be generated from JSON strings with https://mholt.github.io/json-to-go/

type GardenSystem struct {
	Identifier   string
	Announcement GardenSystemInfo

	// TODO: track alerts per system. possible alerts: flash size mismatch, temperature/humidity out of bounds, etc.

	LastReading Reading
	LastSeen    time.Time
}

type GardenSystemInfo struct {
	SystemInfo struct {
		// If this garden system is an actually an emulator. This field should not be sent by non-virtual systems.
		IsEmulator    bool
		RestartReason string `json:"RestartReason"`

		CoreVersion string `json:"CoreVersion"`
		SdkVersion  string `json:"SdkVersion"`

		FlashSize     int `json:"FlashSize"`
		RealFlashSize int `json:"RealFlashSize"`

		FilesystemUsedSize  int
		FilesystemTotalSize int
	} `json:"System"`
	Sensors []string `json:"Sensors"`
}

type Reading struct {
	Temperature float32
	Humidity    float32
}
