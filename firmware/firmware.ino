#include <ESP8266WiFi.h>
#include <Streaming.h>
#include <PubSubClient.h>

#include "secrets.h"

WiFiClient wclient;
PubSubClient mqtt(wclient);

String chipId;
String baseTopic;

// Sensor publish delay in ms.
int publishDelay = 2500;

bool publish(String topic, String data, bool retain = false);
void callback(const MQTT::Publish& pub);
void fatalError(String msg);

void setup()
{
	Serial.begin(115200);

	// Display non-sensitive settings for validation
	Serial << endl;
	Serial << "Settings:" << endl;
	Serial << "\tWi-Fi SSID:    " << wifiSsid << endl;
	Serial << "\tMQTT broker:   " << mqttAddress << endl;
	Serial << "\tMQTT username: " << mqttUsername << endl;
	Serial << endl;

	// Attempt to connect to the provided Wi-Fi network
	WiFi.begin(wifiSsid, wifiPassword);

	Serial << "Connecting to Wi-Fi";
	while (WiFi.status() != WL_CONNECTED)
	{
		delay(500);
		Serial << ".";
	}
	Serial << endl;

	Serial << "My IP address is " << WiFi.localIP() << endl;

	// Every ESP chip has a 32 bit unique identifier set at the factory. This is used to uniquely identify this system.
	chipId = String(ESP.getChipId(), HEX);

	// Setup the MQTT broker connection.
	mqtt.set_server(mqttAddress);

	Serial << "MQTT using authentication: ";
	if (strlen(mqttUsername) != 0) {
		Serial << "yes" << endl;
	} else {
		Serial << "no" << endl;
	}

	// Connect to the MQTT broker
	if (!mqtt.connect(MQTT::Connect(chipId).set_auth(mqttUsername, mqttPassword))) {
		fatalError("Failed to connect to MQTT broker");
	}

	Serial << "Successfully connected to broker" << endl;

	// Set a callback that is called when a MQTT message arrives.
	mqtt.set_callback(callback);

	// Subscribe to the MQTT topic "garden/module/00000000/cmnd/#" where the zeros are replaced with this chip's id.
	baseTopic = "garden/module/";
	baseTopic.concat(chipId);

	String topic = baseTopic;
	topic.concat("/cmnd/#");

	Serial << "Subscribing to MQTT topic " << topic << endl;
	if (!mqtt.subscribe(topic)) {
		fatalError("Failed to subscribe to MQTT topic");
	}

	Serial << "Successfully subscribed" << endl;

	// Publish a discovery message announcing this system's availability.
	// The discovery message is sent to "garden/module/discovery/00000000".
	publish("garden/module/discovery/" + chipId, "available", true);
}

void loop() {
	// Send data to the broker formatted as JSON.
	String data = F("{\"Temperature\":TEMP,\"Humidity\":HUMI}");
	data.replace("TEMP", String(ESP.random()));
	data.replace("HUMI", String(ESP.random()));

	// Publish readings to "garden/module/00000000/data"
	publish(baseTopic + "/tele/data", data);
	delay(publishDelay);
}

bool publish(String topic, String data, bool retain) {
	// Process any incoming packets.
	mqtt.loop();

	Serial << "Publishing " << data.length() << " bytes to " << topic << ": " << data << endl;

	if (!retain) {
		return mqtt.publish(topic, data);
	}

	auto len = data.length();
	uint8_t* u = (uint8_t*)data.c_str();
	return mqtt.publish(topic, u, len, true);
}

// Publishes the result of running a command.
void publishResult(String topic, String command, bool success, String msg) {
	String json = F("{\"Success\":SUCCESS,\"Command\":\"COMMAND\",\"Message\":\"MESSAGE\"}");
	
	if (success) {
		json.replace("SUCCESS", "true");
	} else {
		json.replace("SUCCESS", "false");
	}

	json.replace("COMMAND", command);
	json.replace("MESSAGE", msg);

	publish(topic, json);
}

void callback(const MQTT::Publish& pub) {
	auto topic = pub.topic();
	auto msg = pub.payload_string();

	/* Extracts the command from the topic. Commands are at the end of the topic.
	 * For example, the topic "garden/module/12345678/cmnd/restart" would parse to "restart".
	 */
	const String command = topic.substring(topic.lastIndexOf("/")+1, command.length());
	const String result = baseTopic + "/result";

	bool resultOk = false;
	String resultMessage = "";

	Serial << topic << " => " << msg << endl;

	// Restarts the system.
	if (command == "restart") {
		if (msg != "1") {
			publishResult(result, "restart", false, "Set payload to 1 to confirm restart");
			return;
		}

		// Can't use resultMessage here because the ESP restarts before that happens
		publishResult(result, "restart", true, "Restarting");

		// Delay to give MQTT time to receive the message
		delay(500);
		ESP.restart();

	// Adjusts publishing frequency.
	} else if (command == "frequency") {
		const int freq = msg.toInt();

		if (freq < 1000) {
			resultMessage = "Invalid publish duration";
		} else {
			publishDelay = freq;
			resultOk = true;
			resultMessage = "Successfully set publish duration";
		}

	} else {
		resultMessage = "Invalid command";
	}

	publishResult(result, command, resultOk, resultMessage);
}

void fatalError(String msg) {
	Serial << "Fatal error: " << msg << endl << "The system will now shutdown." << endl;
	Serial.flush();

	while(true) {
		// Yield time back to the OS so the watchdog doesn't reset the system.
		yield();
	}
}