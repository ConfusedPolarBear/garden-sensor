#include <firmware.h>

const String tag = "i2c";
sensorData latest;

// Vector of sensors attached to this system.
std::vector<sensor> sensors;

void initializeSensors() {
    sensors.clear();

    LOGD(tag, "scanning for devices");

    // I2C addresses range from 1 to 127 (0x7F). Address 0 is general call.
    for (int addr = 1; addr <= 127; addr++) {
        Wire.beginTransmission(addr);
        uint8_t error = Wire.endTransmission();

        if (error != 0) {
            continue;
        }

        LOGD(tag, "detected device at 0x" + String(addr, HEX));

        sensors.push_back(getSensorType(addr));
    }

    LOGD(tag, "device scan completed. " + String(sensors.size()) + " devices found.");
}

sensor getSensorType(int address) {
    sensor s;

    s.name = "unknown";
    s.address = address;
    s.populate = unknown;

    s.hexAddress = "0x" + String(address, HEX);
    if (s.hexAddress.length() == 3) {       // as in "0x1" through "0xf"
        s.hexAddress = "0" + s.hexAddress;
    }

    uint8_t shortResult;
    uint16_t result;

    switch (address) {
        case 0x18:
            // Default address for the MCP9808.

            // Validate the manufacturer ID
            if (!readRegister16(address, 0x06, &result) || result != 0x0054) {
                LOGW(tag, "device at " + s.hexAddress + " is not a MCP9808 (invalid manufacturer)");
                break;
            }
            LOGD(tag, "MCP9808: validated manufacturer ID");

            // Validate the device ID
            if (!readRegister16(address, 0x07, &result) || result != 0x0400) {
                LOGW(tag, "device at " + s.hexAddress + " is not a MCP9808 (invalid device)");
                break;
            }
            LOGD(tag, "MCP9808: validated device ID");

            s.name = "MCP9808";
            s.populate = mcp9808;

            break;
    }

    LOGD(tag, "detected device " + s.hexAddress + " as " + s.name);

    return s;
}

sensorData getSensorData() {
    latest.error = false;
    latest.temperature = INVALID_DATA;
    latest.humidity = INVALID_DATA;

    for (sensor s : sensors) {
        s.populate(s);
    }

    if (latest.temperature == INVALID_DATA || latest.humidity == INVALID_DATA) {
        latest.error = true;
    }

    return latest;
}

void mcp9808(sensor s) {
    // Datasheet is at https://ww1.microchip.com/downloads/en/DeviceDoc/25095A.pdf

    /* Registers:
     *   0x00: Reserved
     *   0x01: Configuration
     *   0x02: Alert upper boundary
     *   0x03: Alert lower boundary
     *   0x04: Critical temperature
     *   0x05: Ambient temperature
     *   0x06: Manufacturer ID
     *   0x07: Device ID
     *   0x08: Resolution
     *   0x09: Reserved (same for all higher registers)
    */

   /* The temperature is stored in register 0x05. It is a 16 bit word.
    * Bits 0 through 12 is the temperature.
    * Bit 13 is the sign.
    * Bits 14 through 16 are result of comparing the temperature with the various alert points
   */

  uint16_t data;
  if (!readRegister16(s.address, 0x05, &data)) {
      latest.error = true;
      return;
  }

  latest.temperature = (data & 0x0FFF) / 16.0;
  if (data & 0x1000) {
      latest.temperature *= -1;
  }

  LOGD(tag, "MCP9808: raw temperature is " + String(data, HEX) + ", converted to " + String(latest.temperature));
}

void unknown(sensor s) {
    if (latest.error) {
        return;
    }

    LOGW(tag, "sensor at " + s.hexAddress + " is unimplemented");
}
