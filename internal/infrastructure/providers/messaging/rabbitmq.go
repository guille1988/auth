package messaging

import (
	"auth/internal/infrastructure/rabbitmq"
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
	return &RabbitMQRegister{
		publisher: publisher,
		routes:    make(map[reflect.Type]Route),
	}
}

func (provider *RabbitMQRegister) Register(dto any, route Route) error {
	dtoType := reflect.TypeOf(dto)

	if dtoType.Kind() == reflect.Ptr {
		dtoType = dtoType.Elem()
	}

	err := provider.publisher.DeclareExchange(route.Exchange, route.ExchangeType)

	if err != nil {
		return fmt.Errorf("failed to declare exchange for %T: %w", dto, err)
	}

	provider.routes[dtoType] = route

	return nil
}

func (provider *RabbitMQRegister) Publish(ctx context.Context, dto any) error {
	dtoType := reflect.TypeOf(dto)

	if dtoType.Kind() == reflect.Ptr {
		dtoType = dtoType.Elem()
	}

	route, ok := provider.routes[dtoType]

	if !ok {
		return fmt.Errorf("no route registered for %T", dto)
	}

	return provider.publisher.Publish(ctx, route.Exchange, route.RoutingKey, dto)
}

func (provider *RabbitMQRegister) Close() error {
	if provider.publisher != nil {
		return provider.publisher.Close()
	}

	return nil
}
