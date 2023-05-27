package main

import (
	"context"
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

const maxAge = 3600

func main() {
	conf := &config.Config{}
	opts := env.Options{UseFieldNameByDefault: true}

	if err := env.Parse(conf, opts); err != nil {
		log.Fatal(err)
	}
	log.SetLevel(conf.LogLevel)

	store, err := db.NewMongoDataStore(&conf.Mongo)
	if err != nil {
		log.Fatal(err)
	}

	srv, err := server.NewServer(conf, store)
	if err != nil {
		log.Fatal(err)
	}

	defer srv.Close()

	app := fiber.New(fiber.Config{
		BodyLimit: units.GiB,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Get("/metrics", monitor.New(monitor.Config{Title: "dss-main Metrics Page"}))

	app.Post("/upload", srv.Upload)
	app.Post("/mkdir", srv.Mkdir)

	app.Post("/rename", srv.Rename)

	dfs, err := fs.New(store)
	if err != nil {
		log.Error(err)
	}

	app.Use(
		filesystem.New(filesystem.Config{
			Root:   dfs,
			Browse: true,
			Index:  "/",
			MaxAge: maxAge,
		}))

	_, err = dfs.Open("/")
	if err != nil {
		if err := srv.CreateDir(context.Background(), "/", "/"); err != nil {
			log.Fatal(err)
		}
		log.Info("created root dir")
	}

	serverAddr := fmt.Sprintf(":%s", conf.Port)
	log.Error(app.Listen(serverAddr))
}
