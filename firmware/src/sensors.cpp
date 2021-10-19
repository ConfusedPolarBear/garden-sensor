#include <firmware.h>

#warning TODO: make modular

// Sensors
DHT dht(4, DHT22);

const String tag = "sensor";
sensorData latest;

void initializeSensors() {
    latest.temperature = 0;
    latest.humidity = 0;

    LOGD(tag, "initializing DHT22");

    dht.begin();

    LOGD(tag, "initialization completed successfully");
}

sensorData getSensorData() {
    latest.error = true;
    
    LOGD(tag, "reading data from DHT22");

    float tmp = dht.readTemperature();
    float hum = dht.readHumidity();

    if (isnan(tmp) || isnan(hum)) {
        LOGW(tag, "invalid data received from DHT22");
        return latest;
    }

    LOGD(tag, "data successfully read");

    latest.temperature = tmp;
    latest.humidity = hum;
    latest.error = false;

    return latest;
}
