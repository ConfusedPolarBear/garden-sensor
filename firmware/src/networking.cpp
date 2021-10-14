#include <firmware.h>

void startAccessPoint() {
    /* Starts an access point which does not allow any clients to connect. This is enforced through two measures:
     *   1. The maximum number of client connections is set to 0.
     *   2. The password is generated securely and is 192 bits long.
     * The access point is only used to find the most efficient route between nodes in mesh mode.
     */

    // By default, access point settings (SSID and PSK) are written to flash whenever they are changed.
    // Since the PSK is randomized every time it is started, this would wear out the flash very quickly.
    WiFi.persistent(false);

    // Remove the colons from the MAC address and lowercase it
    String apSsid = "m-" + WiFi.macAddress();
    apSsid.replace(":", "");
    apSsid.toLowerCase();

    String apPass = secureRandom();

    LOGD("ap", "Starting access point. SSID: " + apSsid + ", PSK: " + apPass);
    WiFi.softAP(apSsid.c_str(), apPass.c_str(), 1, 0, 0);
}

void stopAccessPoint() {
    LOGD("ap", "Stopping access point");
    WiFi.softAPdisconnect(false);
}

void startNetworkScan() {
    if (WiFi.scanComplete() == -1) {
        LOGD("wifi", "Network scan already in progress");
        return;
    }

    LOGD("wifi", "Starting async network scan");
    WiFi.scanNetworks(true);
}

void processNetworkScan() {
    // If still scanning, returns -1. If no scan triggered yet, returns -2. Otherwise returns the number of networks.
    int n = WiFi.scanComplete();
    if (n <= 0) {
        return;
    }

    LOGD("wifi", "found " + String(n) + " network(s)");
    for (int i = 0; i < n; i++) {
        String msg = "Details for network " + String(i) + ": ";
        msg += "SSID:" + WiFi.SSID(i) + ",";
        msg += "RSSI:" + String(WiFi.RSSI(i));
    
        LOGD("wifi", msg);
    }

    LOGD("wifi", "End scan results");

    // Delete the stored scan results.
    WiFi.scanDelete();
}
