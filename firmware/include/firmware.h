#pragma once

#ifdef ESP32
#include <WiFi.h>
#include <esp_now.h>        // why is the name different between the two boards
#else
#include <ESP8266WiFi.h>
#include <espnow.h>         // this is dumb
#endif

#include <algorithm>
#include <cmath>
#include <vector>

#include <Arduino.h>
#include <ArduinoJson.h>
#include <MQTT.h>
#include <Streaming.h>
#include <Wire.h>

#include <crypto.h>
#include <filesystem.h>
#include <logger.h>
#include <mesh.h>
#include <mqtt.h>
#include <networking.h>
#include <sensors.h>

// ========== Paths to configuration files ==========
#define FILE_WIFI_SSID "/wifiSSID"
#define FILE_WIFI_PASS "/wifiPass"

#define FILE_MQTT_HOST "/mqttHost"
#define FILE_MQTT_USER "/mqttUser"
#define FILE_MQTT_PASS "/mqttPass"

#define FILE_MESH_CONTROLLER "/meshController"
#define FILE_MESH_PEERS      "/meshPeers"
#define FILE_MESH_CHANNEL    "/meshChannel"
#define FILE_MESH_KEY        "/meshKey"

// ========== Command handling ==========
// Checks if the Serial connection has a command. If it does, handle it.
void parseSerial();

// Process a sent command.
void processCommand(String command);

// ========== Utility functions ==========
uint32_t secureRandom();
String secureRandomNonce();
void memzero(void* ptr, size_t size);
