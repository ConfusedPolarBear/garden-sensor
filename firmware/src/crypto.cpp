#include "firmware.h"

#ifdef ESP32
#include "mbedtls/md.h"
#else
#include "Crypto.h"
#endif

String _hmacKey = "";

void loadKey() {
    _hmacKey = ReadFile(FILE_MESH_KEY);

    if (_hmacKey == "") {
        LOGF("crypto", "mesh key is not set");
    }
}

String arrayToString(const uint8_t* raw, size_t length) {
    String result;

    for (unsigned int i = 0; i < length; i++) {
        // bytes 0 and 1 of the hex encoded byte + terminating null
        char tmp[3];
        snprintf(tmp, 3, "%02X", raw[i]);

        result += tmp;
    }

    return result;
}

#ifdef ESP32
uint8_t* hmac(const uint8_t* data, size_t dataLength) {
    if (_hmacKey == "") { loadKey(); }

    void* result = malloc(32);
    if (result == NULL) {
        LOGF("crypto", "unable to allocate memory");
    }

    memzero(result, 32);

    mbedtls_md_context_t ctx;
    mbedtls_md_type_t md_type = MBEDTLS_MD_SHA256;

    // Initialize the mbed TLS context and set it up for HMAC mode.
    mbedtls_md_init(&ctx);

    if (mbedtls_md_setup(&ctx, mbedtls_md_info_from_type(md_type), 1) != 0) {
        LOGF("crypto", "setup failed");
    }

    // Setup the HMAC key.
    if (mbedtls_md_hmac_starts(&ctx, (const unsigned char*)_hmacKey.c_str(), _hmacKey.length()) != 0) {
        LOGF("crypto", "key write failed");
    }

    // Write the data to the started HMAC.
    if (mbedtls_md_hmac_update(&ctx, data, dataLength) != 0) {
        LOGF("crypto", "payload write failed");
    }

    // Finish and free the HMAC.
    if (mbedtls_md_hmac_finish(&ctx, (uint8_t*)result) != 0) {
        LOGF("crypto", "finalization failed");
    }

    mbedtls_md_free(&ctx);

    return (uint8_t*)result;
}
#else
uint8_t* hmac(const uint8_t* data, const size_t dataLength) {
    using namespace experimental::crypto;

    if (_hmacKey == "") { loadKey(); }

    void* result = malloc(32);
    if (result == NULL) {
        LOGF("crypto", "unable to allocate memory");
    }

    memzero(result, 32);

    SHA256::hmac(data, dataLength, _hmacKey.c_str(), _hmacKey.length(), result, 32);

    return (uint8_t*)result;
}
#endif

bool hmacCompare(const uint8_t* lhs, const uint8_t* rhs) {
    int result = 0;
    for (unsigned int i = 0; i < 32; i++) {
        result |= (lhs[i] ^ rhs[i]);
    }

    return result == 0;
}
