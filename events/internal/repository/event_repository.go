package repository

import (
	"context"
	"errors"

	"soc-mvp-distribuido/events/internal/model"
)

// ErrNotFound indica que no existe un evento con el ID solicitado.
var ErrNotFound = errors.New("event not found")

// EventRepository define el contrato de acceso a datos para eventos.
// El controller depende de esta interfaz, no de la implementación concreta.
type EventRepository interface {
	Create(ctx context.Context, eventType, severity string) (model.Event, error)
	List(ctx context.Context) ([]model.Event, error)
	Get(ctx context.Context, id uint) (model.Event, error)
}
