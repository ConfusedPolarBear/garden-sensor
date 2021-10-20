package mqtt

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ConfusedPolarBear/garden/internal/api"
	"github.com/ConfusedPolarBear/garden/internal/config"
	"github.com/ConfusedPolarBear/garden/internal/util"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

var clientIdRe *regexp.Regexp = regexp.MustCompile("^garden/module/([a-fA-F0-9]+)/")
var mqttClient mqtt.Client

// TODO: warn when a packet has been queued for longer than 5 seconds
type meshPacket struct {
	ArrivalTime time.Time

	Number  uint16
	Total   uint16
	Topic   string
	Payload []byte
}

var meshPacketsLock sync.Mutex
var meshPackets map[string][]meshPacket = map[string][]meshPacket{}

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
		Subscribe("garden/module/#", onMqttMessage)
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
func onMqttMessage(c mqtt.Client, m mqtt.Message) {
	topic := m.Topic()
	payload := m.Payload()

	// If this message was sent by a garden system, extract it's client id from the topic.
	client := "SYSTEM"
	if clientIdRe.MatchString(topic) {
		client = clientIdRe.FindStringSubmatch(topic)[1]
	}

	handleMqttMessage(client, topic, payload)
}

// TODO: create a function that loops through all queued packets and alerts if any are older than 5 seconds.

func handleMqttMessage(client, topic string, payload []byte) {
	// Minified discovery message. Must be compatible with the full GardenSystemInfo struct.
	type miniInfo struct {
		System struct {
			IsEmulator          bool
			IsMesh              bool   `json:"ME"`
			RestartReason       string `json:"RR"`
			CoreVersion         string `json:"CV"`
			SdkVersion          string `json:"SV"`
			FilesystemUsedSize  int    `json:"FU"`
			FilesystemTotalSize int    `json:"FT"`
		}
		Sensors []string
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

		var miniInfo miniInfo
		if err := json.Unmarshal(payload, &miniInfo); err != nil {
			logrus.Warnf("[mqtt] failed to unmarshal discovery message from %s: %s", id, err)
			return
		}

		info := util.GardenSystemInfo(miniInfo)
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

	if strings.Contains(topic, "/tele/") {
		if strings.HasSuffix(topic, "/data") {
			// Sensor readings
			if err := json.Unmarshal(payload, &system.LastReading); err != nil {
				logrus.Warnf("[mqtt] unable to unmarshal reading: %s\n", err)
				return
			}

		} else if strings.HasSuffix(topic, "/networks") {
			// Wi-Fi scan results
			type network struct {
				Known bool
				MAC   string
				RSSI  int
			}

			var results []network
			if err := json.Unmarshal(payload, &results); err != nil {
				logrus.Warnf("[mqtt] unable to unmarshal scan results: %s\n", err)
				return
			}

			logrus.Debugf("[mqtt] found %d networks", len(results))
			for i, n := range results {
				// RSSI ranges from 0 to -100 (ish)
				signal := -1 * n.RSSI
				known := "     "
				if n.Known {
					known = "!!!!!"
				}
				logrus.Debugf("[mqtt] network %d (%s) belongs to %s and has signal %d", i+1, known, n.MAC, signal)
			}

		} else if strings.HasSuffix(topic, "/packet") {
			/*
				| Payload index | Description    |
				|---------------|----------------|
				| 0 - 3         | Correlation ID |
				| 4             | Packet number  |
				| 5             | Total packets  |
				| 6 - end       | Payload        |

				Correlation IDs are random 32 bit unsigned integers generated by the node that created the packet.
					They are used by the backend server to recombine fragmented ESP-NOW packets.

				Packet numbers and the total number of packets start at 1 and can go up to 255. The theoretical maximum
					payload size is 62,220 bytes, or about 60 KB (244 * 255 bytes).
			*/

			logrus.Tracef("[mqtt] raw packet is %s", hex.EncodeToString(payload))

			correlation := hex.EncodeToString(payload[:3])

			// Nodes only send 8 bit unsigned integers but the smallest number in the binary package is uint16
			number := binary.BigEndian.Uint16([]byte{0x00, payload[4]})
			total := binary.BigEndian.Uint16([]byte{0x00, payload[5]})

			// In the first packet, the payload is: the original MQTT topic, the byte 0x01, and the MQTT payload.
			// In the second and later packets, the payload is just the remaining bytes in the MQTT payload.
			packetPayload := payload[6:]
			for bytes.HasSuffix(packetPayload, []byte{0x00}) {
				packetPayload = bytes.TrimSuffix(packetPayload, []byte{0x00})
			}

			// If this is the first packet, the payload has the topic prepended to it.
			packetTopic := ""
			if number == 1 {
				if bytes.Contains(packetPayload, []byte{0x01}) {
					parts := bytes.Split(packetPayload, []byte{0x01})
					packetTopic = string(parts[0])
					packetPayload = parts[1]
				} else {
					logrus.Warn("[mqtt] first mesh packet does not have separator")
				}
			}

			logrus.Tracef("[mqtt] got packet %s (%d/%d): %s", correlation, number, total, packetPayload)

			// Store the packet
			meshPacketsLock.Lock()

			packets := meshPackets[correlation]
			packets = append(packets, meshPacket{
				ArrivalTime: time.Now(),
				Number:      number,
				Total:       total,
				Topic:       packetTopic,
				Payload:     packetPayload,
			})
			meshPackets[correlation] = packets

			meshPacketsLock.Unlock()

			first := packets[0]
			if len(packets) != int(first.Total) {
				return
			}

			// Once all parts of the packet have been received, reassemble and handle it.
			sort.Slice(packets, func(i, j int) bool {
				return packets[i].Number < packets[j].Number
			})

			// MQTT topics are one of: garden/module/XXXXXXXXXX/tele/data OR garden/module/discovery/XXXXXXXXXX
			clientId := ""
			parts := strings.Split(first.Topic, "/")
			if strings.Contains(first.Topic, "/tele/") {
				logrus.Tracef("[mqtt] mesh client id is a telemetry packet")
				clientId = parts[2]
			} else {
				logrus.Tracef("[mqtt] mesh client id is a discovery packet")
				clientId = parts[len(parts)-1]
			}

			// Reassemble the payload and handle it
			var meshPayload []byte
			var expectedNumber uint16 = 0
			handle := true
			for _, packet := range packets {
				if expectedNumber++; expectedNumber != packet.Number {
					logrus.Warnf("[mqtt] encountered unexpected packet while reassembling packet chunks (expected %d, got %d",
						expectedNumber,
						packet.Number)

					handle = false
					break
				}

				meshPayload = append(meshPayload, packet.Payload...)
			}

			if handle {
				handleMqttMessage(clientId, first.Topic, meshPayload)
			}

		} else {
			logrus.Warnf("[mqtt] unhandled MQTT topic: %s", topic)
			return
		}
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
