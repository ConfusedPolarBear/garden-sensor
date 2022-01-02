package api

import (
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/ConfusedPolarBear/garden/internal/db"
	"github.com/ConfusedPolarBear/garden/internal/util"
	"github.com/ConfusedPolarBear/garden/internal/websocket"

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
	r.HandleFunc("/system/command/{id}", SendCommandHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/system/update/{id}", StartOTA).Methods("POST", "OPTIONS")

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

func SendCommandHandler(w http.ResponseWriter, r *http.Request) {
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

	encrypt := r.Form.Has("encrypt")

	if err := sendCommand(id, command, encrypt); err != nil {
		logrus.Warnf("[server] unable to send command: %s", err)
		w.WriteHeader(http.StatusBadRequest)
	}
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

	forceError := r.Form.Get("error")
	if forceError == "host" {
		host = "127.0.0.1:6969"
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

	// Construct the short firmware download URL to use.
	shortCode := "fw32"
	if chipset == "esp8266" {
		shortCode = "fw82"
	}

	if forceError == "url" {
		shortCode = "dead"
	}

	ota.URL = fmt.Sprintf("%s/%s", host, shortCode)

	// If a forced error was requested it, make it happen
	if forceError != "" {
		logrus.Errorf("[server] injecting %s error in update for %s!", forceError, id)

		random := hex.EncodeToString(util.SecureRandom(16))

		switch forceError {
		case "ssid":
			ota.SSID = random

		case "psk":
			ota.PSK = random

		case "host", "url":
			// handled above

		case "size":
			ota.Size /= 2

		case "checksum":
			ota.Checksum = random

		default:
			logrus.Errorf("[server] error type %s is unknown", forceError)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	logrus.Debugf("[server] set OTA url to %s (used host %s)", ota.URL, host)

	pw := ota.PSK
	ota.PSK = "[redacted]"
	logrus.Debugf("[server] constructed OTA payload %#v", ota)
	ota.PSK = pw

	if err := sendCommand(id, string(util.Marshal(ota)), true); err != nil {
		logrus.Warnf("[server] unable to initiate OTA for %s: %s", id, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
