package config

import log "github.com/sirupsen/logrus"

type Config struct {
	Port      string `env:"PORT,required"`
	RabbitUrl string `env:"RABBIT_URL,required"`
	QueueName string `env:"QUEUE_NAME,required"`

	LogLevel log.Level `env:"LOG_LEVEL,required"`
}
