#!/bin/bash
esptool.py --baud 408000 write_flash --flash_size 4MB 0x0 build/esp8266.esp8266.generic/firmware.ino.bin
