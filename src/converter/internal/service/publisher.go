package service

import (
	"context"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/ziliscite/video-to-mp3/converter/internal/domain"
)

type EmailNotification interface {
	PublishEmailNotification(ctx context.Context, data *domain.Metadata, email string) error
}

type NotificationService interface {
	EmailNotification
}

type Publisher struct {
	ac *amqp.Connection
	nq amqp.Queue
}

func NewPublisher(ac *amqp.Connection, queueName string) (NotificationService, error) {
	// create a new channel for the publisher
	ch, err := ac.Channel()
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	// declare the video queue once during publisher initialization
	nq, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		ac: ac,
		nq: nq,
	}, nil
}

func (p *Publisher) PublishEmailNotification(ctx context.Context, data *domain.Metadata, email string) error {
	ch, err := p.ac.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	msg, err := json.Marshal(&data)
	if err != nil {
		return err
	}

	return ch.PublishWithContext(ctx,
		"",
		p.nq.Name,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         msg,
			Headers: amqp.Table{
				"email": email,
			},
		},
	)
}
