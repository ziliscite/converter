package main

import (
	"errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/ziliscite/video-to-mp3/mailer/internal"
	"log/slog"
)

func (s *listener) listen() error {
	ch, err := s.amc.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	mails, err := ch.Consume(s.mq.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	// consume til application exits
	forever := make(chan bool)
	go func() {
		for m := range mails {
			email, ok := m.Headers["email"].(string)
			if !ok {
				m.Nack(false, false)
				continue
			}

			if err = s.sendNotification(m.Body, email); err != nil {
				switch {
				case errors.Is(err, internal.ErrConnection) || errors.Is(err, amqp.ErrClosed):
					m.Nack(false, true)
				default:
					m.Nack(false, false)
				}
				continue
			}

			m.Ack(false)
		}
	}()

	slog.Info("Listening for notifications...")
	<-forever
	return nil
}
