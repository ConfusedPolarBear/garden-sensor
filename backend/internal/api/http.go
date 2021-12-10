package api

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ConfusedPolarBear/garden/internal/db"
	"github.com/ConfusedPolarBear/garden/internal/mqtt"
	"github.com/ConfusedPolarBear/garden/internal/util"
	"github.com/ConfusedPolarBear/garden/internal/websocket"
	"golang.org/x/crypto/chacha20poly1305"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	bind := "0.0.0.0:8081"

	r := mux.NewRouter()
	r.Use(corsMiddleware)

	r.HandleFunc("/ping", PingHandler).Methods("GET")

	r.HandleFunc("/systems", GetSystems).Methods("GET")
	r.HandleFunc("/system/{id}", GetSystem).Methods("GET")
	r.HandleFunc("/system/delete/{id}", DeleteSystem).Methods("POST")
	r.HandleFunc("/system/command/{id}", SendCommand).Methods("POST", "OPTIONS")

	r.HandleFunc("/firmware/manifest.json", ManifestHandler).Methods("GET")
	r.HandleFunc("/firmware/{board}/{file}", DownloadFirmware).Methods("GET")

	r.HandleFunc("/socket", websocket.WebSocketHandler)

	logrus.Printf("[server] API server listening on http://%s", bind)
	if err := http.ListenAndServe(bind, r); err != nil {
		panic(err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// TODO: add authorization
		// w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		next.ServeHTTP(w, r)
	})
}

func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func GetSystems(w http.ResponseWriter, r *http.Request) {
	w.Write(util.Marshal(db.GetAllSystems()))
}

func GetSystem(w http.ResponseWriter, r *http.Request) {
	id, err := getId(w, r)
	if err != nil {
		return
	}

	system, err := db.GetSystem(id, true)
	if err != nil {
		logrus.Warnf("[api] error getting system %s: %s", id, err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Write(util.Marshal(system))
}

func DeleteSystem(w http.ResponseWriter, r *http.Request) {
	id, err := getId(w, r)
	if err != nil {
		return
	}

	if err := db.DeleteSystem(id); err != nil {
		logrus.Warnf("[server] unable to delete system %s: %s", id, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func SendCommand(w http.ResponseWriter, r *http.Request) {
	// Command sent to a coordinator to publish an arbitrary mesh message.
	type meshPublishCommand struct {
		// Command as seen by the coordinator. Must be "Publish".
		Command string

		// Payload to publish. Must be deserializable as JSON.
		Payload string
	}

	id, err := getId(w, r)
	if err != nil {
		return
	}

	// Parse the form and extract the command
	if err := r.ParseForm(); err != nil {
		logrus.Warnf("[server] unable to parse form: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	command := r.Form.Get("command")
	if command == "" || len(command) > 210 {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}

	// If the user passes a hex encoded key, use that to encrypt the command.
	if rKey := r.Form.Get("key"); len(rKey) == 64 {
		logrus.Debugf("[server] encrypting command")

		// Mesh messages are limited to 212 bytes in size. The nonce (12) + tag (16) drop that limit down to 184 bytes.
		if len(command) > 184 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// User is expected to pass a hex encoded key
		key, err := hex.DecodeString(rKey)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Create the cipher.
		// TODO: should be cached.
		chacha, err := chacha20poly1305.New(key)
		if err != nil {
			logrus.Warnf("[crypto] unable to initialize cipher: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Create the nonce and seal the box.
		// Ensure that nulls and carriage returns aren't present because the C++ firmware can't handle them correctly
		// and it will fail to authenticate the ciphertext.
		// TODO: root cause this and *properly* fix it.
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

		// logrus.Debugf("nonce is %X, tag is %X, and ciphertext is %X", nonce, tag, ciphertext)
	}

	// Get the system
	isMesh := false
	if id != "FFFFFFFFFFFF" {
		// If this is not a broadcast message, lookup the individual system to send the message to
		system, err := db.GetSystem(id, false)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
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
		logrus.Debugf("[server] system %s networking mode is mesh", id)

		// Lookup the coordinator for this system.
		coordinator, err := db.GetCoordinator()
		if err != nil {
			logrus.Errorf("[server] unable to find coordinator: %s", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		mqttDest = coordinator.Identifier

		// Construct the mesh payload
		// {"Command":"Publish", "Payload": "{'D':'dst-123457890AB','Command':'Ping'}"}

		// If this is an unencrypted message, insert the destination info as a key.
		if strings.HasPrefix(mqttPayload, "{") {
			// Unmarshal the command in order to add the destination key to it
			var rawCommand map[string]interface{}
			if err := json.Unmarshal([]byte(command), &rawCommand); err != nil {
				logrus.Errorf("[server] unable to unmarshal command as JSON: %s", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			rawCommand["D"] = fmt.Sprintf("dst-%s", id)

			// Remarshal the command in coordinator format.
			meshCommand := util.Marshal(meshPublishCommand{
				Command: "Publish",
				Payload: string(util.Marshal(rawCommand)),
			})

			mqttPayload = string(meshCommand)

		} else {
			// Since this is an encrypted payload, just prepend the id of the final system.
			mqttPayload = fmt.Sprintf(`{"Command":"Publish","Payload":"h%x"}`, "dst-"+id+mqttPayload)
		}

	} else {
		logrus.Debugf("[server] system %s networking mode is wifi", id)
	}

	logrus.Debugf("[server] commanding \"%s\"", mqttDest)
	logrus.Debugf("[server] mqtt payload is \"%s\"", mqttPayload)

	mqtt.Publish(fmt.Sprintf("garden/module/%s/cmnd", mqttDest), mqttPayload)
}
