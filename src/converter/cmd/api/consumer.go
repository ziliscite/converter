package main

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/ziliscite/video-to-mp3/converter/internal/service"
	"time"
)

type consumer struct {
	cfg Config
	ac  *amqp.Connection
	cvs service.ConverterService
	vq  amqp.Queue
}

func newConsumer(cfg Config, ac *amqp.Connection, cvs service.ConverterService) (*consumer, error) {
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

	// notification queue will be declared on the publisher
	//nq, err := ch.QueueDeclare(cfg.rabbit.queue.notification, true, false, false, false, nil)
	//if err != nil {
	//	return nil, err
	//}

	return &consumer{
		cfg: cfg,
		ac:  ac,
		cvs: cvs,
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
				v.Nack(false, true)
				continue
			}

			v.Ack(false)
		}
	}()

	<-forever
	return nil
}

func (c *consumer) consumeVideo(ctx context.Context, body []byte) error {
	type request struct {
		UserId    int64  `json:"user_id"`
		UserEmail string `json:"user_email"`
		FileName  string `json:"file_name"`
		FileSize  int64  `json:"file_size"`
		FileKey   string `json:"file_key"`
	}

	var video request
	if err := json.Unmarshal(body, &video); err != nil {
		// reject
		return fmt.Errorf("error unmarshalling video: %v", err)
	}

	mp3Key, err := c.cvs.ConvertMP4(ctx, video.UserId, video.FileSize, video.FileName, video.FileKey)
	if err != nil {
		return fmt.Errorf("error converting video: %v", err)
	}

	// publish to notification queue

	return nil
}
