package main

import (
	"dss-main/config"
	"dss-main/server"
	"dss-main/sizes"
	"fmt"
	"github.com/caarlos0/env/v7"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	log "github.com/sirupsen/logrus"
)

func main() {
	conf := &config.Config{}

	if err := env.Parse(conf); err != nil {
		log.Fatal(err)
	}
	log.SetLevel(conf.LogLevel)

	srv, err := server.NewServer(conf)
	if err != nil {
		log.Fatal(err)
	}

	defer srv.Close()

	app := fiber.New(fiber.Config{
		BodyLimit: sizes.Gib,
	})

	app.Use(recover.New())

	app.Post("/upload", srv.Upload)

	if err := app.Listen(fmt.Sprintf(":%s", conf.Port)); err != nil {
		log.Fatal(err)
	}
}
