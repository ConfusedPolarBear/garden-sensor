package api

import (
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"time"

	"github.com/ConfusedPolarBear/garden/internal/db"
	"github.com/ConfusedPolarBear/garden/internal/firmware"
	"github.com/ConfusedPolarBear/garden/internal/util"
	"github.com/ConfusedPolarBear/garden/internal/websocket"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func ManifestHandler(w http.ResponseWriter, _ *http.Request) {
	w.Write(firmware.ReadFirmwareManifest())
}

func DownloadFirmware(w http.ResponseWriter, r *http.Request) {
	// If this is a garden system downloading this binary, broadcast the fact that it just started downloading.
	id := r.Header.Get("System-ID")
	if util.SystemIdentifierRegex.MatchString(id) {
		system, err := db.GetSystem(id, false)
		if err == nil {
			system.UpdatedAt = time.Now()
			system.UpdateStatus = util.OTAStatus{
				Success: true,
				Message: fmt.Sprintf("Backend: device %s started downloading update", id),
			}
			websocket.BroadcastWebsocketMessage("update", system)

			// no need to call db.UpdateSystem() as UpdateStatus isn't stored persistently
		}
	}

	// Test if this is a short URL handler.
	name := mux.CurrentRoute(r).GetName()
	if name == "esp8266" || name == "esp32" {
		sendFirmware(w, name, "firmware.bin")
		return
	}

	// Since this is not a short URL handler, extract the parameters from the route & send that instead.

	// Extract and validate the board name
	board := mux.Vars(r)["board"]
	if board != "esp32" && board != "esp8266" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Extract and validate the filename
	file := mux.Vars(r)["file"]
	valid := regexp.MustCompile(`^[a-zA-Z0-9\-_]{1,32}\.bin$`)
	if !valid.MatchString(file) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sendFirmware(w, board, file)
}

func sendFirmware(w http.ResponseWriter, board, file string) {
	// Open the firmware binary
	p := path.Join("data/firmware", board, file)

	f, err := os.Open(p)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	defer f.Close()

	// Send the appropiate content-length header
	info, err := f.Stat()
	if err != nil {
		logrus.Warnf("[server] unable to stat %s: %s", p, err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))

	// Serve the firmware
	_, err = io.Copy(w, f)
	if err != nil {
		logrus.Warnf("[server] unable to send firmware: %s", err)
	}
}
