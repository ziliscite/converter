package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/ziliscite/video-to-mp3/converter/internal/service"
	"net"
	"time"
)

type consumer struct {
	cfg Config
	ac  *amqp.Connection
	cvs service.ConverterService
	np  service.NotificationService
	vq  amqp.Queue
}

func newConsumer(cfg Config, ac *amqp.Connection, cvs service.ConverterService, np service.NotificationService) (*consumer, error) {
	ch, err := ac.Channel()
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	// make queue durable
	vq, err := ch.QueueDeclare(cfg.rabbit.queue.video, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	return &consumer{
		cfg: cfg,
		ac:  ac,
		cvs: cvs,
		np:  np,
		vq:  vq,
	}, nil
}

func (c *consumer) consume() error {
	ch, err := c.ac.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// consume video queue
	videos, err := ch.Consume(
		c.vq.Name, // queue
		"",        // consumer
		false,     // no auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	forever := make(chan bool)
	go func() {
		for v := range videos {
			if err = c.consumeVideo(ctx, v.Body); err != nil {
				var netErr net.Error
				var amqpErr *amqp.Error
				switch {
				case errors.Is(err, amqp.ErrClosed):
					v.Nack(false, true)
				case errors.As(err, &netErr) && netErr.Temporary():
					v.Nack(false, true)
				case errors.As(err, &amqpErr) && amqpErr.Code == 320:
					v.Nack(false, true)
				case errors.Is(err, context.DeadlineExceeded):
					v.Nack(false, true)
				case errors.Is(err, service.ErrInternal):
					v.Nack(false, true)
				default:
					v.Nack(false, false)
				}
				continue
			}

			v.Ack(false)
		}
	}()

	<-forever
	return nil
}

func (c *consumer) consumeVideo(ctx context.Context, body []byte) error {
	var request struct {
		UserId    int64  `json:"user_id"`
		UserEmail string `json:"user_email"`
		FileSize  int64  `json:"file_size"`
		FileKey   string `json:"file_key"`
	}

	if err := json.Unmarshal(body, &request); err != nil {
		// reject
		return fmt.Errorf("error unmarshalling video: %v", err)
	}

	result, err := c.cvs.ConvertMP4(ctx, request.UserId, request.FileSize, request.FileKey)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInternal):
			return fmt.Errorf("error converting video: %w", err)
		default:
			return fmt.Errorf("error converting video: %v", err)
		}
	}

	// publish to notification queue
	if err = c.np.PublishEmailNotification(ctx, result, request.UserEmail); err != nil {
		return fmt.Errorf("error publishing notification: %v", err)
	}

	return nil
}
