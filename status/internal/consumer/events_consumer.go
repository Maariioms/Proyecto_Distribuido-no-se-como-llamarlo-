package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"

	"soc-mvp-distribuido/status/internal/controller"
	"soc-mvp-distribuido/status/internal/kafkax"
)

// lee el topico events.created y por cada evento sube el contador del Tracker
// es el ultimo eslabon -> accounts.activity -> events -> events.created -> status
// y se ve en el campo events_seen del /health de status
func StartEventsConsumer(ctx context.Context, brokers string) {
	list := strings.Split(brokers, ",")

	if err := kafkax.EnsureTopic(list[0], kafkax.TopicEventsCreated, kafkax.Partitions); err != nil {
		log.Printf("WARNING: no se pudo crear el topico %s: %v", kafkax.TopicEventsCreated, err)
	}

	handle := func(_ context.Context, msg kafka.Message) error {
		var m kafkax.EventMessage
		if err := json.Unmarshal(msg.Value, &m); err != nil {
			return fmt.Errorf("mensaje invalido: %w", err)
		}
		controller.RegisterEvent(m.Severity)
		return nil
	}

	// su propia goroutine y si Consume falla se vuelve a entrar despues de 3s
	go func() {
		for ctx.Err() == nil {
			err := kafkax.Consume(ctx, list, kafkax.TopicEventsCreated, kafkax.GroupStatusCounter, handle, nil)
			if err != nil {
				log.Printf("kafka: consumidor de events.created se cayo, reintento: %v", err)
			}
			select {
			case <-ctx.Done():
			case <-time.After(3 * time.Second):
			}
		}
	}()
}
