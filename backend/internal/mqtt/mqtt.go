package mqtt

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/ConfusedPolarBear/garden/internal/api"
	"github.com/ConfusedPolarBear/garden/internal/util"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

var clientIdRe *regexp.Regexp = regexp.MustCompile("^garden/module/([a-f0-9]+)/")

func Setup() {
	// Get configuration from environment variables
	host := os.Getenv("MQTT_HOST")
	username := os.Getenv("MQTT_USERNAME")
	password := os.Getenv("MQTT_PASSWORD")

	if host == "" || username == "" || password == "" {
		panic("host, username, or password is invalid")
	}

	// Setup local MQTT client options
	opts := mqtt.NewClientOptions().
		AddBroker(fmt.Sprintf("tcp://%s", host)).
		SetClientID("garden-backend").
		SetConnectTimeout(5 * time.Second).
		SetOrderMatters(false).
		SetUsername(username).
		SetPassword(password).
		SetKeepAlive(10 * time.Second).
		SetPingTimeout(2 * time.Second)

	// Connect to the MQTT broker
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	mustSubscribe(client, "garden/module/#", handleMqttMessage)
}

// Handle an incoming MQTT message.
func handleMqttMessage(c mqtt.Client, m mqtt.Message) {
	topic := m.Topic()
	payload := m.Payload()

	// If this message was sent by a garden system, extract it's client id from the topic.
	client := "SYSTEM"
	if clientIdRe.MatchString(topic) {
		client = clientIdRe.FindStringSubmatch(topic)[1]
	}

	logrus.Debugf("[mqtt] Message from %s: %s: %s\n", client, topic, payload)

	// Handle discovery messages
	if client == "SYSTEM" {
		// discovery message is garden/module/discovery/deadbeef
		p := strings.Split(topic, "/")
		id := p[len(p)-1]

		var info util.GardenSystemInfo
		if err := json.Unmarshal(payload, &info); err != nil {
			logrus.Warnf("[mqtt] failed to unmarshal discovery message from %s: %s", id, err)
			return
		}

		util.UpdateSystem(util.GardenSystem{
			Identifier:   id,
			Announcement: info,
			LastSeen:     time.Time{},
		})

		return
	}

	system, err := util.GetSystem(client)
	if err != nil {
		logrus.Warnf("[mqtt] unable to find system with id %s", client)
		return
	}

	system.LastSeen = time.Now()

	if err := json.Unmarshal(payload, &system.LastReading); err != nil {
		logrus.Warnf("[mqtt] unable to unmarshal reading: %s\n", err)
		return
	}

	// TODO: fix concurrency issues here
	util.UpdateSystem(system)

	api.BroadcastWebsocketMessage("update", system)
}

// Subscribe to the provided MQTT topic or panic.
func mustSubscribe(client mqtt.Client, topic string, callback func(c mqtt.Client, m mqtt.Message)) {
	if token := client.Subscribe(topic, 0, callback); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}
