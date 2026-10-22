package controller

import (
	"encoding/json"
	"net/http"
	"os"
	"time"
)

// anota una consulta en el Tracker y regresa como va todo
// la usan el http y el grpc pa no repetir la logica
func CheckAndSnapshot() (uptime time.Duration, checks int, err error) {
	if err := instance.RegisterCheck(); err != nil {
		return 0, 0, err
	}
	uptime, checks = instance.Snapshot()
	return uptime, checks, nil
}

func Status(w http.ResponseWriter, r *http.Request) {
	uptime, checks, err := CheckAndSnapshot()
	if err != nil {
		http.Error(w, "could not register check", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":      "ok",
		"service_id":  os.Getenv("SERVICE_ID"),
		"uptime":      uptime.String(),
		"checks_done": checks,
	})
}

// la ruta que revisa consul cada 5s
// va aparte de Status a proposito pq Status cuenta cada llamada en checks_done y escribe en sqlite
// y no queremos que el health check de consul infle ese contador
func Heartbeat(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
