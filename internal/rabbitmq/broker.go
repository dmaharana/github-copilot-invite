package rabbitmq

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

// Broker handles RabbitMQ connections and operations
type Broker struct {
	config *Config
	conn   *amqp.Connection
	ch     *amqp.Channel
}

// NewBroker creates a new RabbitMQ broker
func NewBroker(config *Config) (*Broker, error) {
	broker := &Broker{
		config: config,
	}
	if err := broker.connect(); err != nil {
		return nil, err
	}
	return broker, nil
}

// connect establishes connection to RabbitMQ
func (b *Broker) connect() error {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		b.config.Username,
		b.config.Password,
		b.config.Host,
		b.config.Port,
		b.config.VHost,
	)

	conn, err := amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %v", err)
	}

	b.conn = conn
	b.ch = ch

	return nil
}

// Close closes the RabbitMQ connection
func (b *Broker) Close() error {
	if err := b.ch.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %v", err)
	}
	if err := b.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %v", err)
	}
	return nil
}

// PublishMessage publishes a message to a queue
func (b *Broker) PublishMessage(ctx context.Context, queueName string, message []byte) error {
	// Ensure queue exists
	_, err := b.ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %v", err)
	}

	// Publish message
	err = b.ch.PublishWithContext(
		ctx,
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message,
			Timestamp:   time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	log.Debug().
		Str("queue", queueName).
		Int("message_size", len(message)).
		Msg("Message published to queue")

	return nil
}

// CreateTemporaryQueue creates a temporary response queue
func (b *Broker) CreateTemporaryQueue(name string) (*amqp.Queue, error) {
	q, err := b.ch.QueueDeclare(
		name,
		false, // durable
		true,  // auto-delete
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	return &q, err
}

// DeleteQueue deletes a queue
func (b *Broker) DeleteQueue(name string) error {
	_, err := b.ch.QueueDelete(
		name,
		false, // ifUnused
		false, // ifEmpty
		false, // noWait
	)
	return err
}

// ConsumeResponse consumes a single response message from a queue
func (b *Broker) ConsumeResponse(ctx context.Context, queueName, correlationID string) ([]byte, error) {
	msgs, err := b.ch.Consume(
		queueName,
		"",    // consumer
		true,  // auto-ack
		true,  // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return nil, err
	}

	select {
	case msg := <-msgs:
		return msg.Body, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("timeout waiting for response")
	}
}
