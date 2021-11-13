package firmware

import (
	"archive/zip"
	_ "embed"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path"

	"github.com/ConfusedPolarBear/garden/internal/util"
	"github.com/sirupsen/logrus"
)

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
// Also extracts support files for the ESP32.
func extractFirmwareManifest() {
	// Only create the default manifest if there isn't one already
	info, err := os.Stat(manifestPath)
	if err == nil && info.Size() > 10 {
		extractEsp32Blobs()
		return
	}

	logrus.Trace("[firmware] extracting default firmware manifest")

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

	extractEsp32Blobs()
}

// The ESP32 requires three firmware blobs to successfully flash and boot: the bootloader,
// boot_app0 (tells it which partition to boot from), and the partition table. These are available in the Git repository
// as a ZIP file which the user can download separately.
func extractEsp32Blobs() {
	basePath := "data/firmware/esp32"

	// Don't do anything if the partition table exists.
	if _, err := os.Stat(path.Join(basePath, "partitions.bin")); err == nil {
		logrus.Tracef("[firmware] not extracting blobs, partitions.bin exists")
		return
	}

	logrus.Tracef("[firmware] esp32 blobs need to be extracted")

	// Open the blobs ZIP
	r, err := zip.OpenReader("data/esp32.zip")
	if err != nil {
		// TODO: link to a wiki page that explains this
		logrus.Errorf("[firmware] ESP32 firmware blobs need to be extracted but were not found locally")
		return
	}
	defer r.Close()

	logrus.Tracef("[firmware] starting blob extraction")

	// Extract every file in it to the data/firmware/esp32 directory.
	for _, compressed := range r.File {
		// Sanitize the incoming filename to prevent traversal attacks
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

		// Copy at most 17K from the file to the host.
		size := int64(math.Min(float64(compressed.UncompressedSize64), 17*1024))
		logrus.Tracef("[firmware] extracting file %s (%d bytes) into %s", n, size, dst)

		if _, err := io.CopyN(ef, f, size); err != nil {
			logrus.Errorf("[firmware] unable to extract %s: %s", n, err)
			continue
		}

		logrus.Tracef("[firmware] extracted %s successfully", n)
	}

	logrus.Debugf("[firmware] extracted ESP32 blobs")
}
