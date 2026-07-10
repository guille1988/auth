package messaging

import (
	"auth/internal/infrastructure/providers/messaging"
	"sync"
	"testing"
)

type fakeDTO struct{}

/*
This test does not reflect the real bootstrap order (Register always
completes before Publish is ever called), but it proves the topics map
has no unsynchronized concurrent access under -race, defensively guarding
against that ordering assumption ever breaking.
*/
func TestKafkaPublisherRegisterAndPublishConcurrentAccessIsRaceFree(test *testing.T) {
	publisher := messaging.NewKafkaPublisher("kafka:9092")
	defer func() { _ = publisher.Close() }()

	var waitGroup sync.WaitGroup

	for i := 0; i < 50; i++ {
		waitGroup.Add(2)

		go func() {
			defer waitGroup.Done()
			_ = publisher.Register(fakeDTO{}, messaging.Route{RoutingKey: "test.topic"})
		}()

		go func() {
			defer waitGroup.Done()
			_ = publisher.Publish(fakeDTO{})
		}()
	}

	waitGroup.Wait()
}
