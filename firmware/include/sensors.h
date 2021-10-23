#pragma once

#define INVALID_DATA 32768

struct sensor {
    String name;

    uint8_t address;
    String hexAddress;

    std::function<void(sensor)> populate;
};

typedef std::function<void(sensor)> sensorFunc;

struct sensorData {
    bool  error;
    float temperature;
    float humidity;
};

// Initialize all sensors or reboot.
void initializeSensors();

// Reads the provided number of bytes from the specified address and register. Returns true on success.
bool readRegister(int address, int reg, int count, byte* data);

// Reads the specified register and returns a 8 bit result.
bool readRegister8(int address, int reg, uint8_t* data);

// Reads the specified register and returns a 16 bit result.
bool readRegister16(int address, int reg, uint16_t* data);

// Returns the latest sensor data.
sensorData getSensorData();

// Determines the type of a sensor given an address.
sensor getSensorType(int address);

// Unknown generic sensor (for testing).
void unknown(sensor s);

// The MCP9808 is a high precision temperature sensor.
void mcp9808(sensor s);
