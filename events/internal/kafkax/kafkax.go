// kafkax -> helpers chicos encima de segmentio/kafka-go
// mismo archivo en backend, events y status igual que consul.go
// basado en el kafkax del kafka_example del profe
package kafkax

import (
	"context"
	"errors"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

// crea el topico si no existe y si ya existe no pasa nada
// las particiones son el maximo de paralelismo que puede tener un consumer group
func EnsureTopic(broker, topic string, partitions int) error {
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		return err
	}
	defer conn.Close()

	// los topicos solo se pueden crear hablando con el controller del cluster
	// con un solo broker es el mismo pero el codigo funciona igual con varios
	controller, err := conn.Controller()
	if err != nil {
		return err
	}
	ctrl, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return err
	}
	defer ctrl.Close()

	return ctrl.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     partitions,
		ReplicationFactor: 1, // un solo broker, en produccion serian 3
	})
}

// el productor de un topico
// mensajes con la misma key caen siempre en la misma particion y asi se mantiene el orden por key
func NewWriter(brokers []string, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.Hash{},    // particion = hash de la key
		RequiredAcks: kafka.RequireAll, // espera a que todas las replicas confirmen
		BatchTimeout: 10 * time.Millisecond,
	}
}

// procesa UN mensaje y si regresa error es que fallo
type Handler func(ctx context.Context, msg kafka.Message) error

// entra a un consumer group y procesa mensajes hasta que se cancele el ctx
// el offset se confirma hasta DESPUES de que el handler termina bien
// osea at-least-once -> si el servicio se cae a la mitad el mensaje llega otra vez
// por eso el handler debe aguantar duplicados
//
// dlq es opcional y puede ser nil
// si un mensaje falla todos los intentos se manda ahi y el consumidor sigue con el siguiente
func Consume(ctx context.Context, brokers []string, topic, groupID string, handle Handler, dlq *kafka.Writer) error {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		GroupID:     groupID,
		StartOffset: kafka.FirstOffset, // un group nuevo empieza desde el primer mensaje
		MinBytes:    1,
		MaxBytes:    10e6,
	})
	defer r.Close()

	for {
		// FetchMessage no confirma solo y por eso el commit se hace a mano abajo
		msg, err := r.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}

		if err := processWithRetry(ctx, handle, msg, 3); err != nil {
			log.Printf("kafka: me rindo con %s partition=%d offset=%d: %v", topic, msg.Partition, msg.Offset, err)
			if dlq != nil {
				// se manda tal cual el mensaje original
				if err := dlq.WriteMessages(ctx, kafka.Message{Key: msg.Key, Value: msg.Value}); err != nil {
					log.Printf("kafka: tampoco se pudo mandar a la dlq: %v", err)
				}
			}
		}

		if err := r.CommitMessages(ctx, msg); err != nil {
			return err
		}
	}
}

// reintenta el handler y espera un poco mas en cada intento, 0.5s 1s 1.5s
func processWithRetry(ctx context.Context, h Handler, msg kafka.Message, attempts int) error {
	var err error
	for i := 1; i <= attempts; i++ {
		if err = h(ctx, msg); err == nil {
			return nil
		}
		log.Printf("kafka: intento %d/%d fallo: %v", i, attempts, err)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(i) * 500 * time.Millisecond):
		}
	}
	return err
}
