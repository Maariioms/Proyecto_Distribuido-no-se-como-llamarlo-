package controller

import (
	"net/http"
	"os"
)

func Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":     "ok",
		"service_id": os.Getenv("SERVICE_ID"),
	})
}
