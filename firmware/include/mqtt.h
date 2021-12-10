#pragma once

#include <Arduino.h>
#include <PubSubClient.h>

// Connect to the MQTT broker or restart.
void connectToBroker();

// Callback when a MQTT message has been received.
void mqttReceiveCallback(const MQTT::Publish& pub);

// Process MQTT messages.
void processMqtt();

// Publish a message over ESP-NOW or MQTT.
bool publish(String data, String teleTopic = "data");

// Publish a JSON document over ESP-NOW or MQTT.
bool publish(JsonDocument* doc, String teleTopic);

// Returns the MQTT client identifier
String getClientId();