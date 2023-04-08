package server

import (
	"bytes"
	"context"
	"dss-main/config"
	"fmt"
	"github.com/docker/go-units"
	"github.com/dustin/go-humanize"
	"github.com/gofiber/fiber/v2"
	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"time"
)

type Server struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
	Queue   amqp.Queue
}

const DefaultFileLimit = units.MiB * 8

func NewServer(conf *config.Config) (*Server, error) {
	conn, channel, err := Connect(conf)
	if err != nil {
		return nil, err
	}

	queue, err := channel.QueueDeclare(
		conf.QueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare a queue: %w", err)
	}

	return &Server{
		Conn:    conn,
		Channel: channel,
		Queue:   queue,
	}, nil
}

func Connect(conf *config.Config) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(conf.RabbitUrl)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open a channel: %w", err)
	}
	return conn, channel, nil
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

	log.Info("got file with size ", humanize.IBytes(uint64(file.Size)))
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

	content := &bytes.Buffer{}

	log.Info("Total fragments ", totalFragments)

	for i := 1; i <= totalFragments; i++ {
		io.CopyN(content, src, DefaultFileLimit)

		publishContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = s.Channel.PublishWithContext(publishContext,
			"",           // exchange
			s.Queue.Name, // routing key
			true,         // mandatory
			false,        // immediate
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				Timestamp:    time.Now(),
				Headers:      map[string]interface{}{"filename": file.Filename, "fragment_number": i, "total_fragments": totalFragments},
				Body:         content.Bytes(),
			})

		if err != nil {
			log.Fatal(err)
			ctx.Status(http.StatusInternalServerError)
			return fmt.Errorf("upload failed")
		}

		log.Debug("pushed fragment number ", i)

		content.Reset()
	}

	return nil
}
