#pragma once

#include "firmware.h"

// Returns a human readable string describing the update result.
String getUpdateMessage();

// Starts an OTA firmware update by connecting to the specified Wi-Fi network and downloading the firmware from the URL.
// The firmware's integrity is verified with the provided MD5 checksum.
void startUpdate(String wifi, String psk, String url, size_t length, String hash);