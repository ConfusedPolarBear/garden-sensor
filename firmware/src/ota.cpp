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
}

void startUpdate(String wifi, String psk, String url, size_t length, String hash) {
    LOGD("ota", "starting OTA update from " + url + " through " + wifi + ". current connection is " + WiFi.SSID());
    LOGD("ota", "firmware is " + String(length) + " and has hash " + hash);

    bool isController = FileExists(FILE_MESH_CONTROLLER);
    HTTPClient http;
    WiFiClient client;
    int httpCode;
    size_t written;

    bool connect = false;
    if (wifi == "") {
        // If no network is specified, only proceed if we're already connected.
        if (WiFi.SSID() == "") {
            LOGW("ota", "no current connection & none provided");
            updateResult = "not connected to wifi network and none was specified";
            return;
        } else {
            // Assume that this network will work.
            LOGD("ota", "no connection specified but connected, assuming this is okay");
            connect = false;
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
            connect = false;
        }
    }

    if (connect) {
        if (isController) {
            // TODO: pause MQTT watchdog until OTA has a result & we're reconnected to the hardcoded network
            WiFi.disconnect();
        }

        // Connect to the new network.
        LOGD("ota", "connecting to " + wifi);
        WiFi.disconnect();
        LOGD("ota", "beginning new connection");
        WiFi.begin(wifi.c_str(), psk.c_str());
        LOGD("ota", "connecting");

        // Only try to connect to the network for 15s to prevent
        unsigned long start = millis();
        while (WiFi.status() != WL_CONNECTED && millis() - start <= 15 * 1000) {
            Serial << ".";
            yield();
        }
        Serial << endl;

        if (WiFi.status() != WL_CONNECTED) {
            // If the connection attempt failed, reconnect to the main network (connection is a no-op if not controller).
            // TODO: does this fuck with the mesh?
            updateResult = "Failed to connect to Wi-Fi";
            goto fail;
        }
    } else {
        LOGD("ota", "not changing wifi connections");
    }

    // Okay, we're connected to a network with access to the update server.
    // Perform initial internal setup and get ready to download the file.
    LOGD("ota", "setting update hash to " + hash + " and expected length to " + String(length));

    if (!Update.begin(length)) {
        updateResult = "not enough space to store update";
        goto fail;
    }

    if (!Update.setMD5(hash.c_str())) {
        updateResult = "failed to set checksum";
        goto fail;
    }

    Update.onProgress(onUpdateProgress);

    // Specifying the protocol is optional to save space in the JSON.
    if (!url.startsWith("http://")) {
        url = "http://" + url;
    }

    LOGD("ota", "starting download from " + url);

    // Ensure that we don't hang forever trying to get the new firmware.
    #ifdef ESP32
    http.setConnectTimeout(10000);
    #endif

    http.setRedirectLimit(5);
    http.setTimeout(10000);

    if (!http.begin(client, url)) {
        LOGD("ota", "unable to begin with specified url");
        updateResult = "failed to begin HTTP connection";
        goto fail;
    }

    // Make the request and make sure that the response is 200 OK.
    httpCode = http.GET();
    LOGD("ota", "download response code is " + String(httpCode));

    if (httpCode != HTTP_CODE_OK) {
        updateResult = "HTTP GET request failed with code " + String(httpCode);
        goto fail;
    }

    written = Update.writeStream(http.getStream());
    if (written != length) {
        updateResult = "bytes written does not equal firmware length";
        LOGW("ota", "wrote " + String(written) + " but firmware is known to be " + String(length));
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
    connectToWifi();
    return;
}
