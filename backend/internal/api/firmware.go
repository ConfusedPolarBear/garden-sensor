package api

import (
	_ "embed"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

//go:embed firmware_manifest.json
var defaultManifest []byte
var manifestPath string = "data/firmware/manifest.json"

func checkFirmwareManifest() {
	// Only create the default manifest if there isn't one already
	info, err := os.Stat(manifestPath)
	if err == nil && info.Size() > 10 {
		return
	}

	// TODO: wrap os.Mkdir() as a common function that logs a fatal error if a directory can't be created
	os.Mkdir("data/firmware", 0700)

	if err = ioutil.WriteFile(manifestPath, defaultManifest, 0700); err != nil {
		logrus.Warnf("[server] unable to extract default firmware manifest: %s", err)
		return
	}

	logrus.Debug("[server] extracted default firmware manifest")
}

func ManifestHandler(w http.ResponseWriter, _ *http.Request) {
	w.Write(getFirmwareManifest())
}

func DownloadFirmware(w http.ResponseWriter, r *http.Request) {
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

// Return the contents of the manifest file to the requesting client.
func getFirmwareManifest() []byte {
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		panic(err)
	}

	return raw
}
