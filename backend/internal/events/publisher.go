package events

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	eventsv1 "soc-mvp-distribuido/backend/gen/eventspb/events/v1"
)

// lo que accounts necesita pa avisar "paso esto" sin saber COMO se avisa
// hoy es por grpc y luego puede ser kafka y accounts ni se entera
// misma idea que AccountRepository pero pa mandar eventos en vez de guardar datos
type Publisher interface {
	// "best effort" -> no regresa error
	// si events se cae el login NO debe fallar asi que solo se loguea
	Publish(ctx context.Context, eventType, severity string)
}

// no hace nada
// se usa cuando no hay EVENTS_GRPC_ADDR, ej cuando corres accounts solo
type Noop struct{}

func (Noop) Publish(ctx context.Context, eventType, severity string) {}

var _ Publisher = Noop{}

// le avisa a events por grpc llamando a CreateEvent
type GRPCPublisher struct {
	client eventsv1.EventServiceClient
}

var _ Publisher = (*GRPCPublisher)(nil)

func NewGRPCPublisher(addr string) (*GRPCPublisher, error) {
	// NewClient no conecta de una vez y conecta en la primera llamada
	// insecure -> sin TLS pq es red interna de docker
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &GRPCPublisher{client: eventsv1.NewEventServiceClient(conn)}, nil
}

func (p *GRPCPublisher) Publish(_ context.Context, eventType, severity string) {
	// va en goroutine pa que el login no espere a events
	// no usa el ctx del request pq ese se cancela apenas se responde
	// y la goroutine seguiria corriendo
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		_, err := p.client.CreateEvent(ctx, &eventsv1.CreateEventRequest{
			Type:     eventType,
			Severity: severity,
		})
		if err != nil {
			log.Printf("no se pudo avisar a events (%s): %v", eventType, err)
		}
	}()
}
