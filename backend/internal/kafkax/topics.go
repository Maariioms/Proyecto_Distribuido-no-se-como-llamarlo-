package kafkax

import "time"

// los topicos y los consumer groups de todo el proyecto
// mismo archivo en los 3 modulos pa que todos hablen igual
// igual que events.go del kafka_example del profe
const (
	// accounts publica aqui cuando hay registro o login
	TopicAccountsActivity = "accounts.activity"
	// aqui van los que fallaron todos los intentos
	TopicAccountsActivityDLQ = "accounts.activity.dlq"
	// events publica aqui cada evento que guarda en la base
	TopicEventsCreated = "events.created"

	// events lee accounts.activity y lo guarda en postgres
	GroupEventsWriter = "events-writer"
	// status lee events.created y cuenta
	GroupStatusCounter = "status-counter"

	// 3 particiones pa todos
	Partitions = 3
)

// el mensaje que viaja por los topicos y es json
type EventMessage struct {
	Type     string    `json:"type"`
	Severity string    `json:"severity"`
	At       time.Time `json:"at"`
}
