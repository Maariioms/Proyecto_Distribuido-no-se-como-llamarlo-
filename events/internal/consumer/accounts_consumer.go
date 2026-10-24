package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"

	"soc-mvp-distribuido/events/internal/kafkax"
	"soc-mvp-distribuido/events/internal/repository"
)

// lee el topico accounts.activity y guarda cada mensaje como evento en postgres
// es el mismo trabajo que hacia CreateEvent por grpc pero ahora llega por la cola
//
// repo es el PublishingRepository asi que al guardar tambien avisa en events.created
// accounts.activity -> events -> events.created -> status
func StartAccountsConsumer(ctx context.Context, brokers string, repo repository.EventRepository) {
	list := strings.Split(brokers, ",")

	// los topicos primero pa que existan antes de leer o escribir
	for _, topic := range []string{kafkax.TopicAccountsActivity, kafkax.TopicAccountsActivityDLQ, kafkax.TopicEventsCreated} {
		if err := kafkax.EnsureTopic(list[0], topic, kafkax.Partitions); err != nil {
			log.Printf("WARNING: no se pudo crear el topico %s: %v", topic, err)
		}
	}

	dlq := kafkax.NewWriter(list, kafkax.TopicAccountsActivityDLQ)

	handle := func(ctx context.Context, msg kafka.Message) error {
		var m kafkax.EventMessage
		if err := json.Unmarshal(msg.Value, &m); err != nil {
			return fmt.Errorf("mensaje invalido: %w", err)
		}
		if m.Type == "" || m.Severity == "" {
			return fmt.Errorf("faltan type o severity")
		}

		if _, err := repo.Create(ctx, m.Type, m.Severity); err != nil {
			return err
		}
		log.Printf("kafka: %s partition=%d offset=%d -> evento %s guardado", kafkax.TopicAccountsActivity, msg.Partition, msg.Offset, m.Type)
		return nil
	}

	// corre en su propia goroutine pa no bloquear el servidor http
	// Consume regresa error si kafka se cae, asi que se vuelve a entrar despues de 3s
	go func() {
		for ctx.Err() == nil {
			err := kafkax.Consume(ctx, list, kafkax.TopicAccountsActivity, kafkax.GroupEventsWriter, handle, dlq)
			if err != nil {
				log.Printf("kafka: consumidor de accounts se cayo, reintento: %v", err)
			}
			select {
			case <-ctx.Done():
			case <-time.After(3 * time.Second):
			}
		}
	}()
}
