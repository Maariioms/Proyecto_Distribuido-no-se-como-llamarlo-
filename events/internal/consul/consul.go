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

// aqui el servicio se anota en consul y se borra cuando se apaga
// todo viene de variables de entorno que pone docker-compose
//
//	CONSUL_HTTP_ADDR -> donde esta consul y el default es http://consul:8500
//	SERVICE_NAME     -> nombre en consul, ej accounts events status
//	SERVICE_ID       -> id de ESTA instancia, ej accounts-1
//	SERVICE_ADDRESS  -> hostname pa que consul nos encuentre y es el nombre en compose
//	PORT             -> puerto http y el default es 8080
//	GRPC_PORT        -> puerto grpc y el default es 50051

var (
	consulAddr string
	serviceID  string
	client     = http.Client{Timeout: 3 * time.Second}
)

// saca del entorno lo que se repite en Register y RegisterGRPC
func leerConfig() (name, address string, err error) {
	consulAddr = getEnv("CONSUL_HTTP_ADDR", "http://consul:8500")
	serviceID = os.Getenv("SERVICE_ID")
	name = os.Getenv("SERVICE_NAME")
	address = os.Getenv("SERVICE_ADDRESS")

	if serviceID == "" || name == "" || address == "" {
		return "", "", fmt.Errorf("SERVICE_ID, SERVICE_NAME y SERVICE_ADDRESS son obligatorios para registrarse en Consul")
	}
	return name, address, nil
}

// le dice a consul "aqui toy" por http
// consul le pega a healthPath cada 5s y si falla 1 min seguido nos borra solo
func Register(healthPath string) error {
	name, address, err := leerConfig()
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(getEnv("PORT", "8080"))
	if err != nil {
		return fmt.Errorf("PORT invalido: %w", err)
	}

	check := map[string]any{
		"HTTP":                           fmt.Sprintf("http://%s:%d%s", address, port, healthPath),
		"Interval":                       "5s",
		"Timeout":                        "2s",
		"DeregisterCriticalServiceAfter": "1m",
	}
	return register(serviceID, name, address, port, check)
}

// lo mismo pero pa el puerto grpc
// se anota como otro servicio con el nombre <nombre>-grpc, ej events-grpc
// asi el middleware busca aparte los http y los grpc
// el check es tipo GRPC y consul le pregunta al health de grpc, el grpc.health.v1
// que cada backend ya trae prendido
func RegisterGRPC() error {
	name, address, err := leerConfig()
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(getEnv("GRPC_PORT", "50051"))
	if err != nil {
		return fmt.Errorf("GRPC_PORT invalido: %w", err)
	}

	check := map[string]any{
		"GRPC":                           fmt.Sprintf("%s:%d", address, port),
		"GRPCUseTLS":                     false,
		"Interval":                       "5s",
		"Timeout":                        "2s",
		"DeregisterCriticalServiceAfter": "1m",
	}
	return register(serviceID+"-grpc", name+"-grpc", address, port, check)
}

// el PUT a consul que usan Register y RegisterGRPC
// reintenta pq consul puede estar arrancando todavia
func register(id, name, address string, port int, check map[string]any) error {
	payload := map[string]any{
		"ID":      id,
		"Name":    name,
		"Address": address,
		"Port":    port,
		"Check":   check,
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
				log.Printf("registrado en Consul: %s (%s:%d)", id, address, port)
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
	return fmt.Errorf("no se pudo registrar %s en Consul: %w", id, lastErr)
}

// se borra de consul tanto el http como el grpc
// pa que nadie le mande trafico mientras se apaga
func Deregister() {
	if serviceID == "" {
		return
	}

	for _, id := range []string{serviceID, serviceID + "-grpc"} {
		req, err := http.NewRequest(http.MethodPut, consulAddr+"/v1/agent/service/deregister/"+id, nil)
		if err != nil {
			log.Printf("no se pudo armar la baja de %s en Consul: %v", id, err)
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("no se pudo dar de baja %s en Consul: %v", id, err)
			continue
		}
		resp.Body.Close()
		log.Printf("dado de baja en Consul: %s", id)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
