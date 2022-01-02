#!/bin/bash
set -e

echo "Building all firmware binaries (debug mode)"
pio run -e esp8266_debug -e esp32_debug

echo "Copying to backend server directory"
cp .pio/build/esp8266_debug/firmware.bin ../backend/data/firmware/esp8266/
cp .pio/build/esp32_debug/firmware.bin ../backend/data/firmware/esp32/

echo "Done"
