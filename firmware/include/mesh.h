#pragma once

#include <Arduino.h>

int loadPeers();
bool isKnownPeer(String needle);
bool publishMesh(String data, String topic);

// Sends a preformatted 250 byte message.
void publishMeshRaw(uint8_t* address, uint8_t* data);

// Initializes ESP-NOW or restart.
void initializeMesh(bool isController, int channel);

// Adds ESP-NOW peer and returns result.
bool addMeshPeer(String mac);

// Callback when a mesh message has been sent.
void meshSendCallback(uint8_t* mac, uint8_t status);

// Callback when a mesh message has been received.
void meshReceiveCallback(uint8_t* mac, uint8_t* buf, uint8_t len);

// Broadcast a message to all paired nodes except one.
void broadcastMesh(uint8_t* data, String exclude = "");
