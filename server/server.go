package server

import (
	"bytes"
	"context"
	"dss-main/config"
	"encoding/json"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/docker/go-units"
	"github.com/dustin/go-humanize"
	"github.com/gofiber/fiber/v2"
	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
	"github.com/yakiroren/dss-common/db"
	"github.com/yakiroren/dss-common/models"
)

type Server struct {
	Conn         *amqp.Connection
	Channel      *amqp.Channel
	Queue        amqp.Queue
	datastore    db.DataStore
	fragmentSize int64
}

const DefaultFileLimit = units.MiB * 25

func NewServer(conf *config.Config, datastore db.DataStore) (*Server, error) {
	conn, channel, err := Connect(conf)
	if err != nil {
		return nil, err
	}

	queue, err := channel.QueueDeclare(
		conf.QueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare a queue: %w", err)
	}

	return &Server{
		Conn:         conn,
		Channel:      channel,
		Queue:        queue,
		datastore:    datastore,
		fragmentSize: DefaultFileLimit,
	}, nil
}

func (s *Server) Close() {
	s.Channel.Close()
	s.Conn.Close()
}

func (s *Server) Mkdir(ctx *fiber.Ctx) error {
	name := ctx.FormValue("name")
	if name == "" {
		return errors.New("name cant be empty")
	}

	targetPath := ctx.FormValue("path", "/")

	valid := validatePath(targetPath)
	if !valid {
		return fiber.NewError(http.StatusBadRequest, "the provided path is not valid")
	}

	err := s.CreateDir(ctx.Context(), targetPath, name)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	ctx.Status(http.StatusCreated)
	return nil
}

func (s *Server) CreateDir(ctx context.Context, targetPath string, name string) error {
	newpath := filepath.Join(targetPath, name)

	_, exists := s.datastore.GetMetadataByPath(ctx, newpath)
	if exists {
		return errors.New("file exists")
	}

	_, err := s.datastore.WriteFile(ctx, models.FileMetadata{
		Id:          primitive.NewObjectID(),
		FileName:    name,
		IsDirectory: true,
		Path:        targetPath,
		IsHidden:    false,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) Upload(ctx *fiber.Ctx) error {
	file, err := ctx.FormFile("file")
	if err != nil {
		return err
	}

	targetPath := ctx.FormValue("path", "/")

	valid := validatePath(targetPath)
	if !valid {
		return fiber.NewError(http.StatusBadRequest, "the provided path is not valid")
	}

	log.Info("got file with size ", humanize.IBytes(uint64(file.Size)))
	src, err := file.Open()
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return fmt.Errorf("cant open file")
	}

	defer func(src multipart.File) {
		if err := src.Close(); err != nil {
			log.Error(err)
		}
	}(src)

	totalFragments := int(math.Ceil(float64(file.Size) / float64(s.fragmentSize)))

	filename := s.fixFilename(ctx.Context(), file.Filename, targetPath)

	fileID, err := s.datastore.WriteFile(ctx.Context(), models.FileMetadata{
		Id:             primitive.NewObjectID(),
		FileName:       filename,
		FileSize:       file.Size,
		CurrentSize:    0,
		IsDirectory:    false,
		Path:           targetPath,
		Fragments:      []models.Fragment{},
		IsHidden:       true,
		TotalFragments: totalFragments,
	})
	if err != nil {
		return err
	}

	done, err := s.writeToQueue(totalFragments, src, fileID)
	if !done {
		return err
	}

	ctx.Status(http.StatusCreated).SendString(fileID)
	return nil
}

func (s *Server) writeToQueue(totalFragments int, src io.Reader, id string) (bool, error) {
	content := &bytes.Buffer{}
	log.Info("Total fragments ", totalFragments)

	for i := 1; i <= totalFragments; i++ {
		_, err := io.CopyN(content, src, s.fragmentSize)
		if errors.Is(err, io.EOF) {
		} else if err != nil {
			log.Fatal(err)
			return false, fiber.ErrInternalServerError
		}

		if err = s.sendMessage(id, i, content.Bytes()); err != nil {
			log.Fatal(err)
			return false, fiber.ErrInternalServerError
		}

		log.Debug("pushed fragment number ", i)

		content.Reset()
	}

	return true, nil
}

func (s *Server) Rename(ctx *fiber.Ctx) error {
	oldpath := ctx.FormValue("oldpath")
	if oldpath == "" {
		return fiber.NewError(http.StatusBadRequest, "oldpath cant be empty")
	}

	if oldpath == "/" {
		return fiber.NewError(http.StatusBadRequest, "'/' cant be renamed")
	}

	newpath := ctx.FormValue("newpath")
	if newpath == "" {
		return fiber.NewError(http.StatusBadRequest, "newpath cant be empty")
	}

	return nil
}

func (s *Server) Status(ctx *fiber.Ctx) error {
	objectID := ctx.Params("id")

	hex, err := primitive.ObjectIDFromHex(objectID)
	if err != nil {
		return err
	}

	metadata, exists := s.datastore.GetMetadataByID(ctx.Context(), hex)

	if !exists {
		return fiber.NewError(http.StatusNotFound, "file not found")
	}

	state := "done"
	if metadata.TotalFragments != len(metadata.Fragments) {
		state = "in progress"
	}

	marshal, err := json.Marshal(struct {
		State             string
		TotalFragments    int
		UploadedFragments int
	}{
		State:             state,
		TotalFragments:    metadata.TotalFragments,
		UploadedFragments: len(metadata.Fragments),
	})
	if err != nil {
		return err
	}

	if err := ctx.Send(marshal); err != nil {
		ctx.Status(http.StatusInternalServerError)
		return nil
	}

	return nil
}
