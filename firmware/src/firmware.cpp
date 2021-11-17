#include <firmware.h>

String wifiSsid, wifiPass;
String mqttHost, mqttUser, mqttPass;

bool isController;
String meshController;
int meshChannel;

void setup() {
    Wire.begin(4, 5);  // data, clock

    Serial.begin(115200);
    Serial.setTimeout(500);     // timeout for readStringUntil()

    Serial << endl << endl;

    Mount();

    Serial << "Mesh MAC address: " << WiFi.softAPmacAddress() << endl;

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

            processCommand(Serial.readStringUntil('\n'));
        }
    }

    // Configuration is valid, load and display it for validation
    wifiSsid = ReadFile(FILE_WIFI_SSID);
    wifiPass = ReadFile(FILE_WIFI_PASS);
    mqttHost = ReadFile(FILE_MQTT_HOST);
    mqttUser = ReadFile(FILE_MQTT_USER);
    mqttPass = ReadFile(FILE_MQTT_PASS);

    meshController = ReadFile(FILE_MESH_CONTROLLER);

    String rawChannel = ReadFile(FILE_MESH_CHANNEL);
    if (rawChannel.length() > 0) {
        meshChannel = rawChannel.toInt();

        if (meshChannel <= 0) {
            LOGW("mesh", "invalid mesh channel specified, defaulting to 1");
            meshChannel = 1;
        } else {
            LOGD("mesh", "using mesh channel " + String(meshChannel));
        }
    } else {
        LOGD("mesh", "using default channel of 1");
        meshChannel = 1;
    }

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

    // If this is a controller, connect to dedicated Wi-Fi network. If this is a client, just setup an access point.
    WiFi.mode(WIFI_AP_STA);	// clients are put into ap_sta mode so they can join a network if needed for updates.
    if (WiFi.setSleep(false)) {
        LOGD("wifi", "wifi sleep disabled");
    } else {
        LOGW("wifi", "failed to disable sleep");
    }

    if (isController) {
        // Attempt to connect to the provided Wi-Fi network
        WiFi.begin(wifiSsid.c_str(), wifiPass.c_str());

        Serial << "Connecting to Wi-Fi";
        while (WiFi.status() != WL_CONNECTED)
        {
            Serial << ".";
            processCommand(Serial.readStringUntil('\n'));
        }
        Serial << endl;

        Serial << "IP address: " << WiFi.localIP() << endl;

        connectToBroker(mqttHost, mqttUser, mqttPass);

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

bool sentTest = false;
long lastPublish = 0;
void loop() {
    // Process any Serial commands
    processCommand(Serial.readStringUntil('\n'));

    if (isController) {
        connectToBroker(mqttHost, mqttUser, mqttPass);      // will only reconnect if needed
        processMqtt();
    }

    // If a Wi-Fi scan was requested, process the results
    processNetworkScan();

    // Publish sensor data
    // Note that delay() *cannot* be used here (or anywhere else in the loop function) because if a delay is active
    //    when an ESP-NOW message arrives, the message won't be processed by the system.
    if (millis() - lastPublish >= 60 * 1000) {
        String strReading;

        StaticJsonDocument<100> mesh;
        meshStatistics stats = getStatistics();

        mesh["SE"] = stats.sent;
        mesh["RC"] = stats.received;
        mesh["DL"] = stats.droppedLength;
        mesh["DA"] = stats.droppedAuth;
        mesh["AC"] = stats.accepted;
        serializeJson(mesh, strReading);
        publish(strReading, "mesh");
        
        strReading = "";

        lastPublish = millis();

        if (isController) {
            return;
        }

        sensorData reading = getSensorData();

        StaticJsonDocument<100> json;
        json["Error"] = reading.error;
        json["Temperature"] = reading.temperature;
        json["Humidity"] = reading.humidity;

        serializeJson(json, strReading);

        publish(strReading);
    }
}

void processCommand(String command) {
    bool changed = false;

    if (command.length() == 0) {
        return;
    }

    command.replace("\r", "");

    // Try to deserialize the input as JSON
    LOGD("cmnd", "deserializing '" + command + "'");
    DynamicJsonDocument data(3 * 1024);
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
}
