#include "firmware.h"

#ifdef ESP32
#include <Update.h>         // this is absolutely headass
#include <HTTPClient.h>
#else
#include <Updater.h>        // FFS
#include <ESP8266HTTPClient.h>
#endif

#include <WiFiClient.h>

String updateResult = "Unknown failure";

String getUpdateMessage() {
    LOGD("ota", "result from update is " + updateResult);

    // Only return this result once.
    // TODO: move to a dedicated clearMessage() function.
    String r = updateResult;
    updateResult = "";

    return r;
}

void onUpdateProgress(size_t lhs, size_t rhs) {
    float percent = (lhs * 100) / rhs;
    LOGD("ota", "update progress: " + String(percent) + "%");

    if ((int)percent % 10 == 0) {
        flashLed();
    }
}

void startUpdate(String wifi, String psk, String url, size_t length, String checksum) {
    LOGD("ota", "starting OTA update from " + url + " through " + wifi + ". current connection is " + WiFi.SSID());
    LOGD("ota", "firmware is " + String(length) + " and has checksum " + checksum);

    HTTPClient http;
    WiFiClient client;
    int httpCode;
    size_t totalWritten;
    bool connect = false;

    // Figure out if we need to change Wi-Fi networks.
    if (wifi == "") {
        // If no network is specified, only proceed if we're already connected to any network.
        if (WiFi.SSID() == "") {
            // Not connected, no way to download the firmware.
            LOGW("ota", "no current connection & none provided");
            updateResult = "not connected to wifi network and none was specified";
            return;
        } else {
            // We're connected to something, assume that it will work.
            LOGD("ota", "no connection specified but connected, assuming this is okay");
        }

    } else {
        // A network was specified. If we're already connected to it, great! If not, connect to it.
        if (WiFi.SSID() != wifi) {
            // Current connection is wrong, reconnect.
            LOGD("ota", "connection specified and connected to different network, changing networks");
            connect = true;
        } else {
            // Already connected, don't do anything.
            LOGD("ota", "connection specified and already connected to it");
        }
    }

    // Initial setup is here so we can fail fast if something's wrong.
    LOGD("ota", "setting update checksum to " + checksum + " and expected length to " + String(length));

    if (!Update.begin(length)) {
        updateResult = "not enough space to store update";
        goto fail;
    }

    if (!Update.setMD5(checksum.c_str())) {
        updateResult = "failed to set checksum";
        goto fail;
    }

    Update.onProgress(onUpdateProgress);

    // Specifying the protocol is optional to save space in the JSON.
    if (!url.startsWith("http://")) {
        url = "http://" + url;
    }

    // Connect to the new Wi-Fi network (if needed).
    if (connect) {
        WiFi.begin(wifi.c_str(), psk.c_str());
        LOGD("ota", "connecting to " + wifi);

        // Only try to connect to the network for a short time to prevent the system from hanging forever here with bad creds.
        unsigned long start = millis();
        while (WiFi.status() != WL_CONNECTED && millis() - start <= 15 * 1000) {
            Serial << ".";
            delay(500);
        }
        Serial << endl;

        if (WiFi.status() != WL_CONNECTED) {
            // If the connection attempt failed, reconnect to the main network (connection is a no-op if not controller).
            // TODO: does this fuck with the mesh?
            updateResult = "Failed to connect to Wi-Fi";
            goto fail;
        }
    } else {
        LOGD("ota", "not changing wifi connection");
    }

    // Okay, we're connected to a network with access to the update server, download & install the firmware.
    LOGD("ota", "starting download from " + url);

    #ifdef ESP32
    // Ensure that we don't hang forever trying to get the new firmware.
    http.setConnectTimeout(10000);
    #endif

    http.setRedirectLimit(5);
    http.setTimeout(10000);

    if (!http.begin(client, url)) {
        updateResult = "failed to begin HTTP connection";
        goto fail;
    }

    // The ID of the system is included so the server can monitor when a node starts downloading firmware.
    http.addHeader("System-ID", getIdentifier());

    // Make the request and make sure that the response is 200 OK.
    httpCode = http.GET();
    LOGD("ota", "download response code is " + String(httpCode));

    if (httpCode != HTTP_CODE_OK) {
        updateResult = "HTTP GET request failed with code " + String(httpCode);
        goto fail;
    }

    // Verify that the user sent the correct firmware length
    if ((int)length != http.getSize()) {
        LOGW("ota", "length and content-length header mismatch, expected " + String(length) + ", got " + String(http.getSize()));
        updateResult = "length mismatch";
        goto fail;
    }

    // Here we go...
    totalWritten = Update.writeStream(http.getStream());
    if (totalWritten != length) {
        LOGW("ota", "wrote " + String(totalWritten) + " bytes but firmware is known to be " + String(length) + " bytes");
        updateResult = "bytes written does not equal firmware length";
        goto fail;
    }

    LOGD("ota", "wrote new firmware, verifying checksum");
    if (!Update.end()) {
        updateResult = "failed to verify checksum";
        goto fail;
    }

    LOGD("ota", "wrote new firmware successfully");
    ESP.restart();

    fail:
    LOGW("ota", "error: " + updateResult + " (" + String(Update.getError()) + ")");

    #warning USE CORRECT MESH CHANNEL HERE
    startAccessPoint(7);
}
