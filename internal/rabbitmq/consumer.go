package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
)

// MessageProcessor is a function type for processing messages and returning responses
type MessageProcessor func(context.Context, []byte) (int, map[string]string, []byte, error)

// Consumer handles message consumption from RabbitMQ
type Consumer struct {
	broker      *Broker
	processors  map[string]MessageProcessor
}

// NewConsumer creates a new consumer
func NewConsumer(broker *Broker) *Consumer {
	return &Consumer{
		broker:     broker,
		processors: make(map[string]MessageProcessor),
	}
}

// RegisterProcessor registers a message processor for a specific queue
func (c *Consumer) RegisterProcessor(queueName string, processor MessageProcessor) {
	c.processors[queueName] = processor
}

// StartConsuming starts consuming messages from registered queues
func (c *Consumer) StartConsuming(ctx context.Context) error {
	for queueName, processor := range c.processors {
		// Declare queue
		_, err := c.broker.ch.QueueDeclare(
			queueName,
			true,  // durable
			false, // auto-delete
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %v", queueName, err)
		}

		msgs, err := c.broker.ch.Consume(
			queueName,
			"",    // consumer
			false, // auto-ack
			false, // exclusive
			false, // no-local
			false, // no-wait
			nil,   // args
		)
		if err != nil {
			return fmt.Errorf("failed to consume from queue %s: %v", queueName, err)
		}

		go func(qName string, proc MessageProcessor) {
			for {
				select {
				case msg := <-msgs:
					// Parse request message
					var reqMsg struct {
						CorrelationID string          `json:"correlation_id"`
						ResponseQueue string          `json:"response_queue"`
						Body          json.RawMessage `json:"body"`
					}
					if err := json.Unmarshal(msg.Body, &reqMsg); err != nil {
						log.Error().Err(err).Msg("Failed to unmarshal request")
						msg.Nack(false, false)
						continue
					}

					// Process message
					statusCode, headers, responseBody, err := proc(ctx, reqMsg.Body)
					if err != nil {
						log.Error().Err(err).Msg("Failed to process message")
						statusCode = 500
						responseBody = []byte(fmt.Sprintf(`{"error": "%v"}`, err))
					}

					// Create response message
					respMsg := struct {
						CorrelationID string            `json:"correlation_id"`
						StatusCode    int               `json:"status_code"`
						Headers       map[string]string `json:"headers"`
						Body          json.RawMessage   `json:"body"`
					}{
						CorrelationID: reqMsg.CorrelationID,
						StatusCode:    statusCode,
						Headers:       headers,
						Body:         responseBody,
					}

					// Marshal response
					respBytes, err := json.Marshal(respMsg)
					if err != nil {
						log.Error().Err(err).Msg("Failed to marshal response")
						msg.Nack(false, false)
						continue
					}

					// Publish response
					err = c.broker.PublishMessage(ctx, reqMsg.ResponseQueue, respBytes)
					if err != nil {
						log.Error().Err(err).Msg("Failed to publish response")
						msg.Nack(false, false)
						continue
					}

					msg.Ack(false)
				case <-ctx.Done():
					return
				}
			}
		}(queueName, processor)
	}

	return nil
}
