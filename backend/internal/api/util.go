package api

import (
	"errors"
	"net/http"

	"github.com/ConfusedPolarBear/garden/internal/util"

	"github.com/gorilla/mux"
)

func getId(w http.ResponseWriter, r *http.Request) (string, error) {
	id := mux.Vars(r)["id"]

	if !util.SystemIdentifierRegex.MatchString(id) {
		w.WriteHeader(http.StatusBadRequest)
		return "", errors.New("invalid id")
	}

	return id, nil
}
