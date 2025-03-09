package service

import (
	"context"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/ziliscite/video-to-mp3/gateway/internal/domain"
)

type FilePublisher interface {
	PublishVideo(ctx context.Context, video *domain.Video) error
}

type publisher struct {
	ac *amqp.Connection
	vq amqp.Queue
}

func NewPublisher(ac *amqp.Connection, queueName string) (FilePublisher, error) {
	// create a new channel for the publisher
	ch, err := ac.Channel()
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	// declare the video queue once during publisher initialization
	vq, err := ch.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	return &publisher{
		ac: ac,
		vq: vq,
	}, nil
}

func (p *publisher) PublishVideo(ctx context.Context, video *domain.Video) error {
	ch, err := p.ac.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// encode the video struct to json
	msg, err := json.Marshal(video)
	if err != nil {
		return err
	}

	return ch.PublishWithContext(ctx,
		"",
		p.vq.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        msg,
		},
	)
}
