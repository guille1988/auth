package messaging

import (
	"auth/internal/infrastructure/rabbitmq"
	"auth/internal/shared/messaging/rabbitmq/dtos"
	"context"
	"fmt"
	"reflect"
)

type Route struct {
	Exchange     string
	RoutingKey   string
	ExchangeType string
}

type RabbitMQRegister struct {
	publisher *rabbitmq.Publisher
	routes    map[reflect.Type]Route
}

func NewRabbitMQRegister(publisher *rabbitmq.Publisher) *RabbitMQRegister {
	provider := &RabbitMQRegister{
		publisher: publisher,
		routes:    make(map[reflect.Type]Route),
	}

	provider.routes[reflect.TypeFor[dtos.WelcomeEmail]()] = Route{
		Exchange:     "auth.events",
		RoutingKey:   "user.created",
		ExchangeType: "topic",
	}

	return provider
}

func (provider *RabbitMQRegister) Close() error {
	if provider.publisher != nil {
		return provider.publisher.Close()
	}
	return nil
}

func (provider *RabbitMQRegister) Publish(ctx context.Context, dto any) error {
	dtoType := reflect.TypeOf(dto)

	if dtoType.Kind() == reflect.Ptr {
		dtoType = dtoType.Elem()
	}

	route, ok := provider.routes[dtoType]

	if !ok {
		return fmt.Errorf("no route registered for %T", dtoType)
	}

	return provider.publisher.Publish(ctx, route.Exchange, route.ExchangeType, route.RoutingKey, dto)
}
