#pragma once

#include <Arduino.h>
#include <Streaming.h>

#ifdef DEBUG
#define LOGD(tag, msg) logAdvanced("D", tag, msg)
#else
#define LOGD(tag, msg)
#endif

void logAdvanced(String level, String tag, String message);

// Logs an informational message.
void LOGI(String tag, String message);

// Log a warning.
void LOGW(String tag, String message);

// Log a fatal message to the Serial port and restart.
void LOGF(String tag, String message);
