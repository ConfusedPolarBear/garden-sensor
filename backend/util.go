package main

import "encoding/json"

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

// Marshal v or panic.
func Marshal(v interface{}) []byte {
	if data, err := json.Marshal(v); err != nil {
		panic(err)
	} else {
		return data
	}
}
