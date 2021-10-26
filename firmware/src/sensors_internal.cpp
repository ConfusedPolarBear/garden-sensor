#include <firmware.h>

const String tag = "i2c";

bool readRegister(int address, int reg, int count, byte* data) {
    String strAddr = String(address, HEX);

    LOGD(tag, "requesting " + String(count) + " bytes from 0x" + strAddr + " (register " + String(reg, HEX) + ")");

    // Get the attention of the device
    Wire.beginTransmission(address);

    // Move the device's pointer to the specified register
    Wire.write(reg);

    // End the transmission
    int err = Wire.endTransmission();
    if (err != 0) {
        LOGW(tag, "unable to get reading from " + strAddr + ". error code " + String(err));
        return false;
    }

    // Read the response
    int returned = Wire.requestFrom(address, count);

    // If the device returned an incorrect number of bytes, fail
    if (returned != count) {
        LOGW(tag, "invalid response received from " + strAddr + ". expected " + String(count) + ", actual " + String(returned));
        return false;
    }

    // Store the data
    for (int i = 0; i < count; i++) {
        data[i] = (byte)Wire.read();
        LOGD(tag, "byte " + String(i+1) + "/" + String(count) + ": " + String(data[i], HEX));
    }

    return true;
}

bool readRegister8(int address, int reg, uint8_t* data) {
    byte buf[1];

    if (!readRegister(address, reg, 1, buf)) {
        return false;
    }

    *data = buf[0];
    return true;
}

bool readRegister16(int address, int reg, uint16_t* data) {
    byte buf[2];

    if (!readRegister(address, reg, 2, buf)) {
        return false;
    }

    *data = (buf[0] << 8) | buf[1];
    return true;
}
