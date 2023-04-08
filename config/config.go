package config

import log "github.com/sirupsen/logrus"

import "github.com/yakiroren/dss-common/db"

type Config struct {
	Port      string    `env:",required,notEmpty"`
	RabbitUrl string    `env:",required,notEmpty"`
	QueueName string    `env:",required,notEmpty"`
	LogLevel  log.Level `env:",required,notEmpty"`
	Mongo     db.MongoConfig
}
