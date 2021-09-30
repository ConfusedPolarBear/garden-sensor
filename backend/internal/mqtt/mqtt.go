package mqtt

import (
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/ConfusedPolarBear/garden/internal/api"
	"github.com/ConfusedPolarBear/garden/internal/config"
	"github.com/ConfusedPolarBear/garden/internal/util"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

var clientIdRe *regexp.Regexp = regexp.MustCompile("^garden/module/([a-f0-9]+)/")
var mqttClient mqtt.Client

func Setup(isServer bool) {
	// Get configuration from environment variables
	host := config.GetString("mqtt.host")
	username := config.GetString("mqtt.username")
	password := config.GetString("mqtt.password")

	if host == "" {
		panic("mqtt host is required")
	}

	if h, p, err := net.SplitHostPort(host); err != nil {
		panic("mqtt host is malformed. required format is ADDRESS:PORT")
	} else {
		logrus.Debugf("[mqtt] will connect to broker %s on port %s", h, p)
	}

	clientId := "garden-backend"
	if !isServer {
		clientId = "123456"
	}

	logrus.Debugf("[mqtt] backend server mode: %t, using client id %s", isServer, clientId)

	// Setup local MQTT client options
	opts := mqtt.NewClientOptions().
		AddBroker(fmt.Sprintf("tcp://%s", host)).
		SetClientID(clientId).
		SetConnectTimeout(5 * time.Second).
		SetOrderMatters(false).
		SetKeepAlive(30 * time.Second).
		SetPingTimeout(2 * time.Second)

	if username != "" {
		logrus.Debug("[mqtt] connection will be authenticated")

		opts.
			SetUsername(username).
			SetPassword(password)
	} else {
		logrus.Debug("[mqtt] connection will be unauthenticated")
	}

	// Connect to the MQTT broker
	mqttClient = mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(fmt.Errorf("failed to connect to MQTT broker: %s", token.Error()))
	}

	if isServer {
		Subscribe("garden/module/#", handleMqttMessage)
	}
}

// Publishes to the provided topic or panics.
func Publish(topic, payload string) {
	PublishAdvanced(topic, payload, 0, false)
}

func PublishAdvanced(topic, payload string, qos int, retain bool) {
	if token := mqttClient.Publish(topic, byte(qos), retain, payload); token.WaitTimeout(2*time.Second) && token.Error() != nil {
		panic(fmt.Errorf("failed to publish message to %s: %s", topic, token.Error()))
	}

	logrus.Debugf("[mqtt] published message to %s (l %d, q %d, r %t)", topic, len(payload), qos, retain)
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
	if strings.Contains(topic, "/discovery") {
		if len(payload) == 0 {
			return
		}

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
func Subscribe(topic string, callback func(c mqtt.Client, m mqtt.Message)) {
	logrus.Debugf("[mqtt] subscribing to topic %s", topic)

	if token := mqttClient.Subscribe(topic, 0, callback); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}
