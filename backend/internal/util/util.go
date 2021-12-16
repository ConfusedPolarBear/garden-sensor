package util

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
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

// Creates the provided directory, returning any error except ErrExist.
func Mkdir(path string) error {
	if err := os.Mkdir(path, 0700); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	return nil
}

// Calculates the MD5 checksum of the input data or panic. Only present because the ESP8266 updater exclusively supports MD5.
//
// DO NOT USE FOR ANYTHING ELSE!
func MD5(data []byte) string {
	hasher := md5.New()

	if _, err := hasher.Write(data); err != nil {
		panic(err)
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

// Calculates the SHA256 hash of the input data or panic.
func SHA256(data []byte) string {
	sha256 := sha256.New()

	if _, err := sha256.Write(data); err != nil {
		panic(err)
	}

	return hex.EncodeToString(sha256.Sum(nil))
}
