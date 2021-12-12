#include <firmware.h>

void startAccessPoint(int channel) {
    /* Starts an access point which does not allow any clients to connect. This is enforced through two measures:
     *   1. The maximum number of client connections is set to 0.
     *   2. The password is generated securely and is 192 bits long.
     * The access point is only used to find the most efficient route between nodes in mesh mode.
     */

    // By default, access point settings (SSID and PSK) are written to flash whenever they are changed.
    // Since the PSK is randomized every time it is started, this would wear out the flash very quickly.
    WiFi.persistent(false);

    String apSsid = "m-" + getIdentifier();
    apSsid.toLowerCase();

    String apPass = secureRandomNonce();

    LOGD("ap", "Starting access point on channel " + String(channel) + ". SSID: " + apSsid);
    WiFi.softAP(apSsid.c_str(), apPass.c_str(), channel, 1, 0);       // ssid, psk, channel, hidden, max connections
}

void stopAccessPoint() {
    LOGD("ap", "Stopping access point");
    WiFi.softAPdisconnect(false);
}

void connectToWifi() {
    String wifiSsid = ReadFile(FILE_WIFI_SSID);
    String wifiPass = ReadFile(FILE_WIFI_PASS);

    if (wifiSsid == "") {
        return;
    }

    // Attempt to connect to the provided Wi-Fi network
    WiFi.begin(wifiSsid.c_str(), wifiPass.c_str());

    Serial << "Connecting to Wi-Fi";
    while (WiFi.status() != WL_CONNECTED)
    {
        Serial << ".";
        processCommand(Serial.readStringUntil('\n'), true);
    }
    Serial << endl;

    Serial << "IP address: " << WiFi.localIP() << endl;

    connectToBroker();
}

void startNetworkScan() {
    if (WiFi.scanComplete() == -1) {
        LOGD("wifi", "Network scan already in progress");
        return;
    }

    LOGD("wifi", "Starting async network scan");
    WiFi.scanNetworks(true, true);
}

struct network {
    bool   Known;
    String BSSID;
    int    RSSI;
};

bool compareNetworks(network i, network j) {
    // Expected to return the result of <, but RSSI is negative.
    return i.RSSI > j.RSSI;
}

void processNetworkScan() {
    // If still scanning, returns -1. If no scan triggered yet, returns -2. Otherwise returns the number of networks.
    int n = WiFi.scanComplete();
    if (n <= 0) {
        return;
    }

    DynamicJsonDocument doc(1024);
    std::vector<network> networks;

    LOGD("wifi", "scan found " + String(n) + " network(s)");
    for (int i = 0; i < n; i++) {
        bool hidden = false;
        #ifdef ESP32
        hidden = WiFi.SSID(i).length() == 0;
        #else
        hidden = WiFi.isHidden(i);
        #endif

        // Access points generated by nodes are hidden.
        if (!hidden) {
            continue;
        }

        String bssid = WiFi.BSSIDstr(i);
        networks.push_back(network{ isKnownPeer(bssid), bssid, WiFi.RSSI(i) });
    }

    std::sort(networks.begin(), networks.end(), compareNetworks);

    // Delete the stored scan results.
    WiFi.scanDelete();

    JsonArray jsonNetworks = doc.to<JsonArray>();
    for (network i : networks) {
        JsonObject net = jsonNetworks.createNestedObject();
        net["Known"] = i.Known;
        net["MAC"] = i.BSSID;
        net["RSSI"] = i.RSSI;
    }

    publish(doc, "networks");
}

#warning pass a string vector with discovered sensors
void sendDiscoveryMessage(bool useMqtt) {
    StaticJsonDocument<250> info;
    
    // Store reset reason and sdk version. Since the ESP32 does not expose the sdk version, it's only sent by the ESP8266.
    #ifdef ESP8266
    info["RR"] = ESP.getResetReason();
    info["CV"] = ESP.getCoreVersion();
    #endif

    info["SV"] = ESP.getSdkVersion();

    // Store filesystem used and total byte counts.
    info["FU"] = 0;
    info["FT"] = 0;
    FSInfo fsInfo;
    if (GetFSInfo(&fsInfo)) {
        info["FU"] = fsInfo.usedBytes;
        info["FT"] = fsInfo.totalBytes;
    }

    // If this uses the mesh or not. If not (i.e. connected through MQTT), report the wifi channel.
    info["ME"] = !useMqtt;
    if (useMqtt) {
        info["CH"] = WiFi.channel();
    }

    JsonArray sensors = info.createNestedArray("Sensors");
    sensors.add("temperature");
    sensors.add("humidity");

    publish(info, "discovery");
}

String getIdentifier(bool includeColons) {
    String clientId = WiFi.softAPmacAddress();

    if (!includeColons) {
        clientId.replace(":", "");
    }

    return clientId;
}
