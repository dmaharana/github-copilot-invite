package middleware

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "github-copilot-invite/internal/rabbitmq"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/rs/zerolog/log"
)

// RequestMessage represents an API request message
type RequestMessage struct {
    CorrelationID string            `json:"correlation_id"`
    ResponseQueue string            `json:"response_queue"`
    Method        string            `json:"method"`
    Path          string            `json:"path"`
    Headers       map[string]string `json:"headers"`
    Body          json.RawMessage   `json:"body"`
}

// ResponseMessage represents an API response message
type ResponseMessage struct {
    CorrelationID string            `json:"correlation_id"`
    StatusCode    int               `json:"status_code"`
    Headers       map[string]string `json:"headers"`
    Body          json.RawMessage   `json:"body"`
}

// RabbitMQMiddleware creates a middleware that publishes requests to RabbitMQ and waits for responses
func RabbitMQMiddleware(broker *rabbitmq.Broker, queueName string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Generate correlation ID and response queue name
        correlationID := uuid.New().String()
        responseQueue := fmt.Sprintf("response-%s", correlationID)

        // Create temporary response queue
        _, err := broker.CreateTemporaryQueue(responseQueue)
        if err != nil {
            log.Error().Err(err).Msg("Failed to create response queue")
            c.AbortWithStatus(http.StatusInternalServerError)
            return
        }
        defer broker.DeleteQueue(responseQueue)

        // Read request body
        var bodyBytes []byte
        if c.Request.Body != nil {
            bodyBytes, _ = io.ReadAll(c.Request.Body)
            c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
        }

        // Create headers map
        headers := make(map[string]string)
        for k, v := range c.Request.Header {
            if len(v) > 0 {
                headers[k] = v[0]
            }
        }

        // Create request message
        msg := RequestMessage{
            CorrelationID: correlationID,
            ResponseQueue: responseQueue,
            Method:       c.Request.Method,
            Path:         c.Request.URL.Path,
            Headers:      headers,
            Body:         bodyBytes,
        }

        // Publish message
        msgBytes, _ := json.Marshal(msg)
        ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
        defer cancel()

        if err := broker.PublishMessage(ctx, queueName, msgBytes); err != nil {
            log.Error().Err(err).Msg("Failed to publish message")
            c.AbortWithStatus(http.StatusInternalServerError)
            return
        }

        // Wait for response
        response, err := broker.ConsumeResponse(ctx, responseQueue, correlationID)
        if err != nil {
            log.Error().Err(err).Msg("Failed to get response")
            c.AbortWithStatus(http.StatusGatewayTimeout)
            return
        }

        // Parse response
        var respMsg ResponseMessage
        if err := json.Unmarshal(response, &respMsg); err != nil {
            log.Error().Err(err).Msg("Failed to parse response")
            c.AbortWithStatus(http.StatusInternalServerError)
            return
        }

        // Set response headers
        for k, v := range respMsg.Headers {
            c.Header(k, v)
        }

        // Send response
        c.Data(respMsg.StatusCode, "application/json", respMsg.Body)
    }
}
