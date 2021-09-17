package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func SetupAPIServer() {
	bind := "127.0.0.1:8081"

	r := mux.NewRouter()
	r.Use(corsMiddleware)

	r.HandleFunc("/systems", GetSystems)
	r.HandleFunc("/ping", PingHandler)

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
	var clients []GardenSystem

	systemLock.Lock()
	defer systemLock.Unlock()

	for _, c := range systems {
		clients = append(clients, c)
	}

	w.Write(Marshal(clients))
}

// Marshal v or panic.
func Marshal(v interface{}) []byte {
	if data, err := json.Marshal(v); err != nil {
		panic(err)
	} else {
		return data
	}
}
