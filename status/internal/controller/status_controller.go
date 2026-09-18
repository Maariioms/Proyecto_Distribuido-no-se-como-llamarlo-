package controller

import (
	"encoding/json"
	"net/http"
	"os"
)

func Status(w http.ResponseWriter, r *http.Request) {
	if err := instance.RegisterCheck(); err != nil {
		http.Error(w, "could not register check", http.StatusInternalServerError)
		return
	}

	uptime, checks := instance.Snapshot()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":      "ok",
		"service_id":  os.Getenv("SERVICE_ID"),
		"uptime":      uptime.String(),
		"checks_done": checks,
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
