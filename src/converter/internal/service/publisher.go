package service

import amqp "github.com/rabbitmq/amqp091-go"

type publisher struct {
	ac *amqp.Connection
	vq amqp.Queue
}

func NewPublisher(ac *amqp.Connection, queueName string) (*publisher, error) {
	// create a new channel for the publisher
	ch, err := ac.Channel()
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	// declare the video queue once during publisher initialization
	vq, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	return &publisher{
		ac: ac,
		vq: vq,
	}, nil
}
