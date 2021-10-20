package api

import (
	"net/http"
	"gorm.io/gorm"
	"github.com/ConfusedPolarBear/garden/internal/util"
	"github.com/ConfusedPolarBear/garden/dbconn"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"fmt"
)

var DB *gorm.DB

func StartServer(db *gorm.DB) {
	DB = db
	bind := "0.0.0.0:8081"

	r := mux.NewRouter()
	r.Use(corsMiddleware)

	r.HandleFunc("/ping", PingHandler)

	r.HandleFunc("/systems", GetSystems)

	r.HandleFunc("/socket", WebSocketHandler)

	r.HandleFunc("/testdb", TestDB)

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
	fmt.Fprintf(w, "hello\n")
	w.WriteHeader(http.StatusNoContent)
}

func GetSystems(w http.ResponseWriter, r *http.Request) {
	w.Write(util.Marshal(util.SystemMapToSlice()))
}

func TestDB(w http.ResponseWriter, r *http.Request) {
	// testReading := util.Reading{
	// 	Temperature: 123,
	// 	Humidity: 456,
	// }
	// if err := dbconn.CreateReading(DB, testReading); err != nil {
	// 	panic(err)
	// }
	// reading := &util.Reading{
	// 	Temperature: 123,
	// }

	// if err := DB.Where(reading).First(reading).Error; err != nil {
	// 	panic(err)
	// }

	// fmt.Printf("%+v\n",reading)

	testGardenSystem := util.GardenSystemInfo{
		Identifier: "TEST69",
		RestartReason: "TEST456",
		CoreVersion: "TEST789",
		SdkVersion:  "TEST123",
		FlashSize: 987987987,
		RealFlashSize: 3942,
	}

	if err := dbconn.CreateGardenSystem(DB, testGardenSystem); err != nil {
		panic(err)
	}

	retrievedStruct, err := dbconn.GetGardenSystem(DB, "TEST69")
	if err != nil {
		panic(err)
	}
	fmt.Println("Retrieved struct:")
	fmt.Printf("%+v\n",retrievedStruct)
}
