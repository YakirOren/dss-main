package config

import log "github.com/sirupsen/logrus"

type Config struct {
	Port      string    `env:"PORT,required,notEmpty"`
	RabbitUrl string    `env:"RABBIT_URL,required,notEmpty"`
	QueueName string    `env:"QUEUE_NAME,required,notEmpty"`
	LogLevel  log.Level `env:"LOG_LEVEL,required,notEmpty"`
}
