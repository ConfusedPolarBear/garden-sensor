#pragma once

#include <Arduino.h>

struct meshStatistics {
    // Total messages sent by this system.
    int sent;

    // Total messages received.
    int received;

    // Total messages dropped due to being short (< 250 bytes).
    int droppedLength;

    // Total messages dropped due to a bad HMAC.
    int droppedAuth;

    // Total messages accepted for processing.
    int accepted;
};

int loadPeers();
bool isKnownPeer(String needle);
bool publishMesh(String data, String topic);

// Sends a preformatted 250 byte message.
void publishMeshRaw(uint8_t* address, uint8_t* data);

// Initializes ESP-NOW or restart.
void initializeMesh(bool isController, int channel);

bool parseMac(String mac, uint8_t dst[6]);

// Adds ESP-NOW peer and returns result.
bool addMeshPeer(String mac);

// Internal callback handler function when a mesh message has been sent.
void meshSendCallbackHandler(const uint8_t* mac, bool success);

// Internal callback handler function when a mesh message has been received.
void meshReceiveCallbackHandler(const uint8_t* mac, const uint8_t* buf, int length);

// Broadcast a message to all paired nodes except one.
void broadcastMesh(uint8_t* data, String exclude = "");

// Return the latest statistics about the local mesh performance.
meshStatistics getStatistics();
