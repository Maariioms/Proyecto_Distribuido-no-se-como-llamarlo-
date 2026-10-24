package events

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"

	"soc-mvp-distribuido/backend/internal/kafkax"
)

// le avisa a events publicando en el topico accounts.activity
// accounts ya no sabe quien lee ni si esta vivo, solo deja el mensaje en la cola
// events lo procesa cuando pueda y por eso si events esta caido no se pierde nada
type KafkaPublisher struct {
	writer *kafka.Writer
}

var _ Publisher = (*KafkaPublisher)(nil)

// brokers viene como "kafka:9092" o varios separados por coma
func NewKafkaPublisher(brokers string) (*KafkaPublisher, error) {
	list := strings.Split(brokers, ",")

	// crea el topico si no existe, si falla solo avisa pq el broker lo crea al primer mensaje
	if err := kafkax.EnsureTopic(list[0], kafkax.TopicAccountsActivity, kafkax.Partitions); err != nil {
		log.Printf("WARNING: no se pudo crear el topico %s: %v", kafkax.TopicAccountsActivity, err)
	}

	return &KafkaPublisher{writer: kafkax.NewWriter(list, kafkax.TopicAccountsActivity)}, nil
}

func (p *KafkaPublisher) Publish(_ context.Context, eventType, severity string) {
	// goroutine y ctx propio por lo mismo que el GRPCPublisher
	// el login no espera a kafka y el ctx del request se cancela apenas se responde
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		body, _ := json.Marshal(kafkax.EventMessage{Type: eventType, Severity: severity, At: time.Now()})

		err := p.writer.WriteMessages(ctx, kafka.Message{
			Key:   []byte(eventType), // mismo tipo -> misma particion -> orden
			Value: body,
		})
		if err != nil {
			log.Printf("no se pudo publicar en kafka (%s): %v", eventType, err)
		}
	}()
}
