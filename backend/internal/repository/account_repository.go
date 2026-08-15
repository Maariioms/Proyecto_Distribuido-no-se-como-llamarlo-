package repository

import (
	"context"
	"errors"

	"soc-mvp-distribuido/backend/internal/model"
)

// ErrNotFound indica que no existe un registro con el criterio solicitado.
var ErrNotFound = errors.New("resource not found")

// ErrInvalidCredentials indica username inexistente o password incorrecto.
var ErrInvalidCredentials = errors.New("invalid credentials")

// AccountRepository define el contrato de acceso a datos para cuentas.
// El controller depende de esta interfaz, no de la implementación concreta.
type AccountRepository interface {
	Create(ctx context.Context, username, password string) (model.Account, error)
	Authenticate(ctx context.Context, username, password string) (model.Account, error)
}
