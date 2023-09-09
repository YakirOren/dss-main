package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"time"

	"dss-main/config"
	"dss-main/server/rabbit"

	"github.com/dustin/go-humanize"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"github.com/yakiroren/dss-common/db"
	"github.com/yakiroren/dss-common/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Server struct {
	datastore    db.DataStore
	fragmentSize int64
	Publisher    rabbit.Config
}

func NewServer(conf *config.Config, datastore db.DataStore) (*Server, error) {
	return &Server{
		Publisher:    conf.Publisher,
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
		if closeErr := src.Close(); closeErr != nil {
			log.Error(closeErr)
		}
	}(src)

	totalFragments := int(math.Ceil(float64(file.Size) / float64(s.fragmentSize)))

	filename := s.fixFilename(ctx.Context(), file.Filename, targetPath)

	fileID, err := s.datastore.WriteFile(ctx.Context(), models.FileMetadata{
		Id:             primitive.NewObjectID(),
		FileName:       filename,
		FileSize:       file.Size,
		CurrentSize:    0,
		CreationTime:   time.Now().Unix(),
		Tags:           []string{},
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

	_ = ctx.Status(http.StatusCreated).SendString(fileID)

	return nil
}

func (s *Server) fragment(totalFragments int, src io.Reader, id string) (bool, error) {
	logger := log.New()

	pub, err := rabbit.New(s.Publisher, logger)
	if err != nil {
		return false, fmt.Errorf("failed to create rabbit publisher: %w", err)
	}

	content := &bytes.Buffer{}
	log.Info("Total fragments ", totalFragments)

	pub.NotifyConsumers()

	for i := 1; i <= totalFragments; i++ {
		_, err = io.CopyN(content, src, s.fragmentSize)
		if errors.Is(err, io.EOF) {
			log.Debug("got EOF")
		} else if err != nil {
			log.Error(err)
			return false, fiber.ErrInternalServerError
		}

		if err = pub.PushMessage(id, i, content.Bytes()); err != nil {
			log.Error(err)
			return false, fiber.ErrInternalServerError
		}

		log.Debug("pushed fragment number ", i)

		content.Reset()
	}

	pub.Close()

	return true, nil
}
