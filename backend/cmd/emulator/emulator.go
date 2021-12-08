package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/ConfusedPolarBear/garden/internal/config"
	"github.com/ConfusedPolarBear/garden/internal/mqtt"
	"github.com/ConfusedPolarBear/garden/internal/util"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

var id string
var baseTopic string
var publishDelay float32 = 10

// If all current system discovery messages should be removed.
var flagClearSystems bool

func init() {
	// Setup logging
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	logrus.SetLevel(logrus.DebugLevel)

	// Parse CLI flags
	flag.BoolVar(&flagClearSystems, "c", false, "If all garden discovery messages should be cleared")
	flag.StringVar(&id, "i", "1234567890AB", "Sets the 12 character identifier for this system. Only 0-9 and A-F are permitted.")

	flag.Parse()

	if !util.SystemIdentifierRegex.MatchString(id) {
		panic(fmt.Sprintf("provided identifier must match %s", &util.SystemIdentifierRegex))
	}

	baseTopic = "garden/module/" + id
}

func main() {
	config.Load()
	mqtt.Setup(false)

	mqtt.Subscribe("garden/module/discovery/+", parseDiscoveryMessage)
	mqtt.Subscribe(baseTopic+"/cmnd/#", handleCommand)

	discovery := `{"RR":"External System","CV":"0.0.0","SV":"2.2.2-dev(38a443e)",` +
		`"IsEmulator":true,"Sensors":["temperature","humidity"]}`

	mqtt.PublishAdvanced("garden/module/discovery/"+id, discovery, 0, true)

	temp, humidity := -10, 0
	for {
		if temp += 2; temp >= 45 {
			temp = -10
		}

		if humidity += 3; humidity >= 100 {
			humidity = 0
		}

		payload := fmt.Sprintf(`{"Error":false,"Temperature":%d,"Humidity":%d}`, temp, humidity)
		mqtt.Publish(baseTopic+"/tele/data", payload)

		time.Sleep(time.Duration(publishDelay) * time.Second)
	}
}

func parseDiscoveryMessage(c paho.Client, m paho.Message) {
	if len(m.Payload()) == 0 {
		return
	}

	var system util.GardenSystemInfo

	if err := json.Unmarshal(m.Payload(), &system); err != nil {
		logrus.Warnf("[discovery] failed to unmarshal message from %s: %s", m.Topic(), err)
		return
	}

	if !flagClearSystems {
		return
	}

	mqtt.PublishAdvanced(m.Topic(), "", 0, true)
}

func handleCommand(_ paho.Client, m paho.Message) {
	command := m.Topic()
	payload := string(m.Payload())

	command = getLastSlash(command)

	logrus.Debugf("[mqtt] got command %s with payload %s", command, payload)
}

// Returns the last item in a slash separated string. Example: "a/b/c/d" will return "d".
func getLastSlash(raw string) string {
	if !strings.Contains(raw, "/") {
		panic(fmt.Errorf("input \"%s\" does not contain any slashes", raw))
	}

	parts := strings.Split(raw, "/")
	return parts[len(parts)-1]
}
