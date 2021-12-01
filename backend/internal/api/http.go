package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ConfusedPolarBear/garden/internal/db"
	"github.com/ConfusedPolarBear/garden/internal/mqtt"
	"github.com/ConfusedPolarBear/garden/internal/util"
	"github.com/ConfusedPolarBear/garden/internal/websocket"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	bind := "0.0.0.0:8081"

	r := mux.NewRouter()
	r.Use(corsMiddleware)

	r.HandleFunc("/ping", PingHandler).Methods("GET")

	r.HandleFunc("/systems", GetSystems).Methods("GET")
	r.HandleFunc("/system/{id}", GetSystem).Methods("GET")
	r.HandleFunc("/system/delete/{id}", DeleteSystem).Methods("POST")
	r.HandleFunc("/system/command/{id}", SendCommand).Methods("POST", "OPTIONS")

	r.HandleFunc("/socket", websocket.WebSocketHandler)

	logrus.Printf("[server] API server listening on http://%s", bind)
	if err := http.ListenAndServe(bind, r); err != nil {
		panic(err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// TODO: add authorization
		// w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		next.ServeHTTP(w, r)
	})
}

func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func GetSystems(w http.ResponseWriter, r *http.Request) {
	w.Write(util.Marshal(db.GetAllSystems()))
}

func GetSystem(w http.ResponseWriter, r *http.Request) {
	id, err := getId(w, r)
	if err != nil {
		return
	}

	system, err := db.GetSystem(id, true)
	if err != nil {
		logrus.Warnf("[api] error getting system %s: %s", id, err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Write(util.Marshal(system))
}

func DeleteSystem(w http.ResponseWriter, r *http.Request) {
	id, err := getId(w, r)
	if err != nil {
		return
	}

	if err := db.DeleteSystem(id); err != nil {
		logrus.Warnf("[server] unable to delete system %s: %s", id, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func SendCommand(w http.ResponseWriter, r *http.Request) {
	// Command sent to a coordinator to publish an arbitrary mesh message.
	type meshPublishCommand struct {
		// Command as seen by the coordinator. Must be "Publish".
		Command string

		// Payload to publish. Must be deserializable as JSON.
		Payload string
	}

	id, err := getId(w, r)
	if err != nil {
		return
	}

	// Parse the form and extract the command
	if err := r.ParseForm(); err != nil {
		logrus.Warnf("[server] unable to parse form: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	command := r.Form.Get("command")
	if command == "" || len(command) > 210 {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}

	// Get the system
	isMesh := false
	if id != "FFFFFFFFFFFF" {
		// If this is not a broadcast message, lookup the individual system to send the message to
		system, err := db.GetSystem(id, false)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		isMesh = system.Announcement.IsMesh
	} else {
		isMesh = true
	}

	logrus.Debugf("[server] sending command to %s: %s", id, command)

	mqttDest := id
	mqttPayload := command

	// If this system is connected over MQTT, send the raw command
	if isMesh {
		// Mesh connected systems are controlled by sending a command (MQTT) to the coordinator who will rebroadcast it (ESP-NOW)
		logrus.Debugf("[server] system %s networking mode is mesh", id)

		// Lookup the coordinator for this system.
		coordinator, err := db.GetCoordinator()
		if err != nil {
			logrus.Errorf("[server] unable to find coordinator: %s", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		mqttDest = coordinator.Identifier

		// Construct the mesh payload
		// {"Command":"Publish", "Payload": "{'D':'dst-123457890AB','Command':'Ping'}"}

		// Unmarshal the command in order to add the destination key to it
		var rawCommand map[string]interface{}
		if err := json.Unmarshal([]byte(command), &rawCommand); err != nil {
			logrus.Errorf("[server] unable to unmarshal command as JSON: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		rawCommand["D"] = fmt.Sprintf("dst-%s", id)

		// Remarshal the command in coordinator format.
		meshCommand := util.Marshal(meshPublishCommand{
			Command: "Publish",
			Payload: string(util.Marshal(rawCommand)),
		})

		mqttPayload = string(meshCommand)

	} else {
		logrus.Debug("[server] system %s networking mode is wifi")
	}

	logrus.Debugf("[server] commanding \"%s\"", mqttDest)
	logrus.Debugf("[server] mqtt payload is \"%s\"", mqttPayload)

	mqtt.Publish(fmt.Sprintf("garden/module/%s/cmnd", mqttDest), mqttPayload)
}
