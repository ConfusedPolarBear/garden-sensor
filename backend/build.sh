#!/bin/bash
set -e

go build

# TODO: look into why go build ./... doesn't build all binaries
cd cmd/emulator
go build -o ../../emulator

cd -
