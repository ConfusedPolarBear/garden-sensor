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
    delay(2000);

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
            LOGD("mesh", "using mesh channel " + meshChannel);
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
    if (!isController && millis() - lastPublish >= 10 * 1000) {
        sensorData reading = getSensorData();

        StaticJsonDocument<100> json;
        json["Error"] = reading.error;
        json["Temperature"] = reading.temperature;
        json["Humidity"] = reading.humidity;

        String strReading;
        serializeJson(json, strReading);

        publish(strReading);

        lastPublish = millis();
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
        WriteFile(FILE_MESH_CONTROLLER, data["MeshController"]);
        changed = true;
    }

    if (data.containsKey("MeshChannel")) {
        WriteFile(FILE_MESH_CHANNEL, data["MeshChannel"]);
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

    } else if (FileExists(FILE_MESH_CONTROLLER) || FileExists(FILE_MESH_PEERS)) {
        LOGD("cmnd", "mesh is configured, setting flag");
        SetConfigured(true);

    } else {
        LOGD("cmnd", "not configured");
    }
}
