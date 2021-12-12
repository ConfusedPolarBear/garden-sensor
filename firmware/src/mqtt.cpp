#include <firmware.h>

WiFiClient wclient;
PubSubClient mqtt(wclient);
String clientId, baseTopic;
bool everConnected = false;

void setupTopics() {
    clientId = getIdentifier();
	baseTopic = "garden/module/" + clientId;
}

void connectToBroker() {
    if (mqtt.connected()) {
        return;
    }

    if (everConnected) {
        LOGW("mqtt", "Lost connection to broker, attempting to reconnect");
    }

    setupTopics();

    String host = ReadFile(FILE_MQTT_HOST);
    String user = ReadFile(FILE_MQTT_USER);
    String pass = ReadFile(FILE_MQTT_PASS);

    // Log the connection attempt and try to connect
    bool auth = user.length() > 0;
    LOGD("mqtt", "Connecting to " + host + ", username: " + (auth ? user : "[anonymous]") + ", client id: " + clientId);

    mqtt.set_server(host);
    if (!mqtt.connect(MQTT::Connect(clientId).set_auth(user, pass))) {
		LOGF("mqtt", "Failed to connect to broker");
	}

    LOGD("mqtt", "Connected to broker");

    // Set receive callback
	mqtt.set_callback(mqttReceiveCallback);

    // Subscribe to command topic
	String topic = baseTopic + "/cmnd";
	LOGD("mqtt", "Subscribing to " + topic);

	if (!mqtt.subscribe(topic)) {
		LOGF("mqtt", "Failed to subscribe to command topic");
	}

    LOGD("mqtt", "Listening for MQTT payloads");

    everConnected = true;
}

void mqttReceiveCallback(const MQTT::Publish& pub) {
    String topic = pub.topic();
    String payload = pub.payload_string();

    LOGD("mqtt", "Received message. Topic: " + topic);

    processCommand(payload);
}

void processMqtt() {
    if (!mqtt.connected()) {
        LOGF("mqtt", "process() called but not connected");
    }

    mqtt.loop();
}

bool publish(String data, String teleTopic) {
    // If this is called by a mesh client, the clientId & base topics won't have been populated yet so we have to do it manually
    if (clientId.length() == 0) {
        setupTopics();
    }

    bool isDiscovery = (teleTopic == "discovery");
    bool retain = isDiscovery;
    String topic = isDiscovery ? "garden/module/discovery/" + clientId : baseTopic + "/tele/" + teleTopic;

    if (everConnected) {
        LOGD("mqtt", "publishing " + String(data.length()) + " bytes to " + topic + ". retain: " + String(retain));

        if (!retain) {
            return mqtt.publish(topic, data);
        }

        auto len = data.length();
        uint8_t* u = (uint8_t*)data.c_str();
        return mqtt.publish(topic, u, len, true);
    }

    return publishMesh(data, topic);
}

bool publish(const JsonDocument& doc, const String topic) {
    String json;
    serializeJson(doc, json);
    return publish(json, topic);
}

String getClientId() {
    return clientId;
}
