package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type GardenSystem struct {
	Identifier  string
	LastReading Reading
	LastSeen    time.Time
}

type Reading struct {
	Temperature float32
	Humidity    float32
}

var systems map[string]GardenSystem
var systemLock sync.Mutex

func main() {
	systems = map[string]GardenSystem{}

	host := os.Getenv("MQTT_HOST")
	username := os.Getenv("MQTT_USERNAME")
	password := os.Getenv("MQTT_PASSWORD")

	if host == "" || username == "" || password == "" {
		panic("host, username, or password is invalid")
	}

	opts := mqtt.NewClientOptions().
		AddBroker(fmt.Sprintf("tcp://%s", host)).
		SetClientID("garden-backend").
		SetConnectTimeout(5 * time.Second).
		SetOrderMatters(false).
		SetUsername(username).
		SetPassword(password).
		SetKeepAlive(10 * time.Second).
		SetPingTimeout(2 * time.Second)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	systemRe := regexp.MustCompile("^garden/module/([a-f0-9]+)/")
	mustSubscribe(client, "garden/module/#", func(c mqtt.Client, m mqtt.Message) {
		topic := m.Topic()
		payload := string(m.Payload())

		client := "SYSTEM"
		if systemRe.MatchString(topic) {
			client = systemRe.FindStringSubmatch(topic)[1]
		}

		fmt.Printf("message from %s: %s: %s\n", client, topic, payload)

		if client == "SYSTEM" {
			systemLock.Lock()

			// discovery message is garden/module/discovery/deadbeef
			p := strings.Split(topic, "/")
			id := p[len(p)-1]

			if _, okay := systems[id]; !okay {
				systems[id] = GardenSystem{
					Identifier: id,
					LastSeen:   time.Time{},
				}
			}

			systemLock.Unlock()

			return
		}

		systemLock.Lock()
		system := systems[client]
		system.LastSeen = time.Now()

		if err := json.Unmarshal([]byte(payload), &system.LastReading); err != nil {
			fmt.Printf("unable to unmarshal reading: %s\n", err)
			systemLock.Unlock()
			return
		}

		systems[client] = system
		systemLock.Unlock()
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmp, err := template.ParseFiles("index.html")
		if err != nil {
			panic(err)
		}

		systemLock.Lock()
		defer systemLock.Unlock()

		type htmlData struct {
			Systems map[string]GardenSystem
		}

		if err := tmp.Execute(w, htmlData{Systems: systems}); err != nil {
			panic(err)
		}
	})

	bind := "127.0.0.1:8080"
	fmt.Printf("listening on %s\n", bind)
	if err := http.ListenAndServe(bind, nil); err != nil {
		panic(err)
	}
}

func mustSubscribe(client mqtt.Client, topic string, callback func(c mqtt.Client, m mqtt.Message)) {
	if token := client.Subscribe(topic, 0, callback); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}
