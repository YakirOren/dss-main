package main

import (
	"dss-main/config"
	"dss-main/fs"
	"dss-main/server"
	"fmt"
	"github.com/caarlos0/env/v7"
	"github.com/docker/go-units"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	log "github.com/sirupsen/logrus"
	"github.com/yakiroren/dss-common/db"
)

func main() {
	conf := &config.Config{}
	opts := env.Options{UseFieldNameByDefault: true}

	if err := env.Parse(conf, opts); err != nil {
		log.Fatal(err)
	}
	log.SetLevel(conf.LogLevel)

	srv, err := server.NewServer(conf)
	if err != nil {
		log.Fatal(err)
	}

	defer srv.Close()

	app := fiber.New(fiber.Config{
		BodyLimit: units.GiB,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Get("/metrics", monitor.New(monitor.Config{Title: "MyService Metrics Page"}))

	app.Post("/upload", srv.Upload)

	store, err := db.NewMongoDataStore(&conf.Mongo)
	if err != nil {
		log.Fatal(err)
	}

	myfs, err := fs.New(store)
	if err != nil {
		log.Fatal(err)
	}

	app.Use(
		filesystem.New(filesystem.Config{
			Root:   myfs,
			Browse: true,
			Index:  "/",
			MaxAge: 3600,
		}))

	if err := app.Listen(fmt.Sprintf(":%s", conf.Port)); err != nil {
		log.Fatal(err)
	}
}
