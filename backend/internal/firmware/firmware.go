package firmware

import (
	"archive/zip"
	_ "embed"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/ConfusedPolarBear/garden/internal/config"
	"github.com/ConfusedPolarBear/garden/internal/util"

	"github.com/sirupsen/logrus"
)

var defaultBlobUrl string = "https://raw.githubusercontent.com/ConfusedPolarBear/garden-sensor/main/esp32/esp32.zip"
var defaultBlobChecksum string = "1ce366054001f1c71dc9bad23be38398c050b89670b91df20218c5aded8ae96f"

// Path to the firmware manifest on the host.
var manifestPath string = "data/firmware/manifest.json"

//go:embed default_manifest.json
var defaultManifest []byte

// Returns the contents of the firmware manifest. Will create a default manifest file if necessary.
func ReadFirmwareManifest() []byte {
	extractFirmwareManifest()

	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		panic(err)
	}

	return raw
}

// If a firmware manifest cannot be found, extract the default one embedded into the server binary.
// Also extracts support files for the ESP32. These are downloaded from the Git repository, verified, and extracted.
func extractFirmwareManifest() {
	// Only create the default manifest if there isn't one already
	info, err := os.Stat(manifestPath)
	if err == nil && info.Size() > 10 {
		extractEsp32Blobs()
		return
	}

	logrus.Trace("[firmware] extracting default firmware manifest")

	// Create the destination directory and extract the embedded manifest
	if err := util.Mkdir("data/firmware"); err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile(manifestPath, defaultManifest, 0700); err != nil {
		panic(err)
	}

	logrus.Debug("[firmware] extracted default firmware manifest")

	// Create destination directories for all supported boards
	for _, d := range []string{"esp8266", "esp32"} {
		if err := util.Mkdir("data/firmware/" + d); err != nil {
			logrus.Errorf("[firmware] unable to create %s firmware directory: %s", d, err)
			return
		}
	}

	// Since the manifest had to be extracted, odds are that the ESP32 blobs haven't been downloaded yet either
	extractEsp32Blobs()
}

// The ESP32 requires three firmware blobs to successfully flash and boot: the bootloader,
// boot_app0 (tells it which partition to boot from), and the partition table. These are available in the Git repository
// as a ZIP file which the user can download separately.
func extractEsp32Blobs() {
	basePath := "data/firmware/esp32"

	// Don't do anything if it looks like blobs have already been extracted.
	if _, err := os.Stat(path.Join(basePath, "partitions.bin")); err == nil {
		logrus.Tracef("[firmware] not downloading blobs, partitions.bin exists")
		return
	}

	logrus.Tracef("[firmware] esp32 blobs need to be extracted")

	// Download the blobs ZIP
	if err := downloadEsp32Blobs(); err != nil {
		logrus.Errorf("[firmware] unable to download ESP32 firmware blobs: %s", err)
		return
	}

	r, err := zip.OpenReader("data/esp32.zip")
	if err != nil {
		logrus.Errorf("[firmware] failed to extract ESP32 firmware blobs: %s", err)
		return
	}
	defer r.Close()

	logrus.Tracef("[firmware] starting blob extraction")

	// Extract every file from the ZIP to the data/firmware/esp32 directory.
	for _, compressed := range r.File {
		// Sanitize the incoming filename to prevent directory traversal attacks
		n := path.Base(compressed.Name)
		dst := path.Join(basePath, n)

		// Open the file in the archive
		f, err := compressed.Open()
		if err != nil {
			logrus.Errorf("[firmware] unable to open %s from archive: %s", n, err)
			continue
		}
		defer f.Close()

		// Open the destination file on the host. ef = extracted file
		ef, err := os.Create(dst)
		if err != nil {
			logrus.Errorf("[firmware] unable to create %s: %s", dst, err)
			continue
		}
		defer ef.Close()

		logrus.Tracef("[firmware] extracting file %s into %s", n, dst)

		if _, err := io.Copy(ef, f); err != nil {
			logrus.Errorf("[firmware] unable to extract %s: %s", n, err)
			continue
		}

		logrus.Tracef("[firmware] extracted %s successfully", n)
	}

	logrus.Debugf("[firmware] extracted ESP32 blobs")
}

// Download and verify the ESP32 blobs.
func downloadEsp32Blobs() error {
	// Get the user specified blob URL or use the default if none specified.
	url := config.GetString("esp32.url")

	if url == "" {
		url = defaultBlobUrl
	}

	// Get the user specified blob SHA256 checksum or use the default if none specified.
	expectedHash := config.GetString("esp32.hash")

	if expectedHash == "" {
		expectedHash = defaultBlobChecksum
	}

	// If the blobs have already been downloaded, don't download them again
	if _, err := os.Stat("data/esp32.zip"); err == nil {
		logrus.Tracef("[firmware] esp32 blobs already downloaded")
		return nil
	}

	// Download the blob ZIP archive
	logrus.Print("[firmware] esp32 blobs not found locally, downloading")
	logrus.Debugf("[firmware] downloading blobs from %s and verifying sha256 is %s", url, expectedHash) // TODO: change to trace

	http.DefaultClient.Timeout = 15 * time.Second
	res, err := http.Get(url)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	// Read at most 1M from the HTTP response as there's no reason the blobs should be larger than that.
	// The stock blobs are a partition table (3K), OTA partition (8K), and a bootloader (~17K) for a total of ~37K.
	r := io.LimitReader(res.Body, 1024*1024)

	blobs, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	logrus.Trace("[firmware] blobs downloaded, verifying checksum")

	if actualHash := util.SHA256(blobs); actualHash != expectedHash {
		return fmt.Errorf("archive checksums do not match: expected %s, actual %s", expectedHash, actualHash)
	}

	logrus.Trace("[firmware] blob checksum verified")

	if err := os.WriteFile("data/esp32.zip", blobs, 0600); err != nil {
		return err
	}

	return nil
}
