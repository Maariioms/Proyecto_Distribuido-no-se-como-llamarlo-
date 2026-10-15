package consul

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Claude explanation para que entendiera bien como funciona xd.

// Este paquete registra el servicio en Consul y lo da de baja al apagarse.
// Toda la config viene de variables de entorno (las pone docker-compose):
//
//	CONSUL_HTTP_ADDR  donde esta Consul        (default http://consul:8500)
//	SERVICE_NAME      nombre en Consul         (accounts, events, status)
//	SERVICE_ID        id unico de la instancia (accounts-1)
//	SERVICE_ADDRESS   hostname donde Consul nos encuentra (nombre en compose)
//	PORT              puerto donde escuchamos  (default 8080)

var (
	consulAddr string
	serviceID  string
	client     = http.Client{Timeout: 3 * time.Second}
)

// Register le dice a Consul "aqui estoy" y le pide que haga un health check
// HTTP a healthPath cada 5s. Si falla 1 minuto seguido, Consul nos borra solo.
// Reintenta porque Consul puede estar todavia arrancando.
func Register(healthPath string) error {
	consulAddr = getEnv("CONSUL_HTTP_ADDR", "http://consul:8500")
	serviceID = os.Getenv("SERVICE_ID")
	name := os.Getenv("SERVICE_NAME")
	address := os.Getenv("SERVICE_ADDRESS")

	if serviceID == "" || name == "" || address == "" {
		return fmt.Errorf("SERVICE_ID, SERVICE_NAME y SERVICE_ADDRESS son obligatorios para registrarse en Consul")
	}

	port, err := strconv.Atoi(getEnv("PORT", "8080"))
	if err != nil {
		return fmt.Errorf("PORT invalido: %w", err)
	}

	payload := map[string]any{
		"ID":      serviceID,
		"Name":    name,
		"Address": address,
		"Port":    port,
		"Check": map[string]any{
			"HTTP":                           fmt.Sprintf("http://%s:%d%s", address, port, healthPath),
			"Interval":                       "5s",
			"Timeout":                        "2s",
			"DeregisterCriticalServiceAfter": "1m",
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	const maxAttempts = 10
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := http.NewRequest(http.MethodPut, consulAddr+"/v1/agent/service/register", bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				log.Printf("registrado en Consul: %s (%s:%d)", serviceID, address, port)
				return nil
			}
			lastErr = fmt.Errorf("Consul respondio %d", resp.StatusCode)
		} else {
			lastErr = err
		}

		if attempt < maxAttempts {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}
	return fmt.Errorf("no se pudo registrar en Consul: %w", lastErr)
}

// Deregister saca esta instancia del registry para que nadie le mande
// trafico mientras se apaga.
func Deregister() {
	if serviceID == "" {
		return
	}

	req, err := http.NewRequest(http.MethodPut, consulAddr+"/v1/agent/service/deregister/"+serviceID, nil)
	if err != nil {
		log.Printf("no se pudo armar la baja en Consul: %v", err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("no se pudo dar de baja %s en Consul: %v", serviceID, err)
		return
	}
	resp.Body.Close()
	log.Printf("dado de baja en Consul: %s", serviceID)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
