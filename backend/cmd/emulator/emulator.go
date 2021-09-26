package main

import (
	"fmt"
	"strings"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"

	"github.com/ConfusedPolarBear/garden/internal/config"
	"github.com/ConfusedPolarBear/garden/internal/mqtt"
	"github.com/sirupsen/logrus"
)

// Emulates a garden sensor system for testing without physical hardware.

var baseTopic string = "garden/module/656d75"
var publishDelay float32 = 0.1

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	logrus.SetLevel(logrus.DebugLevel)

	config.Load()
	mqtt.Setup(false)

	discovery := `{"System":{"RestartReason":"External System","CoreVersion":"0.0.0","SdkVersion":"2.2.2-dev(38a443e)",` +
		`"FlashSize":4194304,"RealFlashSize":4194304},"Sensors":["temperature","humidity"]}`

	mqtt.Subscribe(baseTopic+"/cmnd/#", func(_ paho.Client, m paho.Message) {
		command := m.Topic()
		payload := string(m.Payload())

		lastSlash := strings.LastIndex(command, "/") + 1
		command = command[lastSlash:]

		logrus.Debugf("[mqtt] got command %s with payload %s", command, payload)
	})

	mqtt.PublishAdvanced("garden/module/discovery/656d75", discovery, 0, true)

	temp, humidity := 0, 0
	for {
		if temp += 2; temp >= 45 {
			temp = 0
		}

		if humidity += 3; humidity >= 100 {
			humidity = 0
		}

		payload := fmt.Sprintf(`{"Temperature":%d,"Humidity":%d}`, temp, humidity)
		mqtt.Publish(baseTopic+"/tele/data", payload)

		time.Sleep(time.Duration(publishDelay) * time.Second)
	}
}
