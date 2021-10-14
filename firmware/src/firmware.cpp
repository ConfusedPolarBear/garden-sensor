#include <firmware.h>

String wifiSsid, wifiPass;
String mqttHost, mqttUser, mqttPass;

bool isController;
String meshController;
std::vector<String> meshPeers;
int meshChannel;

String apSsid;

void setup() {
    Serial.begin(115200);
    Serial.setTimeout(500);     // timeout for readStringUntil()

    Serial << endl << endl;

    InitializeFS();

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
    #warning TODO: load mesh peers
    #warning TODO: parse mesh channel

    Serial << "Settings:" << endl;
    if (meshController == "") {
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
    if (isController) {
        WiFi.mode(WIFI_AP_STA);

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
    } else {
        WiFi.mode(WIFI_AP);
    }

    startAccessPoint();

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

   // Initialize ESP-NOW

   // Publish discovery message with system information and configured sensors
}

void loop() {
    // Process any Serial commands
    processCommand(Serial.readStringUntil('\n'));

    // If a Wi-Fi scan was requested, process the results
    processNetworkScan();

    // Publish sensor data

    // TODO: if controller coordinates deep sleep, do that
}

// Returns a 192-bit secure random number
String secureRandom() {
    String rand;
    for (int i = 0; i < 6; i++) {
        rand += String(ESP.random(), HEX);
    }

    return rand;
}

void processCommand(String command) {
    if (command.length() == 0) {
        return;
    }

    LOGD("command", "Processing command '" + command + "'");

    if (command == "scan") {
        startNetworkScan();
    }
}
