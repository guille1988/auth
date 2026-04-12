package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Route struct {
	Exchange     string
	RoutingKey   string
	ExchangeType string
}

type Publisher interface {
	Publish(dto any) error
	Close() error
}

type KafkaPublisher struct {
	client *kgo.Client
	topics map[reflect.Type]string
}

func NewKafkaPublisher(brokers string) *KafkaPublisher {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers),
		kgo.AllowAutoTopicCreation(),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create kafka client: %v", err))
	}

	return &KafkaPublisher{
		client: client,
		topics: make(map[reflect.Type]string),
	}
}

func (publisher *KafkaPublisher) Register(dto any, route Route) error {
	t := reflect.TypeOf(dto)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	publisher.topics[t] = route.RoutingKey

	return nil
}

// Publish is fire-and-forget (async produce).
func (publisher *KafkaPublisher) Publish(dto any) error {
	t := reflect.TypeOf(dto)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	topic, ok := publisher.topics[t]

	if !ok {
		return fmt.Errorf("no topic registered for %T", dto)
	}

	body, err := json.Marshal(dto)

	if err != nil {
		return err
	}

	publisher.client.Produce(
		context.Background(),
		&kgo.Record{Topic: topic, Value: body},
		func(record *kgo.Record, err error) {
			if err != nil {
				slog.Error("kafka produce failed", "topic", record.Topic, "error", err)
			}
		},
	)

	return nil
}

func (publisher *KafkaPublisher) Close() error {
	publisher.client.Close()
	return nil
}
