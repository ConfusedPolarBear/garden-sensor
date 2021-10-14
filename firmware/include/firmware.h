#pragma once

#ifdef ESP32
#include <WiFi.h>
#include <esp_now.h>        // why is the name different between the two boards
#else
#include <ESP8266WiFi.h>
#include <espnow.h>         // this is dumb
#endif

#include <vector>

#include <Arduino.h>
#include <ArduinoJson.h>
#include <filesystem.h>
#include <logger.h>
#include <MQTT.h>
#include <PubSubClient.h>
#include <Streaming.h>

// ========== Paths to configuration files ==========
#define FILE_WIFI_SSID "/wifiSSID"
#define FILE_WIFI_PASS "/wifiPass"

#define FILE_MQTT_HOST "/mqttHost"
#define FILE_MQTT_USER "/mqttUser"
#define FILE_MQTT_PASS "/mqttPass"

#define FILE_MESH_CONTROLLER "/meshController"
#define FILE_MESH_PEERS      "/meshPeers"
#define FILE_MESH_CHANNEL    "/meshChannel"

// ========== Command handling ==========
// Checks if the Serial connection has a command. If it does, handle it.
void parseSerial();

// Process a sent command.
void processCommand(String command);

// ========== Mesh ==========
// Initializes ESP-NOW or restart.
void initializeMesh();

// Adds ESP-NOW peer and returns result.
bool addMeshPeer(String mac);

// Callback when a mesh message has been sent.
void meshSendCallback(uint8_t* mac, uint8_t status);

// Callback when a mesh message has been received.
void meshReceiveCallback(uint8_t* mac, uint8_t* buf, uint8_t len);

// ========== MQTT ==========
// Connect to the MQTT broker or restart.
void setupMQTT();

// Callback when a MQTT message has been received.
void mqttReceiveCallback(const MQTT::Publish& pub);

// Publish a message over ESP-NOW or MQTT.
bool publish(String data, bool isReading);

// ========== Networking ==========
void startAccessPoint();
void stopAccessPoint();
void startNetworkScan();
void processNetworkScan();

// ========== Miscellaneous ==========
// Returns a secure randomly generated string.
String secureRandom();