package api

import (
	"net/http"

	"github.com/ConfusedPolarBear/garden/internal/db"
	"github.com/ConfusedPolarBear/garden/internal/util"

	"github.com/sirupsen/logrus"
)

func MeshInfoHandler(w http.ResponseWriter, _ *http.Request) {
	type meshInfo struct {
		Key        string
		Controller string
		Channel    int
	}

	config, err := db.GetConfiguration()
	if err != nil {
		logrus.Errorf("[server] no configuration information found")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// TODO: support multiple coordinators & allow the user to pick which one they want
	coordinator, err := db.GetCoordinator()
	if err != nil {
		logrus.Errorf("[server] no coordinators defined")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	addr := util.IdentifierToAddress(coordinator.Identifier)

	info := meshInfo{Key: config.MeshKey, Controller: addr, Channel: coordinator.Announcement.Channel}

	w.Write(util.Marshal(info))
}
