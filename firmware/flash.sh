#!/bin/bash
set -e 

# Check that the firmware has been compiled
[[ -f build/esp8266.esp8266.generic/firmware.ino.bin ]] || (echo "Failed to find compiled firmware"; exit 1)

# Check that a path to a character device is the first argument
[[ -c "$1" ]] || (echo "Usage: $0 SERIAL_DEVICE"; exit 1)

docker run --rm --name firmware -v $PWD:/firmware -v /tmp/arduino:/tmp --device "$1":/dev/ttyUSB0:rw garden-firmware flash
