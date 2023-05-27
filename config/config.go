package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/yakiroren/dss-common/db"
)

type Config struct {
	Port      string    `env:",required,notEmpty"`
	RabbitURL string    `env:"RABBIT_URL,required,notEmpty"`
	QueueName string    `env:",required,notEmpty"`
	LogLevel  log.Level `env:",required,notEmpty"`
	Mongo     db.MongoConfig
}
