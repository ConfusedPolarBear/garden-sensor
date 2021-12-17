package util

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

var hashData string = "May the sinful saxophones of devils echo through the hall with dreadful melodies of waltz, tango and quickstep."

func TestMD5(t *testing.T) {
	expected := "8b3f18e7b961910fedd4370d588f269e"
	actual := MD5([]byte(hashData))
	assert.Equal(t, expected, actual)
}

func TestSHA256(t *testing.T) {
	expected := "78f4209ef4a409b8f13fa8a4ce95aea2cac870315299eac7e40fba431b35f126"
	actual := SHA256([]byte(hashData))
	assert.Equal(t, expected, actual)
}

func TestIdentifierToAddress(t *testing.T) {
	expected := "84:cc:a8:ab:cd:ef"
	actual := IdentifierToAddress("84cca8abcdef")
	assert.Equal(t, expected, actual)
}

func TestKeyDerivation(t *testing.T) {
	expected, _ := hex.DecodeString("d8eeaa25ed390dfdcad45f24697c45e94e4ee1788c67335aac0b287bb66ea4f0")
	actual := DeriveKey("chacha-symmetric-key", "4B5DDWMTG346NBVFNIO4MPQ644RIBF52MJM6VATLH3DS2HPT76MF24TV5X7IMSI")
	assert.Equal(t, expected, actual)
}
