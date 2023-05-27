package server

import (
	"context"
	"dss-main/config"
	"fmt"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const PublishTimeout = 5 * time.Second

func Connect(conf *config.Config) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(conf.RabbitURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open a channel: %w", err)
	}
	return conn, channel, nil
}

func (s *Server) sendMessage(id string, i int, content []byte) error {
	publishContext, cancel := context.WithTimeout(context.Background(), PublishTimeout)
	defer cancel()

	err := s.Channel.PublishWithContext(publishContext,
		"",           // exchange
		s.Queue.Name, // routing key
		true,         // mandatory
		false,        // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Headers:      amqp.Table{"id": id, "fragment_number": strconv.Itoa(i)},
			Body:         content,
		})
	return err
}
