package util

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"strings"
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

// Returns a securely generated random buffer of length n.
func SecureRandom(n int) []byte {
	buf := make([]byte, n)

	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}

	return buf
}

func DeriveKey(purpose, root string) []byte {
	h := hmac.New(sha256.New, []byte(root))

	if _, err := h.Write([]byte(purpose)); err != nil {
		panic(err)
	}

	return h.Sum(nil)
}

func IdentifierToAddress(raw string) string {
	addr := ""
	for i := 0; i < 12; i += 2 {
		addr += raw[i:i+2] + ":"
	}
	addr = strings.TrimSuffix(addr, ":")

	return addr
}
