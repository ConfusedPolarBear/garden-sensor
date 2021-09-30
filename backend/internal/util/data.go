package util

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

var systems map[string]GardenSystem = make(map[string]GardenSystem)

// Mutex used to guard access to the map of systems. This is because concurrent map reads and writes are illegal in Go.
// While there is a dedicated sync.Map type, using a plain Mutex here is preferred (ref https://pkg.go.dev/sync#Map).
var systemLock sync.Mutex

// Updates the provided system.
func UpdateSystem(data GardenSystem) {
	systemLock.Lock()
	defer systemLock.Unlock()

	id := data.Identifier
	if _, okay := systems[id]; !okay {
		logrus.Debugf("[mqtt] registered system %s", id)
	}

	systems[id] = data
}

// Returns the system with the provided id or an error if an invalid id was provided.
func GetSystem(id string) (GardenSystem, error) {
	var err error

	systemLock.Lock()
	defer systemLock.Unlock()

	system, found := systems[id]
	if !found {
		err = fmt.Errorf("unable to find system with id %s", id)
	}

	return system, err
}

// Clients are internally stored as a map of client identifier to the client's data but it's more useful to the frontend
// as an array.
func SystemMapToSlice() []GardenSystem {
	var clients []GardenSystem

	systemLock.Lock()
	defer systemLock.Unlock()

	for _, c := range systems {
		clients = append(clients, c)
	}

	return clients
}
