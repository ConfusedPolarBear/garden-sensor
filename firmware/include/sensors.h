#pragma once

#include <DHT.h>

struct sensorData {
    bool  error;
    float temperature;
    float humidity;
};

// Initialize all sensors or reboot.
void initializeSensors();

// Returns the latest sensor data.
sensorData getSensorData();
