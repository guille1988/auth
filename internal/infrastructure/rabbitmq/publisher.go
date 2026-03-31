package rabbitmq

import (
	"auth/internal/infrastructure/config"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

func NewPublisher(cfg config.RabbitMQConfig) (*Publisher, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.User, cfg.Password, cfg.Host, cfg.Port)

	const maxRetries = 5
	backoff := 2 * time.Second

	var connection *amqp.Connection
	var err error

	for i := 1; i <= maxRetries; i++ {
		connection, err = amqp.Dial(url)

		if err == nil {
			break
		}

		slog.Warn("failed to connect to RabbitMQ, retrying...", "attempt", i, "error", err)
		time.Sleep(backoff)
		backoff *= 2
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ after %d attempts: %w", maxRetries, err)
	}

	var channel *amqp.Channel
	var ok bool

	defer func() {
		if !ok {
			if channel != nil {
				_ = channel.Close()
			}
			_ = connection.Close()
		}
	}()

	channel, err = connection.Channel()

	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	ok = true

	return &Publisher{
		connection: connection,
		channel:    channel,
	}, nil
}

func (publisher *Publisher) DeclareExchange(exchange, exchangeType string) error {
	return publisher.channel.ExchangeDeclare(exchange, exchangeType, true, false, false, false, nil)
}

func (publisher *Publisher) Publish(ctx context.Context, exchange, routingKey string, dto any) error {
	body, err := json.Marshal(dto)

	if err != nil {
		return fmt.Errorf("failed to marshal dto: %w", err)
	}

	err = publisher.channel.PublishWithContext(ctx,
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})

	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}

	slog.Info("message published", "exchange", exchange, "routingKey", routingKey)

	return nil
}

func (publisher *Publisher) Close() error {
	err := publisher.channel.Close()

	if err != nil {
		return err
	}

	return publisher.connection.Close()
}
