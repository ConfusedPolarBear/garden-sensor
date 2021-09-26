#!/bin/bash
set -e

# Check that the user is in the correct directory
[[ -f firmware.ino ]] || (echo "Failed to find firmware in current directory"; exit 1)

docker run --rm --name firmware -v $PWD:/firmware -v /tmp/arduino:/tmp garden-firmware build
