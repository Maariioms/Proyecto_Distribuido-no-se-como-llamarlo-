package controller

import (
	"context"
	"log"
	"math/rand"
	"time"

	"soc-mvp-distribuido/events/internal/repository"
)

var eventTypes = []string{
	"trafico_saliente_anomalo",
	"login_ip_no_reconocida",
	"escaneo_puertos_interno",
	"consulta_dns_sospechosa",
	"proceso_no_firmado",
}

var severities = []string{"low", "medium", "high", "critical"}

// StartGenerator lanza UNA goroutine que corre en paralelo al servidor
// HTTP, insertando un evento simulado cada `interval`. Se llama una sola
// vez en main() — la goroutine sigue viva mientras el programa corre.
func StartGenerator(repo repository.EventRepository, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			eventType := eventTypes[rand.Intn(len(eventTypes))]
			severity := severities[rand.Intn(len(severities))]

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_, err := repo.Create(ctx, eventType, severity)
			cancel()

			if err != nil {
				log.Printf("generator: could not create simulated event: %v", err)
				continue
			}
			log.Printf("generator: created simulated event (%s, %s)", eventType, severity)
		}
	}()
}
