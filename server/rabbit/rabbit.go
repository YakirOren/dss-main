package rabbit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/wagslane/go-rabbitmq"
)

type Config struct {
	RabbitURL      string   `env:"RABBIT_URL,required,notEmpty"`
	ConsumerURL    []string `env:"CONSUMER_URL,required,notEmpty"`
	RoutingKey     string   `env:",required,notEmpty"`
	PublishTimeout int      `env:",required,notEmpty"`
}

type Publisher struct {
	publisher   *rabbitmq.Publisher
	routingKey  string
	timeout     time.Duration
	consumerURL []string
	conn        *rabbitmq.Conn
}

func New(conf Config, logger *log.Logger) (*Publisher, error) {
	conn, err := rabbitmq.NewConn(
		conf.RabbitURL,
		rabbitmq.WithConnectionOptionsLogger(logger),
		rabbitmq.WithConnectionOptionsLogging,
	)
	if err != nil {
		return nil, err
	}

	publisher, err := rabbitmq.NewPublisher(
		conn,
		rabbitmq.WithPublisherOptionsLogger(logger),
		rabbitmq.WithPublisherOptionsLogging,
	)
	if err != nil {
		return nil, err
	}

	timeout := time.Duration(conf.PublishTimeout) * time.Second
	logger.Debug("publish timeout ", timeout)

	return &Publisher{
		conn:        conn,
		consumerURL: conf.ConsumerURL,
		publisher:   publisher,
		routingKey:  conf.RoutingKey,
		timeout:     timeout,
	}, nil
}

func (pub *Publisher) Close() {
	pub.publisher.Close()
	pub.conn.Close()
}

func (pub *Publisher) NotifyConsumers() {
	log.Debug("triggering consumers")

	for _, url := range pub.consumerURL {
		err := call(url)
		if err != nil {
			log.Error(err)
		}
	}
}

func call(url string) error {
	start := time.Now()

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	res := struct {
		Status    string    `json:"status"`
		Timestamp time.Time `json:"timestamp"`
	}{}

	if decodeErr := json.NewDecoder(response.Body).Decode(&res); decodeErr != nil {
		return fmt.Errorf("failed to parse jsdon from consumer: %w", decodeErr)
	}
	elapsed := time.Since(start)

	log.Debugf("response from consumer %s, took %s", res.Status, elapsed)
	return nil
}

func (pub *Publisher) PushMessage(id string, fragmentNumber int, content []byte) error {
	headers := rabbitmq.Table{"id": id, "fragment_number": strconv.Itoa(fragmentNumber)}

	publishContext, cancel := context.WithTimeout(context.Background(), pub.timeout)
	defer cancel()

	return pub.publisher.PublishWithContext(publishContext, content, []string{pub.routingKey},
		rabbitmq.WithPublishOptionsHeaders(headers),
		rabbitmq.WithPublishOptionsMandatory,
		rabbitmq.WithPublishOptionsTimestamp(time.Now()),
		rabbitmq.WithPublishOptionsPersistentDelivery)
}
