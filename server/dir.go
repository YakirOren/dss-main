package server

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"github.com/yakiroren/dss-common/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"path/filepath"
)

type DisplayMetadata struct {
	Id           interface{} `json:"Id"`
	FileName     string      `json:"FileName"`
	Size         int64       `json:"Size"`
	IsDirectory  bool        `json:"IsDirectory"`
	IsProcessing bool        `json:"IsProcessing"`
}

func (s *Server) Dir(ctx *fiber.Ctx) error {
	path := ctx.Params("*", "/")

	if path != "/" {
		path = "/" + path
	}

	rawFiles, err := s.datastore.ListFiles(context.Background(), path)
	if err != nil {
		log.Error(err)
		return fiber.ErrInternalServerError
	}

	var list []DisplayMetadata

	for _, file := range rawFiles {
		if file.FileName == "/" {
			continue
		}

		list = append(list, DisplayMetadata{
			Id:           file.Id,
			FileName:     file.FileName,
			Size:         file.FileSize,
			IsDirectory:  file.IsDirectory,
			IsProcessing: file.IsHidden,
		})
	}
	if err := ctx.JSON(list); err != nil {
		ctx.Status(http.StatusInternalServerError)
		return nil
	}

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
