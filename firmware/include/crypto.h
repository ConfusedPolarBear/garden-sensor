#pragma once

#include "Arduino.h"

// Converts the uint8_t array to a hex string.
String arrayToString(const uint8_t* raw, size_t length);

// Returns the SHA256 HMAC of the provided data. The caller is responsible for free()'ing the returned pointer.
uint8_t* hmac(const uint8_t* data, size_t dataLength);

// Compare two HMACs in constant time.
bool hmacCompare(const uint8_t* lhs, const uint8_t* rhs);

// Decrypts data encrypted with ChaCha20-Poly1305. Returns true when ciphertext is authentic & decrypts successfully.
bool decrypt(const char* cipher, const size_t cipherLength, const char* nonce, const char* tag, char* output);
