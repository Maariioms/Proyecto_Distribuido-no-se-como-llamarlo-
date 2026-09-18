package controller

import (
	"database/sql"
	"sync"
	"time"
)

// Tracker es el Singleton: existe UNA sola instancia en todo el programa,
// compartida por todos los handlers que la usan. Guarda cuántos health
// checks se han hecho y desde cuándo está corriendo el servicio.
type Tracker struct {
	mu        sync.Mutex
	db        *sql.DB
	startedAt time.Time
	checks    int
}

// instance es la única instancia del Tracker. Es una variable a nivel de
// paquete (nunca se crea una segunda), no un struct con "new" por fuera.
var instance *Tracker

// InitTracker crea la única instancia y prepara la tabla en SQLite.
// Se llama UNA vez, en main(), igual que newBalancer() en el middleware.
func InitTracker(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS checks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		checked_at DATETIME NOT NULL
	)`)
	if err != nil {
		return err
	}

	instance = &Tracker{
		db:        db,
		startedAt: time.Now(),
	}
	return nil
}

// RegisterCheck usa Mutex (no atomic) a propósito: además de incrementar
// un contador en memoria, escribe una fila en SQLite. Son dos pasos
// relacionados que deben pasar juntos — el caso donde SÍ conviene Mutex
// en vez de atomic (varios campos/pasos, no un solo valor suelto).
func (t *Tracker) RegisterCheck() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.checks++
	_, err := t.db.Exec("INSERT INTO checks (checked_at) VALUES (?)", time.Now())
	return err
}

func (t *Tracker) Snapshot() (uptime time.Duration, checks int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Round a segundos enteros — sin esto, uptime.String() imprime hasta
	// nanosegundos (ej. "7m31.715621695s"), que se ve sucio en el frontend.
	return time.Since(t.startedAt).Round(time.Second), t.checks
}
