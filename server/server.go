package server

import (
	"bytes"
	"dss-main/config"
	"dss-main/server/rabbit"
	"errors"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"github.com/yakiroren/dss-common/db"
	"github.com/yakiroren/dss-common/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"math"
	"mime/multipart"
	"net/http"
)

type Server struct {
	datastore    db.DataStore
	fragmentSize int64
	publisher    *rabbit.Publisher
}

func NewServer(conf *config.Config, datastore db.DataStore) (*Server, error) {
	logger := log.New()

	pub, err := rabbit.New(conf.Publisher, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create rabbit publisher: %w", err)
	}

	return &Server{
		publisher:    pub,
		datastore:    datastore,
		fragmentSize: conf.FragmentSize,
	}, nil
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

	done, err := s.fragment(totalFragments, src, fileID)
	if !done {
		return err
	}

	ctx.Status(http.StatusCreated).SendString(fileID)
	return nil
}

func (s *Server) fragment(totalFragments int, src io.Reader, id string) (bool, error) {
	content := &bytes.Buffer{}
	log.Info("Total fragments ", totalFragments)

	err := s.publisher.NotifyConsumer()
	if err != nil {
		log.Error(err)
	}

	for i := 1; i <= totalFragments; i++ {
		_, err := io.CopyN(content, src, s.fragmentSize)
		if errors.Is(err, io.EOF) {
		} else if err != nil {
			log.Error(err)
			return false, fiber.ErrInternalServerError
		}

		if err = s.publisher.PushMessage(id, i, content.Bytes()); err != nil {
			log.Error(err)
			return false, fiber.ErrInternalServerError
		}

		log.Debug("pushed fragment number ", i)

		content.Reset()
	}

	return true, nil
}
