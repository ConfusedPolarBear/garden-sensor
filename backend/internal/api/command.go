package api

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/ConfusedPolarBear/garden/internal/db"
	"github.com/ConfusedPolarBear/garden/internal/mqtt"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/chacha20poly1305"
)

func sendCommand(id, command string, encrypt bool) error {
	if encrypt {
		// Mesh messages are limited to 212 bytes in size. The nonce (12) + tag (16) drop that limit down to 184 bytes.
		if len(command) > 184 {
			return errors.New("encrypted commands cannot exceed 184 bytes")
		}

		config, err := db.GetConfiguration()
		if err != nil {
			return err
		}

		// Create the cipher.
		// TODO: should be cached.
		chacha, err := chacha20poly1305.New(config.ChaChaKey)
		if err != nil {
			return err
		}

		// Create the nonce and seal the box.
		// Ensure that nulls and carriage returns aren't present because the C++ firmware can't handle them correctly
		// and it will fail to authenticate the ciphertext.
		// TODO: *properly* fix this.
		var nonce, raw []byte
		forbidden := "\x00\x0d"

		for {
			nonce, raw = nil, nil

			nonce = make([]byte, chacha20poly1305.NonceSize)
			if _, err := rand.Read(nonce); err != nil {
				panic(err)
			}

			raw = chacha.Seal(raw, nonce, []byte(command), nil)

			if !bytes.ContainsAny(nonce, forbidden) && !bytes.ContainsAny(raw, forbidden) {
				break
			}
		}

		// Garden systems expect nonce + tag + ciphertext, not nonce + ciphertext + tag. Swap the bytes to account for this.
		start := len(raw) - 16
		tag, ciphertext := raw[start:], raw[:start]

		command = fmt.Sprintf("e%s%s%s", nonce, tag, ciphertext)
	}

	// Get the system
	isMesh := false
	if id != "FFFFFFFFFFFF" {
		// If this is not a broadcast message, lookup the individual system to send the message to
		system, err := db.GetSystem(id, false)
		if err != nil {
			return err
		}

		isMesh = system.Announcement.IsMesh
	} else {
		isMesh = true
	}

	logrus.Debugf("[server] sending command to %s: %s", id, command)

	mqttDest := id
	mqttPayload := command

	// If this system is connected over MQTT, send the raw command
	if isMesh {
		// Mesh connected systems are controlled by sending a command (MQTT) to the coordinator who will rebroadcast it (ESP-NOW)
		coordinator, err := db.GetCoordinator()
		if err != nil {
			return err
		}

		mqttDest = coordinator.Identifier

		// +12 bytes for the MAC address and +4 bytes for "dst-"
		logrus.Debugf("[server] mesh payload will be %d bytes long", len(mqttPayload)+12+4)

		// Construct the mesh payload
		mqttPayload = fmt.Sprintf(`{"Command":"Publish","Payload":"h%x"}`, "dst-"+id+mqttPayload)
	}

	logrus.Debugf("[server] commanding \"%s\"", mqttDest)
	logrus.Debugf("[server] mqtt payload \"%s\"", mqttPayload)

	mqtt.Publish(fmt.Sprintf("garden/module/%s/cmnd", mqttDest), mqttPayload)

	return nil
}
