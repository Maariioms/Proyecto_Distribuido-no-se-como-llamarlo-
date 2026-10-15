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

// Heartbeat es la ruta que revisa Consul cada 5s. Va aparte de Status a
// proposito: Status cuenta cada llamada en checks_done y escribe en SQLite,
// y no queremos que el health check de Consul infle ese contador.
func Heartbeat(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
