package util

import (
	"encoding/json"
	"errors"
	"os"
)

// Marshal v or panic.
func Marshal(v interface{}) []byte {
	if data, err := json.Marshal(v); err != nil {
		panic(err)
	} else {
		return data
	}
}

func Mkdir(path string) error {
	if err := os.Mkdir(path, 0700); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	return nil
}
