package actions

import (
	"auth/internal/domain/stress/data"
	"context"

	"github.com/guille1988/go-app-shared/messaging/rabbitmq/dtos"
)

type MessagePublisher interface {
	Publish(ctx context.Context, dto any) error
}

type SendStress struct {
	publisher MessagePublisher
}

func NewSendStress(publisher MessagePublisher) *SendStress {
	return &SendStress{publisher: publisher}
}

func (action *SendStress) Execute(ctx context.Context, stressData data.Stress) error {
	return action.publisher.Publish(ctx, dtos.StressEmail{
		Email: stressData.Email,
		Name:  stressData.Name,
	})
}
