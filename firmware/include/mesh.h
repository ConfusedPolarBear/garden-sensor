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

// Internal callback handler function when a mesh message has been sent.
void meshSendCallbackHandler(const uint8_t* mac, bool success);

// Internal callback handler function when a mesh message has been received.
void meshReceiveCallbackHandler(const uint8_t* mac, const uint8_t* buf, int length);

// Broadcast a message to all paired nodes except one.
void broadcastMesh(uint8_t* data, String exclude = "");
