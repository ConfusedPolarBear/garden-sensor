package util

import "encoding/json"

type Sensor struct {
	GardenSystemInfoID string `gorm:"uniqueIndex:idx_sensors"`
	Name               string `gorm:"uniqueIndex:idx_sensors"`
}

// Sensors sent by the firmware are just an array of strings but that doesn't work with gorm.
func (s *Sensor) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &s.Name)
}

func (s Sensor) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Name)
}
