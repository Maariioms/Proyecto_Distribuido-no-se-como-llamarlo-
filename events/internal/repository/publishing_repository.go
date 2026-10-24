package repository

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"

	"soc-mvp-distribuido/events/internal/kafkax"
	"soc-mvp-distribuido/events/internal/model"
)

// envuelve a otro EventRepository y despues de guardar un evento lo publica en events.created
// asi TODO lo que cree eventos avisa solo -> el generador, el http, el grpc y el consumidor
// y ninguno de ellos sabe que kafka existe
type PublishingRepository struct {
	// embeber la interfaz hace que List y Get pasen directo al repo de adentro
	// solo reescribimos Create
	EventRepository
	writer *kafka.Writer
}

var _ EventRepository = (*PublishingRepository)(nil)

func NewPublishingRepository(inner EventRepository, writer *kafka.Writer) *PublishingRepository {
	return &PublishingRepository{EventRepository: inner, writer: writer}
}

func (r *PublishingRepository) Create(ctx context.Context, eventType, severity string) (model.Event, error) {
	event, err := r.EventRepository.Create(ctx, eventType, severity)
	if err != nil {
		return event, err
	}

	// ya esta guardado asi que si kafka falla solo se avisa y NO se regresa error
	// perder un aviso para status es menos grave que decir que no se guardo el evento
	body, _ := json.Marshal(kafkax.EventMessage{Type: event.Type, Severity: event.Severity, At: event.CreatedAt})

	pubCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := r.writer.WriteMessages(pubCtx, kafka.Message{Key: []byte(event.Type), Value: body}); err != nil {
		log.Printf("no se pudo publicar en %s: %v", kafkax.TopicEventsCreated, err)
	}
	return event, nil
}
