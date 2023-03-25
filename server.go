package main

import (
	"dss-main/config"
	"dss-main/server"
	"fmt"
	"github.com/caarlos0/env"
	log "github.com/sirupsen/logrus"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	conf := new(config.Config)
	if err := env.Parse(conf); err != nil {
		log.Fatal(err)
	}
	log.SetLevel(conf.LogLevel)

	srv, err := server.NewServer(conf)
	if err != nil {
		log.Fatal(err)
	}

	defer srv.Close()

	app := fiber.New()

	app.Use(recover.New())

	app.Post("/", srv.Upload)

	if err := app.Listen(fmt.Sprintf(":%s", conf.Port)); err != nil {
		log.Fatal(err)
	}
}
