#include "firmware.h"
#include "ChaChaPoly.h"
#include "SHA256.h"

String _hmacKey = "";
uint8_t* _symKey;

void loadKey() {
    // The HMAC key is the raw HMAC file contents.
    _hmacKey = ReadFile(FILE_MESH_KEY);

    if (_hmacKey == "") {
        LOGF("crypto", "mesh key is not set");
    }

    // Derive the symmetric encryption key from the raw HMAC one since separating cryptographic keys is a good thing to do.
    String symData = "chacha-symmetric-key";
    _symKey = hmac((const uint8_t*)symData.c_str(), symData.length());
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

uint8_t* hmac(const uint8_t* data, size_t dataLength) {
    if (_hmacKey == "") { loadKey(); }

    void* result = calloc(32, sizeof(uint8_t));
    if (result == NULL) {
        LOGF("crypto", "unable to allocate memory");
    }

    SHA256* hmac = new SHA256();
    hmac->resetHMAC(_hmacKey.c_str(), _hmacKey.length());
    hmac->update(data, dataLength);
    hmac->finalizeHMAC(_hmacKey.c_str(), _hmacKey.length(), result, 32);

    return (uint8_t*)result;
}

bool hmacCompare(const uint8_t* lhs, const uint8_t* rhs) {
    int result = 0;
    for (unsigned int i = 0; i < 32; i++) {
        result |= (lhs[i] ^ rhs[i]);
    }

    return result == 0;
}

bool decrypt(const char* cipher, const size_t cipherLength, const char* nonce, const char* tag, char* output) {
    if (_hmacKey == "") { loadKey(); }

    ChaChaPoly* chacha = new ChaChaPoly();

    if (!chacha->setKey(_symKey, 32)) {
        LOGF("crypto", "symmetric key setup failed");
    }

    // LOGD("crypto", "setting \"iv\" to " + arrayToString((const uint8_t*)nonce, 12));
    if (!chacha->setIV((const uint8_t*)nonce, 12)) {
        LOGF("crypto", "symmetric nonce setup failed");
    }

    // LOGD("crypto", "decrypting ciphertext " + arrayToString((const uint8_t*)cipher, cipherLength));
    chacha->decrypt((uint8_t*)output, (const uint8_t*)cipher, cipherLength);

    // LOGD("crypto", "tag is " + arrayToString((const uint8_t*)tag, 16));
    bool okay = chacha->checkTag(tag, 16);
    if (!okay) {
        LOGW("crypto", "message authentication failed");

        // Zeroize the output if the tag is wrong because ciphertext that fails authentication should never ever be released
        // to callers ever but the library does so anyway.
        memzero(output, cipherLength);
    }

    chacha->clear();
    delete chacha;

    return okay;
}
