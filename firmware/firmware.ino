#include <ESP8266WiFi.h>
#include <PubSubClient.h>
#include "libraries/Streaming.h"
#include "libraries/ArduinoJson.h"
#include "libraries/Filesystem.h"

WiFiClient wclient;
PubSubClient mqtt(wclient);

String wifiSsid, wifiPassword, mqttAddress, mqttUsername, mqttPassword;
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

	Serial << endl << "Press s within in the next 2 seconds to enter setup mode" << endl;
	delay(2000);

	// Go into setup mode if there is no configuration data saved or the user pressed 's'.
	Filesystem* fs = Filesystem::Get();
	if (!fs->IsConfigured() || Serial.read() == 's') {
		Serial << "Entered Setup mode" << endl;

		while(true) {
			parseSerial();
		}
	}

	// TODO: config data could probably be saved as JSON
	// Load all configuration data
	wifiSsid = fs->ReadFile("/wifiSSID");
	wifiPassword = fs->ReadFile("/wifiPass");
	mqttAddress = fs->ReadFile("/mqttAddr");
	mqttUsername = fs->ReadFile("/mqttUser");
	mqttPassword = fs->ReadFile("/mqttPass");

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

		parseSerial();
	}
	Serial << endl;

	Serial << "My IP address is " << WiFi.localIP() << endl;

	// Every ESP chip has a 32 bit unique identifier set at the factory. This is used to uniquely identify this system.
	chipId = String(ESP.getChipId(), HEX);

	// Setup the MQTT broker connection.
	mqtt.set_server(mqttAddress);

	Serial << "MQTT using authentication: ";
	if (mqttUsername.length() != 0) {
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
	baseTopic = "garden/module/" + chipId;

	String topic = baseTopic + "/cmnd/#";

	Serial << "Subscribing to MQTT topic " << topic << endl;
	if (!mqtt.subscribe(topic)) {
		fatalError("Failed to subscribe to MQTT topic");
	}

	Serial << "Successfully subscribed" << endl;

	/* Publish a discovery message announcing this system's availability, basic info, and capabilities.
	 * The discovery message is sent to "garden/module/discovery/00000000".
	 * For an example of how to work with ArduinoJSON, go here: https://arduinojson.org/v6/example/generator
	*/
	StaticJsonDocument<256> info;
	info["System"]["RestartReason"] = ESP.getResetReason();
	info["System"]["CoreVersion"] = ESP.getCoreVersion();
	info["System"]["SdkVersion"] = ESP.getSdkVersion();
	info["System"]["FlashSize"] = ESP.getFlashChipSize();
	info["System"]["RealFlashSize"] = ESP.getFlashChipRealSize();

	// TODO: dynamically populate the list of sensors
	JsonArray sensors = info.createNestedArray("Sensors");
	sensors.add("temperature");
	sensors.add("humidity");

	String strInfo;
	serializeJson(info, strInfo);

	publish("garden/module/discovery/" + chipId, strInfo, true);
}

void loop() {
	parseSerial();

	StaticJsonDocument<100> reading;
	reading["Temperature"] = random(0, 50);
	reading["Humidity"] = random(0, 100);

	String strReading;
	serializeJson(reading, strReading);

	// Publish readings to "garden/module/00000000/data"
	publish(baseTopic + "/tele/data", strReading);
	delay(publishDelay);
}

bool publish(String topic, String data, bool retain) {
	// Process any incoming packets.
	mqtt.loop();

	// Log the published packet to the serial port as JSON
	StaticJsonDocument<1024> log;
	log["Success"] = true;
	log["Command"] = "Publish";
	log["Topic"] = topic;
	log["Retain"] = retain;

	if (data.length() >= 512) {
		log["Data"] = data.substring(0, 512);
	} else {
		log["Data"] = data;
	}

	serializeJson(log, Serial);
	Serial.println();

	if (!retain) {
		return mqtt.publish(topic, data);
	}

	auto len = data.length();
	uint8_t* u = (uint8_t*)data.c_str();
	return mqtt.publish(topic, u, len, true);
}

// Publishes the result of running a command.
void publishResult(String topic, String command, bool success, String msg) {
	// TODO: check if msg.length >= 450ish and reset message to an error
	StaticJsonDocument<512> result;
	result["Success"] = success;
	result["Command"] = command;
	result["Message"] = msg;

	String strResult;
	serializeJson(result, strResult);

	publish(topic, strResult);
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

		Filesystem::Get()->Unmount();

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

// Command: {"Command":"restart"}
// Configuration: {"WifiSSID":"","WifiPassword":"","MQTTHost":"","MQTTUsername":"","MQTTPassword":""}
void parseSerial() {
	const long avail = Serial.available();
	if (avail <= 0 || avail >= 1024) {
		return;
	}

	Filesystem* fs = Filesystem::Get();

	StaticJsonDocument<512> result;
	result["Success"] = false;
	result["Message"] = "";

	StaticJsonDocument<1024> data;
	DeserializationError error = deserializeJson(data, Serial);
	if (error) {
		result["Message"] = String("Failed to deserialize Serial input. ") + error.c_str();

		serializeJson(result, Serial);
		Serial.println();

		return;
	}

	// If a command is present, run it first.
	if (data.containsKey("Command")) {
		const String command = data["Command"];

		// Restarts the device immediately
		if (command == "restart") {
			fs->Unmount();
			ESP.restart();

		// Basic info request from the web interface. Used to validate the chosen serial port.
		} else if (command == "info") {
			result["Success"] = true;
			result["Message"] = "modular-garden-system";
			result["ID"] = String(ESP.getChipId(), HEX);

			serializeJson(result, Serial);
			Serial.println();
			return;

		// Factory reset this device WITHOUT CONFIRMATION.
		} else if (command == "reset") {
			result["Success"] = true;
			serializeJson(result, Serial);
			Serial.println();

			delay(500);

			fs->Format();
			ESP.restart();

		} else {
			result["Message"] = "Unknown command";
			result["Command"] = command;

			serializeJson(result, Serial);
			Serial.println();
			return;
		}
	}

	// Wi-Fi SSID must be present to configure Wi-Fi.
	if (data.containsKey("WifiSSID")) {
		fs->WriteFile("/wifiSSID", data["WifiSSID"]);
		fs->WriteFile("/wifiPass", data["WifiPassword"]);

		result["Success"] = true;
		result["ConfiguredWiFi"] = true;
	}

	// If the MQTT host is present, configure MQTT.
	if (data.containsKey("MQTTHost")) {
		fs->WriteFile("/mqttAddr", data["MQTTHost"]);

		if (data.containsKey("MQTTUsername")) {
			fs->WriteFile("/mqttUser", data["MQTTUsername"]);
			fs->WriteFile("/mqttPass", data["MQTTPassword"]);
		}

		result["Success"] = true;
		result["ConfiguredMQTT"] = true;
		result["MQTTAuthenticated"] = fs->FileExists("/mqttUser");
	}

	// If both Wi-Fi and MQTT are configured, set the configured flag.
	if (fs->FileExists("/wifiSSID") && fs->FileExists("/mqttAddr")) {
		fs->SetConfigured(true);
	}

	serializeJson(result, Serial);
	Serial.println();
}

void fatalError(String msg) {
	Serial << "Fatal error: " << msg << endl << "The system will now shutdown." << endl;
	Serial.flush();

	while(true) {
		parseSerial();
	}
}