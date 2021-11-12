package api

import (
	"net/http"

	"github.com/ConfusedPolarBear/garden/internal/db"
	"github.com/ConfusedPolarBear/garden/internal/util"

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

	r.HandleFunc("/firmware/manifest.json", ManifestHandler).Methods("GET")
	r.HandleFunc("/firmware/{board}/{file}", DownloadFirmware).Methods("GET")

	r.HandleFunc("/socket", WebSocketHandler)

	checkFirmwareManifest()

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
