package config

import (
	"dss-main/server/rabbit"
	log "github.com/sirupsen/logrus"
	"github.com/yakiroren/dss-common/db"
)

type Config struct {
	Port         string    `env:",required,notEmpty"`
	LogLevel     log.Level `env:",required,notEmpty"`
	FragmentSize int64     `env:",required,notEmpty"`
	Publisher    rabbit.Config
	Mongo        db.MongoConfig
}
