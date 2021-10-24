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

	r.HandleFunc("/ping", PingHandler)

	r.HandleFunc("/systems", GetSystems)

	r.HandleFunc("/socket", WebSocketHandler)

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
