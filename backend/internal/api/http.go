package api

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
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
	r.HandleFunc("/system/update/{id}", StartOTA).Methods("POST")

	r.HandleFunc("/firmware/manifest.json", ManifestHandler).Methods("GET")
	r.HandleFunc("/firmware/{board}/{file}", DownloadFirmware).Methods("GET")

	// Short URLs to download firmware from. Added to save space in marshalled update commands.
	r.HandleFunc("/fw82", DownloadFirmware).Methods("GET").Name("esp8266")
	r.HandleFunc("/fw32", DownloadFirmware).Methods("GET").Name("esp32")

	r.HandleFunc("/mesh/info", MeshInfoHandler).Methods("GET", "OPTIONS")

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

	// If the user wants to encrypt the command, do that now.
	if r.Form.Has("encrypt") {
		// Mesh messages are limited to 212 bytes in size. The nonce (12) + tag (16) drop that limit down to 184 bytes.
		if len(command) > 184 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		config, err := db.GetConfiguration()
		if err != nil {
			logrus.Warnf("[server] unable to get configuration: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Create the cipher.
		// TODO: should be cached.
		chacha, err := chacha20poly1305.New(config.ChaChaKey)
		if err != nil {
			logrus.Warnf("[crypto] unable to initialize cipher: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
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
		coordinator, err := db.GetCoordinator()
		if err != nil {
			logrus.Errorf("[server] unable to find coordinator: %s", err)
			w.WriteHeader(http.StatusNotFound)
			return
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
}

func StartOTA(w http.ResponseWriter, r *http.Request) {
	type otaCommand struct {
		Command  string
		SSID     string `json:"S"`
		PSK      string `json:"P"`
		URL      string `json:"U"`
		Size     int64  `json:"L"`
		Checksum string `json:"C"`
	}

	ota := otaCommand{Command: "Update"}

	// Get the system that is going to be updated & validate its information
	id, err := getId(w, r)
	if err != nil {
		return
	}

	system, err := db.GetSystem(id, false)
	if err != nil {
		logrus.Warnf("[server] unable to get system with id %s: %s", id, err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	chipset := strings.ToLower(system.Announcement.Chipset)
	if chipset != "esp8266" && chipset != "esp32" {
		logrus.Warnf("[server] system %s has unknown chipset %s", id, chipset)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get the Wi-Fi SSID & PSK
	if err := r.ParseForm(); err != nil {
		logrus.Warnf("[server] unable to parse ota form: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ota.SSID, ota.PSK = r.Form.Get("ssid"), r.Form.Get("psk")

	if ota.SSID == "" || len(ota.SSID) > 32 || ota.PSK == "" || len(ota.PSK) > 64 {
		logrus.Warn("[server] ssid or psk are invalid")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get the server (if specified), otherwise fall back to the host
	host := r.Form.Get("host")
	if host == "" {
		host = r.Host
		logrus.Warn("[server] no host specified for OTA, falling back to HTTP host.")

		if strings.HasPrefix(host, "127.0.0.1") {
			logrus.Warnf("[server] HTTP host is %s, which is inaccessible for systems to update from.", host)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	host = strings.TrimSuffix(host, "/")

	// Read the latest firmware binary to get its length & checksum.
	fw := path.Join("data/firmware/", chipset, "firmware.bin")

	f, err := os.Open(fw)
	if err != nil {
		logrus.Warnf("[server] unable to open firmware for %s (chipset %s) at %s: %s", id, chipset, fw, err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	defer f.Close()

	if stat, err := f.Stat(); err != nil {
		logrus.Warnf("[server] unable to stat firmware for %s (chipset %s) at %s: %s", id, chipset, fw, err)
		w.WriteHeader(http.StatusNotFound)
		return
	} else {
		ota.Size = stat.Size()
	}

	// Calculate the checksum with MD5 (terrible, but it's the best algorithm natively supported).
	contents, err := io.ReadAll(f)
	if err != nil {
		logrus.Warnf("[server] unable to read firmware for %s (chipset %s) at %s: %s", id, chipset, fw, err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ota.Checksum = util.MD5(contents)

	pw := ota.PSK
	ota.PSK = "[redacted]"
	logrus.Debugf("[server] constructed OTA payload %#v", ota)
	ota.PSK = pw

	// Construct the short firmware download URL to use.
	shortCode := "fw32"
	if chipset == "esp8266" {
		shortCode = "fw82"
	}

	ota.URL = fmt.Sprintf("%s/%s", host, shortCode)
	logrus.Debugf("[server] set OTA url to %s (used host %s)", ota.URL, host)

	// TODO: publish the command directly
	w.Write(util.Marshal(ota))
}
