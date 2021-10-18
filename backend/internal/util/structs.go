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
	System struct {
		// If this garden system is an actually an emulator. This field should not be sent by non-virtual systems.
		IsEmulator bool

		// If this system is connected through the mesh.
		IsMesh        bool
		RestartReason string

		CoreVersion string
		SdkVersion  string

		FilesystemUsedSize  int
		FilesystemTotalSize int
	}
	Sensors []string
}

type Reading struct {
	Temperature float32
	Humidity    float32
}
