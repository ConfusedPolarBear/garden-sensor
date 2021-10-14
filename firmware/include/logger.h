#pragma once

#include <Arduino.h>
#include <Streaming.h>

// Log a message to the Serial port at debug level.
void LOGD(String tag, String message);

// Log a fatal message to the Serial port and restart.
void LOGF(String tag, String message);
