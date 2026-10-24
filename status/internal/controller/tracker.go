package controller

import (
	"database/sql"
	"sync"
	"time"
)

// el Singleton -> existe UNA sola instancia en todo el programa
// la comparten todos los handlers que la usan
// guarda cuantos health checks van y desde cuando esta corriendo el servicio
type Tracker struct {
	mu        sync.Mutex
	db        *sql.DB
	startedAt time.Time
	checks    int

	// lo que va contando el consumidor de kafka del topico events.created
	// vive solo en memoria y si el servicio reinicia vuelve a cero
	// kafka ya recuerda hasta donde leyo este group y no repite los mensajes viejos
	eventsSeen int
	bySeverity map[string]int
}

// la unica instancia del Tracker
// es una variable del paquete y nunca se crea una segunda
var instance *Tracker

// crea la unica instancia y prepara la tabla en sqlite
// se llama UNA vez en main
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

		bySeverity: make(map[string]int),
	}
	return nil
}

// lo llama el consumidor de kafka por cada evento nuevo
// Mutex y no atomic pq tocan el contador total y el mapa juntos
func (t *Tracker) RegisterEvent(severity string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.eventsSeen++
	t.bySeverity[severity]++
}

// regresa una COPIA del mapa
// si regresara el mismo, quien lo lea lo estaria leyendo sin el candado mientras el consumidor lo cambia
func (t *Tracker) EventCounts() (total int, bySeverity map[string]int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	copied := make(map[string]int, len(t.bySeverity))
	for k, v := range t.bySeverity {
		copied[k] = v
	}
	return t.eventsSeen, copied
}

// usa Mutex y no atomic a proposito
// ademas de subir el contador en memoria escribe una fila en sqlite
// son dos pasos que deben pasar juntos y por eso aqui si conviene Mutex
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

	// redondea a segundos enteros
	// sin esto uptime.String saca hasta nanosegundos y se ve sucio en el frontend
	return time.Since(t.startedAt).Round(time.Second), t.checks
}
