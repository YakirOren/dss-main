package main

import (
	"context"
	"fmt"

	"dss-main/config"
	"dss-main/fs"
	"dss-main/server"

	"github.com/docker/go-units"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"github.com/caarlos0/env/v7"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	log "github.com/sirupsen/logrus"
	"github.com/yakiroren/dss-common/db"
)

const maxAge = 3600

func main() {
	conf := &config.Config{}
	opts := env.Options{UseFieldNameByDefault: true}

	if err := env.Parse(conf, opts); err != nil {
		log.Fatal("config parsing failed:", err)
	}
	log.SetLevel(conf.LogLevel)

	store, err := db.NewMongoDataStore(&conf.Mongo)
	if err != nil {
		log.Fatal("could not connect to mongodb:", err)
	}

	srv, err := server.NewServer(conf, store)
	if err != nil {
		log.Fatal("server couldn't be created", err)
	}

	const limit = units.GiB * 5

	app := fiber.New(fiber.Config{
		BodyLimit:             limit,
		ReduceMemoryUsage:     true,
		DisableStartupMessage: true,
		UnescapePath:          true,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	api := app.Group("/api")

	v1 := api.Group("/v1")

	v1.Post("/upload", srv.Upload)
	v1.Post("/mkdir", srv.Mkdir)
	v1.Post("/rename/:id", srv.Rename)
	v1.Post("/move/:id", srv.Move)
	v1.Delete("/delete/:id", srv.Delete)
	v1.Get("/status/:id", srv.Status)
	v1.Get("/dir/*", srv.Dir)

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
		createRootDir(srv)
	}

	serverAddr := fmt.Sprintf(":%s", conf.Port)
	log.Error(app.Listen(serverAddr))
}

func createRootDir(srv *server.Server) {
	if err := srv.CreateDir(context.Background(), "/", "/"); err != nil {
		log.Fatal(err)
	}
	log.Info("created root dir")
}
