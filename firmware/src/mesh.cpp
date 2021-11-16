#include <firmware.h>

std::vector<String> peers;
std::vector<String> paired;
bool controller;
int channel;

#ifdef ESP32
void meshSendCallback(const uint8_t* mac, esp_now_send_status_t status) {
    meshSendCallbackHandler(mac, status == ESP_NOW_SEND_SUCCESS);
}

void meshReceiveCallback(const uint8_t* mac, const uint8_t* data, int length) {
    meshReceiveCallbackHandler(mac, data, length);
}
#else
void meshSendCallback(uint8_t* mac, uint8_t status) {
    meshSendCallbackHandler(mac, status == 0);
}

void meshReceiveCallback(uint8_t* mac, uint8_t* data, uint8_t length) {
    meshReceiveCallbackHandler(mac, data, length);
}
#endif

void initializeMesh(bool isController, int chan) {
    controller = isController;
    channel = chan;

    LOGD("mesh", "starting mesh on channel " + String(channel));

    if (esp_now_init() != 0) {
		LOGF("mesh", "unable to start mesh");
	}

    #ifdef ESP32
    LOGD("mesh", "running on esp32, not setting role");
    #else
    LOGD("mesh", "running on esp8266, setting role");

    if (isController) {
        if (esp_now_set_self_role(ESP_NOW_ROLE_CONTROLLER) != 0) {
            LOGF("mesh", "unable to set controller role");
        }
    } else {
        if (esp_now_set_self_role(ESP_NOW_ROLE_SLAVE) != 0) {
            LOGF("mesh", "unable to set client role");
        }
    }
    #endif

    esp_now_register_send_cb(meshSendCallback);
	esp_now_register_recv_cb(meshReceiveCallback);

    LOGD("mesh", "started");
}

int loadPeers() {
    LOGD("mesh", "loading peer list");

    peers.clear();

    String raw = ReadFile(FILE_MESH_CONTROLLER) + ",";
    raw += ReadFile(FILE_MESH_PEERS);
    raw.toLowerCase();

    // Peers are separated by commas. A star at the start of a MAC address means it is connected to us via ESP-NOW.
    int index = raw.indexOf(",");
    while (index != -1) {
        String next = raw.substring(0, index);
        if (next.length() > 0) {
            if (next.startsWith("*")) {
                next.replace("*", "");
                addMeshPeer(next);
            }

            peers.push_back(next);
        }

        // LOGD("mesh", "possible peer is '" + next + "'");

        raw = raw.substring(index+1, raw.length());
        index = raw.indexOf(",");
    }

    int c = peers.size();
    LOGD("mesh", "found " + String(c) + " peers");

    return c;
}

bool parseMac(String mac, uint8_t dst[6]) {
    LOGD("mesh", "attempting to parse mac address " + mac);

    memzero(dst, 6);

    mac.toLowerCase();
    if (mac == "00:00:00:00:00:00" || mac == "ff:ff:ff:ff:ff:ff") {
        LOGW("mesh", "refusing to add forbidden mac address '" + mac + "'");
        return false;
    }

	int values[6];
	int i;

	if(sscanf(mac.c_str(), "%x:%x:%x:%x:%x:%x%*c",
		&values[0], &values[1], &values[2], &values[3], &values[4], &values[5]) == 6) {
		for(i = 0; i < 6; ++i) {
			dst[i] = (uint8_t) values[i];
		}

	} else {
		return false;
	}

    return true;
}

bool addMeshPeer(String mac) {
    uint8_t address[6];
    
    if (!parseMac(mac, address)) {
        LOGD("mesh", "parsed mac address");
        for (int i = 0; i < 6; i++) {
            LOGD("mesh", String(i) + ": " + String(address[i], HEX));
        }
        
        LOGW("mesh", "failed to add peer: failed to parse mac address");
        return false;
    }

    #ifdef ESP32
    esp_now_peer_info info;
    memcpy(info.peer_addr, address, 6);
    info.channel = channel;
    info.ifidx = WIFI_IF_AP;       // The ESP32 version of ESP-NOW requires you to select the iface that transmits the packet
    info.encrypt = false;

    if (esp_now_add_peer(&info) != ESP_OK) {
        LOGW("mesh", "failed to add peer: call to add_peer failed");
		return false;
	}
    #else
	if (esp_now_add_peer(address, ESP_NOW_ROLE_SLAVE, channel, NULL, 0) != 0) {
        LOGW("mesh", "failed to add peer: call to add_peer failed");
		return false;
	}
    #endif

    paired.push_back(mac);

    LOGD("mesh", "peer added");
	return true;
}

void publishMeshRaw(uint8_t* address, uint8_t* data) {
    // Messages sent with publishMesh() are always padded to 250 bytes (with 0x00).
    esp_now_send(address, data, 250);
}

void broadcastMesh(uint8_t* data, String exclude) {
    exclude.toLowerCase();
    LOGD("mesh", "broadcasting data to " + (exclude.length() == 0 ? "all peers" : ("all peers except " + exclude)));

    for (String mac : paired) {
        if (mac == exclude) {
            continue;
        }

        uint8_t address[6];
        if (!parseMac(mac, address)) {
            // This should *never* happen, as this vector is only appended to by addMeshPeer() after it validates the MAC
            LOGF("mesh", "unable to parse paired address '" + mac + "' as a MAC");
        }

        LOGD("mesh", "sending packet to " + mac);
        publishMeshRaw(address, data);
    }
}

bool publishMesh(String message, String topic) {
    // Message reassembly is done in the backend server
    String payload = topic + "\x01" + message;
    if (topic.length() == 0) {
        LOGD("mesh", "publishing payload without topic");
        payload = message;
    }

    float l = payload.length();
    uint32_t correlation = secureRandom();
    uint8_t total = ceil(l / 244.0);        // 244 bytes is the maximum payload (after the header is added).

    LOGD("mesh", "sending len:" + String(payload.length()) + ", cor:" + String(correlation, HEX) + ", tot:" + String(total));
    LOGD("mesh", "payload is '" + payload + "'");

    // Prepare each packet
    uint8_t number = 1;
    bool done = false;
    while(!done) {
        // Prepare the payload
        uint8_t wirePayload[250];
        memzero(wirePayload, 250);

        // Bytes 0 - 3 are the correlation ID.
        for (unsigned int i = 0; i < 4; i++) {
            wirePayload[3-i] = (correlation >> (i*8)) & 0xff;
        }

        // Byte 4 is the packet number.
        wirePayload[4] = number;

        // Byte 5 is the total number of packets.
        wirePayload[5] = total;

        // Bytes 6 - 218 are the payload.
        for(unsigned int i = 0; i < 212; i++) {
            if (i+1 > payload.length()) {
                // Once there's no more remaining data in the payload, the entire payload has been sent
                done = true;
                break;
            }

            wirePayload[6+i] = payload.charAt(i);
        }

        // On both the 8266 and the 32, the substring call checks if the string is shorter than the left position
        // and returns "" if true.
        payload = payload.substring(210, payload.length());

        // Bytes 218 - 249 are the SHA256 HMAC of all preceeding data. 250 bytes.
        uint8_t* auth = hmac(wirePayload, 217);
        for (int i = 0; i <= 32; i++) { wirePayload[i + 217] = auth[i]; }
        free(auth);

        #ifdef TRACE_PACKETS
        // Dump packet for debugging
        LOGD("mesh", "marshalled packet into wire format");
        for (unsigned int i = 0; i < 250; i++) {
            switch(i) {
                case 0:
                    LOGD("mesh", "correlation");
                    break;

                case 4:
                    LOGD("mesh", "number");
                    break;

                case 5:
                    LOGD("mesh", "total");
                    break;

                case 6:
                    LOGD("mesh", "payload");
                    break;

                case 217:
                    LOGD("mesh", "hmac");
            }

            uint8_t c = wirePayload[i];

            String index = String(i);
            if (index.length() == 1) {
                index = "00" + index;
            } else if (index.length() == 2) {
                index = "0" + index;
            }

            String body = String(c, HEX);
            if (body.length() == 1) {
                body = "0" + body;
            }

            LOGD("mesh", "payload byte " + index + " is " + body + " (" + char(c) + ")");
        }
        #endif

        // Send it
        broadcastMesh(wirePayload);

        number++;
    }

    return true;
}

void meshSendCallbackHandler(const uint8_t* mac, bool success) {
    LOGD("mesh", "result from send(): " + String(success));
}

void meshReceiveCallbackHandler(const uint8_t* mac, const uint8_t* buf, int length) {
	String payload;

    if (length != 250) {
        LOGD("mesh", "discarding mesh message: length != 250");
        return;
    }

    char strMacRaw[20];
    memzero(strMacRaw, 20);
    sprintf(strMacRaw, "%02x:%02x:%02x:%02x:%02x:%02x", mac[0], mac[1], mac[2], mac[3], mac[4], mac[5]);
    String strMac = String(strMacRaw);

    LOGD("mesh", "received message from " + strMac);
	for (int i = 0; i < length; ++i) {
		const char b = static_cast<char>(buf[i]);
		payload += b;
  	}

    // Copy the incoming (or actual) hmac out of the buffer for later verification.
    uint8_t actualHmac[32];
    for (int i = 0; i < 32; i++) { actualHmac[i] = buf[i + 217]; }

    // Calculate what the HMAC should be & compare them.
    uint8_t* expectedHmac = hmac(buf, 217);

    const bool hmacOk = hmacCompare(expectedHmac, actualHmac);

    if (!hmacOk) {
        LOGW("mesh", "discarding mesh message: authentication failed");

        LOGD("mesh", "packet hmac:   " + arrayToString(actualHmac, 32));
        LOGD("mesh", "expected hmac: " + arrayToString(expectedHmac, 32));

        free(expectedHmac);
        return;

    } else {
        LOGD("mesh", "accepting mesh message");
    }

    free(expectedHmac);

    if (!controller) {
        // If this is a command packet directed to us, handle it without rebroadcasting it.
        // Commands are addressed to a destination by including the text "dst-XXXXXXXXXXXX"
        // in the payload where X's are replaced with the MQTT client ID.
        String directFlag = "dst-" + getClientId();

        // Commands can also be sent to all connected system.
        String broadcastFlag = "dst-FFFFFFFFFFFF";

        bool isDirect = payload.indexOf(directFlag) != -1;
        bool isBroadcast = payload.indexOf(broadcastFlag) != -1;

        // If this is not a command packet directed at us OR it is a broadcast, rebroadcast the mesh message
        if (!isDirect || isBroadcast) {
            // Since this is a client, rebroadcast the packet to all peers *except* the sending device.
            broadcastMesh(const_cast<uint8_t*>(buf), strMac);
        }

        // If this is a directed or broadcast command, handle it.
        if (isDirect || isBroadcast) {
            payload = payload.substring(6, payload.length());

            LOGD("mesh", "handling command '" + payload + "'");
            processCommand(payload);
        }

    } else {
        // If this is the controller, send the packet over MQTT. Use substring to remove the HMAC.
        payload = payload.substring(0, 216);
	    publish(payload, "packet");
    }
}

bool isKnownPeer(String needle) {
    LOGD("mesh", "searching for peer " + needle);

    needle.toLowerCase();
    for (String haystack : peers) {
        haystack.toLowerCase();
        // LOGD("mesh", "checking against " + haystack);

        if (haystack == needle) {
            return true;
        }
    }

    return false;
}
