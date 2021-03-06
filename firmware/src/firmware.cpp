#include <firmware.h>

bool isController;
String meshController;
int meshChannel;

std::queue<String> commandQueue;

uint32_t lastStack = 3 * 1024;

void setup() {
    Wire.begin(4, 5);  // data, clock

    Serial.begin(115200);
    Serial.setTimeout(500);     // timeout for readStringUntil()

    #ifdef USE_BUILTIN_LED
    pinMode(BUILTIN_LED, OUTPUT);     // wemos d1 minis have the builtin LED on D4 (GPIO2) *active low*.
    digitalWrite(BUILTIN_LED, HIGH);
    #endif

    Serial << endl << endl;

    Mount();

    WiFi.persistent(false);
    WiFi.mode(WIFI_AP_STA);
    Serial << "Mesh MAC address: " << getIdentifier(true) << endl;

    // Check if the system has been configured
    bool configured = false;
    Serial << "Configuration done: ";
    if (IsConfigured()) {
        configured = true;
        Serial << "yes" << endl;
    } else {
        Serial << "no" << endl;
    }

    Serial << "Press 's' to start setup" << endl;
    delay(500);

    // If the user wants to enter setup or if the system has not been configured, wait for configuration data
    if (Serial.read() == 's' || !configured) {
        Serial << "Waiting for commands..." << endl;

        while (true) {
            if (!Serial.available()) {
                continue;
            }

            processCommand(Serial.readStringUntil('\n'), true);
        }
    }

    // Configuration is valid, load and display it for validation
    String wifiSsid = ReadFile(FILE_WIFI_SSID);
    String mqttHost = ReadFile(FILE_MQTT_HOST);
    String mqttUser = ReadFile(FILE_MQTT_USER);

    meshController = ReadFile(FILE_MESH_CONTROLLER);
    meshChannel = getMeshChannel();

    Serial << "Settings:" << endl;
    if (meshController.length() == 0) {
        isController = true;
        Serial << "\tMode:          controller" << endl;
        Serial << "\tWi-Fi SSID:    " << wifiSsid << endl;
        Serial << "\tMQTT broker:   " << mqttHost << endl;
        Serial << "\tMQTT username: " << mqttUser << endl;
    } else {
        isController = false;
        Serial << "\tMode:          client" << endl;
        Serial << "\tController:    " << meshController << endl;
    }

    #ifdef ESP8266
    // Without these four lines of code, the Wi-Fi radio won't come out of deep sleep correctly.
    // This is a bug in the ESP8266 SDK.
    WiFi.forceSleepBegin();
    delay(1);
    WiFi.forceSleepWake();
    delay(1);
    #endif

    if (WiFi.setSleep(false)) {
        LOGD("wifi", "wifi sleep disabled");
    } else {
        LOGW("wifi", "failed to disable sleep");
    }

    if (isController) {
        connectToWifi();

        // Since ESP-NOW is very unreliable if the channels don't match, force them to match.
        meshChannel = WiFi.channel();
        LOGD("mesh", "forcing channel match with channel " + String(meshChannel));
    }

    startAccessPoint(meshChannel);

    /* Mesh details:
     * All nodes default to "client" mode (i.e. not the controller)
     * All nodes:
     *   Have an internal list of all known peers
     *   Scan nearby Wi-Fi network names and filter those down to the known peers
     *   Set the 6 closest ones as mesh peers
     * Controller:
     *   Will either add a node as a peer or instruct another node to do that
     * Clients:
     *   Wait to be contacted by a nearby node
    */

    initializeMesh(isController, meshChannel);
    if (!isController) {
        addMeshPeer(meshController);
    }

    loadPeers();

    initializeSensors();
    sendDiscoveryMessage(isController);
}

void printMemoryStatistics(String location) {
    #ifdef ESP8266
    uint32_t current = ESP.getFreeContStack();

    LOGD("app", "memory stats at " + location);
    LOGD("app", "free heap " + String(ESP.getFreeHeap()));

    uint32_t decrease = lastStack - current;
    String msg = "stack free " + String(current);
    if (decrease > 0) {
        msg.concat(" (decreased by " + String(decrease) + ")");
    }

    if (current <= 512) {
        LOGW("app", msg);
    } else {
        LOGD("app", msg);
    }

    lastStack = current;
    #endif
}

long lastPublish = 0;
void loop() {
    // Process outstanding commands
    processCommand(Serial.readStringUntil('\n'), true);

    if (!commandQueue.empty()) {
        processCommand(commandQueue.front());
        commandQueue.pop();
    }

    if (isController) {
        connectToBroker();      // will only reconnect if needed
        processMqtt();
    }

    // If a Wi-Fi scan was requested, process the results
    processNetworkScan();

    // Publish sensor data
    // Note that delay() *cannot* be used here (or anywhere else in the loop function) because if a delay is active
    //    when an ESP-NOW message arrives, the message won't be processed by the system.
    if (millis() - lastPublish >= 60 * 1000) {
        DynamicJsonDocument mesh(100);;
        meshStatistics stats = getStatistics();

        mesh["SE"] = stats.sent;
        mesh["RC"] = stats.received;
        mesh["DL"] = stats.droppedLength;
        mesh["DA"] = stats.droppedAuth;
        mesh["AC"] = stats.accepted;
        publish(mesh, "mesh");

        lastPublish = millis();

        if (isController) {
            return;
        }

        sensorData reading = getSensorData();

        DynamicJsonDocument json(100);
        json["Error"] = reading.error;
        json["Temperature"] = reading.temperature;
        json["Humidity"] = reading.humidity;
        publish(json, "data");
    }
}

void queueCommand(String command) {
    if (command.startsWith("e")) {
        LOGD("app", "queued encrypted command");
    } else {
        LOGD("app", "queued command " + command);
    }

    commandQueue.push(command);
}

void processCommand(String command, bool secure) {
    bool changed = false;

    if (command.length() == 0) {
        return;
    }

    // printMemoryStatistics("top of processCommand");

    command.replace("\r", "");

    // Encrypted commands start with "e" & need to be decrypted before processing. Format: e || NONCE || TAG || CIPHERTEXT
    // where "||" denotes concatenation.
    
    // Check that the command is at least 31 bytes since the encryption has a fixed overhead of 30 bytes.
    if (command.charAt(0) == 'e' && command.length() > 30) {
        LOGD("cmnd", "command is encrypted");

        size_t len = command.length();
        if (len - 12 - 16 >= 250) {
            LOGW("cmnd", "ciphertext too long");
            return;
        }

        // Nonce is 12 bytes & auth tag is 16 bytes.
        String nonce = command.substring(1, 13);
        String tag = command.substring(13, 29);
        command = command.substring(29, len);
        len = command.length();

        void* plaintext = calloc(len, sizeof(char));

        if (!decrypt(command.c_str(), len, nonce.c_str(), tag.c_str(), (char*)plaintext)) {
            free(plaintext);
            return;
        }

        LOGD("crypto", "successfully decrypted command");

        command = "";
        for (size_t i = 0; i < len; i++) {
            command += ((char*)plaintext)[i];
        }

        free(plaintext);

        secure = true;
    }

    // Ignore MAC addresses that are prefixed to the JSON.
    if (command.indexOf("dst-") == 0) {
        command = command.substring(16);
    }

    // Try to deserialize the input as JSON
    LOGD("cmnd", "deserializing '" + command + "'");
    DynamicJsonDocument data(1024);
    DeserializationError error = deserializeJson(data, command);

    // Log success or failure
    if (error) {
        LOGW("cmnd", String("deserialization failed: ") + error.c_str());
        return;
    } else {
        LOGD("cmnd", "deserialization successful");
    }

    // Check if a command was sent
    if (data.containsKey("Command")) {
        String command = data["Command"];
        command.toLowerCase();

        if (command == "scan") {
            startNetworkScan();
        }
        
        else if (command == "restart") {
            Unmount();
            ESP.restart();
        }

        else if (command == "reset") {
            Format();
        }

        else if (command == "publish") {
            String payload = data["Payload"];
            if (payload.length() == 0) {
                LOGW("app", "the payload property is required");
                return;
            }

            // If the payload has the prefix "h", it is hex encoded. Decode it before sending it.
            if (payload.charAt(0) == 'h') {
                LOGD("app", "decoding hex mesh message before transmitting");
                payload += "00";

                size_t len = payload.length();
                String newPayload = "";
                for (size_t i = 1; i < len; i += 2) {
                    String current = payload.substring(i, i+2);
                    char chr = (char)(int)strtol(current.c_str(), NULL, 16);
                    newPayload += chr;
                }

                payload = newPayload;
            }

            publishMesh(payload, "");

            // If this is a broadcast message, handle it after rebroadcasting it.
            if (payload.indexOf("dst-FFFFFFFFFFFF") != -1) {
                processCommand(payload);
            }
        }

        else if (command == "listpeers") {
            publish(ReadFile(FILE_MESH_PEERS), "peers");
        }

        else if (command == "ping") {
            publish("pong", "ping");
        }

        else if (command == "sleep") {
            // Deep sleep for X seconds
            int period = data["Period"];

            // Ensure that if a sleep period <= 0 is passed, it is changed to 1. Passing 0 to a deep sleep function will
            // result in an infinite sleep.
            period = max(period, 1);

            if (isController && !data["IncludeController"]) {
                LOGW("sleep", "controller received sleep message but IncludeController was not set - ignoring");
                return;
            }

            LOGD("sleep", "entering deep sleep for " + String(period) + " seconds");

            // Immediately sleep the chip for `period` seconds. Does not gracefully terminate Wi-Fi connections.
            #ifdef ESP32
            esp_deep_sleep(period * 1e6);
            #else
            ESP.deepSleep(period * 1e6, RF_DISABLED);
            #endif
        }

        else if (command == "update") {
            DynamicJsonDocument updateResult(200);
            updateResult["Success"] = false;

            if (!secure) {
                updateResult["Message"] = "update command sent insecurely";
                publish(updateResult, "ota");
                return;
            }

            // SSID & PSK are optional but break startUpdate() if they're null, so set to empty strings if not provided.
            String ssid = data["S"] | "";
            String psk = data["P"] | "";

            if (!data.containsKey("U") || !data.containsKey("L") || !data.containsKey("C")) {
                updateResult["Message"] = "url, length, and hash are required";
                publish(updateResult, "ota");
                return;
            }

            String url = data["U"];
            String rawLength = data["L"];
            String checksum = data["C"];

            const size_t length = rawLength.toInt();
            if (length <= 32 * 1024) {
                updateResult["Message"] = "invalid new size for firmware binary";
                publish(updateResult, "ota");
                return;
            }

            // Tell the server we're going to start attempting an update. Use safeDelay() to ensure the message is sent
            // before the Wi-Fi connection is (probably) changed.
            updateResult["Success"] = true;
            updateResult["Message"] = "attempting to download update from " + url + " using network " + ssid;
            publish(updateResult, "ota");
            safeDelay(500);

            startUpdate(ssid, psk, url, length, checksum);

            String m = getUpdateMessage();
            LOGD("ota", "publishing failure with reason " + m);
            updateResult["Message"] = m;

            if (isController) {
                connectToWifi();
            }

            safeDelay(500);

            updateResult["Success"] = false;
            publish(updateResult, "ota");

            safeDelay(500);

            LOGF("ota", "restarting to recover from failed update");
        }

        else {
            LOGW("cmnd", "unknown command");
        }
    }

    #warning TODO: publish success to serial and network (MQTT or ESP-NOW).

    // No command was sent, check if the Wi-Fi settings need to be updated.
    if (data.containsKey("WifiSSID")) {
        WriteFile(FILE_WIFI_SSID, data["WifiSSID"]);
        WriteFile(FILE_WIFI_PASS, data["WifiPassword"]);

        /*
        result["Success"] = true;
        result["ConfiguredWifi"] = true;
        */

        LOGD("cmnd", "updated wifi settings");
        changed = true;
    }

    if (data.containsKey("MQTTHost")) {
        WriteFile(FILE_MQTT_HOST, data["MQTTHost"]);
        LOGD("cmnd", "updated mqtt settings");

        if (data.containsKey("MQTTUsername")) {
            WriteFile(FILE_MQTT_USER, data["MQTTUsername"]);
            WriteFile(FILE_MQTT_PASS, data["MQTTPassword"]);
            LOGD("cmnd", "mqtt is authenticated");
        }

        /*
        result["Success"] = true;
        result["ConfiguredMQTT"] = true;
        result["MQTTAuthenticated"] = fs->FileExists(FILE_MQTT_USER);
        */
       changed = true;
    }

    if (data.containsKey("MeshController")) {
        String mac = data["MeshController"];
        uint8_t discard[6];
        if (parseMac(mac, discard)) {
            WriteFile(FILE_MESH_CONTROLLER, mac);
            changed = true; 
        }
    }

    if (data.containsKey("MeshChannel")) {
        WriteFile(FILE_MESH_CHANNEL, data["MeshChannel"]);
        changed = true;
    }

    if (data.containsKey("MeshKey")) {
        WriteFile(FILE_MESH_KEY, data["MeshKey"]);
        changed = true;
    }

    if (data.containsKey("MeshPeers")) {
        String peers = data["MeshPeers"];
        String current = ReadFile(FILE_MESH_PEERS);

        if (peers.length() == 0) {
            LOGD("mesh", "blanking known peers");
            current = "";
        } else {
            LOGD("mesh", "appending to peer list");
            current += peers;

            if (!current.endsWith(",")) {
                current += ",";
            }
        }

        WriteFile(FILE_MESH_PEERS, current);

        loadPeers();

        changed = true;
    }

    if (!changed) {
        return;
    }

    if (FileExists(FILE_WIFI_SSID) && FileExists(FILE_MQTT_HOST)) {
        LOGD("cmnd", "wifi and mqtt are configured, setting flag");
        SetConfigured(true);

    } else if ((FileExists(FILE_MESH_CONTROLLER) || FileExists(FILE_MESH_PEERS)) && FileExists(FILE_MESH_KEY)) {
        LOGD("cmnd", "mesh is configured, setting flag");
        SetConfigured(true);

    } else {
        LOGD("cmnd", "not configured");
    }

    // printMemoryStatistics("returning from processCommand");
}

void safeDelay(const size_t time) {
    const size_t start = millis();
    while(millis() - start <= time) {
        processMqtt(false);
        yield();
    }
}

void flashLed() {
    #ifndef USE_BUILTIN_LED
    return;
    #endif

    digitalWrite(BUILTIN_LED, LOW);
    delay(50);
    digitalWrite(BUILTIN_LED, HIGH);
}
