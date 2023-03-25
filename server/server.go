package server

import (
	"context"
	"dss-main/config"
	"fmt"
	"github.com/gofiber/fiber/v2"
	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
	"math"
	"mime/multipart"
	"net/http"
	"time"
)

const (
	kilobyte              = 1024
	MB                    = 1024 * kilobyte
	DefaultFileLimit      = 8 * MB
	ClassicNitroFileLimit = 50 * MB
	NitroFileLimit        = 100 * MB
)

type Server struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
	Queue   amqp.Queue
}

func NewServer(conf *config.Config) (*Server, error) {
	conn, err := amqp.Dial(conf.RabbitUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	queue, err := channel.QueueDeclare(
		conf.QueueName, // name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to declare a queue: %w", err)
	}

	return &Server{
		Conn:    conn,
		Channel: channel,
		Queue:   queue,
	}, nil
}

func (s *Server) Close() {
	s.Channel.Close()
	s.Conn.Close()
}

func (s *Server) Upload(ctx *fiber.Ctx) error {
	file, err := ctx.FormFile("file")
	if err != nil {
		return err
	}

	log.Info("got file with size ", file.Size)

	src, err := file.Open()
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return fmt.Errorf("cant open file")
	}

	defer func(src multipart.File) {
		err := src.Close()
		if err != nil {
			log.Error(err)
		}
	}(src)

	totalFragments := int(math.Ceil(float64(file.Size) / DefaultFileLimit))

	for i := 0; i < totalFragments; i++ {
		bytes := make([]byte, DefaultFileLimit)

		_, err := src.Read(bytes)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = s.Channel.PublishWithContext(ctx,
			"",           // exchange
			s.Queue.Name, // routing key
			false,        // mandatory
			false,        // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        bytes,
			})

	}

	return nil
}
