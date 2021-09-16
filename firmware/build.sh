#!/bin/bash
set -e

echo Compiling
arduino compile -b esp8266:esp8266:generic -e

echo Done

