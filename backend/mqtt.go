package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

var systems map[string]GardenSystem = make(map[string]GardenSystem)

// Mutex used to guard access to the map of systems. This is because concurrent map reads and writes are illegal in Go.
// While there is a dedicated sync.Map type, using a plain Mutex here is preferred (ref https://pkg.go.dev/sync#Map).
var systemLock sync.Mutex

var clientIdRe *regexp.Regexp = regexp.MustCompile("^garden/module/([a-f0-9]+)/")

func SetupMQTT() {
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
	payload := string(m.Payload())

	// If this message was sent by a garden system, extract it's client id from the topic.
	client := "SYSTEM"
	if clientIdRe.MatchString(topic) {
		client = clientIdRe.FindStringSubmatch(topic)[1]
	}

	logrus.Debugf("[mqtt] Message from %s: %s: %s\n", client, topic, payload)

	// Handle discovery messages
	if client == "SYSTEM" {
		systemLock.Lock()
		defer systemLock.Unlock()

		// discovery message is garden/module/discovery/deadbeef
		p := strings.Split(topic, "/")
		id := p[len(p)-1]

		// If this is the first time we've seen this system, insert a new entry into the system map
		if _, okay := systems[id]; !okay {
			systems[id] = GardenSystem{
				Identifier: id,
				LastSeen:   time.Time{},
			}
		}

		return
	}

	systemLock.Lock()
	defer systemLock.Unlock()

	system := systems[client]
	system.LastSeen = time.Now()

	if err := json.Unmarshal([]byte(payload), &system.LastReading); err != nil {
		logrus.Warnf("[mqtt] unable to unmarshal reading: %s\n", err)
		return
	}

	systems[client] = system
}

// Subscribe to the provided MQTT topic or panic.
func mustSubscribe(client mqtt.Client, topic string, callback func(c mqtt.Client, m mqtt.Message)) {
	if token := client.Subscribe(topic, 0, callback); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}
