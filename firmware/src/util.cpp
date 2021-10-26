#include <firmware.h>

uint32_t secureRandom() {
    #ifdef ESP32
    return esp_random();
    #else
    return ESP.random();
    #endif
}

// Returns a 192-bit secure random number
String secureRandomNonce() {
    String rand;
    for (int i = 0; i < 6; i++) {
        rand += String(secureRandom(), HEX);
    }

    return rand;
}

void memzero(void* ptr, size_t size) {
    memset(ptr, 0, size);
}
