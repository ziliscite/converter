package main

import (
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/ziliscite/video-to-mp3/mailer/internal"
	"github.com/ziliscite/video-to-mp3/mailer/internal/domain"
)

type listener struct {
	amc *amqp.Connection
	mq  amqp.Queue
	mr  *internal.Mailer
}

func newService(cfg Config, amc *amqp.Connection, mr *internal.Mailer) (*listener, error) {
	ch, err := amc.Channel()
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		cfg.mq.queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &listener{
		amc: amc,
		mq:  q,
		mr:  mr,
	}, nil
}

func (s *listener) sendNotification(body []byte, email string) error {
	var mail domain.Metadata
	if err := json.Unmarshal(body, &mail); err != nil {
		return err
	}

	return s.mr.Send(email, "mp4_audio_notification.tmpl", map[string]interface{}{
		"userID":   mail.UserId,
		"filename": mail.FileName,
		"videoKey": mail.VideoKey,
		"audioKey": mail.AudioKey,
	})
}
