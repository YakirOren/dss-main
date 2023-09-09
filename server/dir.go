package server

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"

	"github.com/dustin/go-humanize"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"github.com/yakiroren/dss-common/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DisplayMetadata struct {
	ID           interface{} `json:"id"`
	FileName     string      `json:"name"`
	Size         string      `json:"size"`
	IsDirectory  bool        `json:"directory"`
	IsProcessing bool        `json:"processing"`
	Path         string      `json:"path"`
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
			ID:           file.Id,
			FileName:     file.FileName,
			Size:         humanize.IBytes(uint64(file.FileSize)),
			IsDirectory:  file.IsDirectory,
			IsProcessing: file.IsHidden,
			Path:         filepath.Join(path, file.FileName),
		})
	}
	if jsonEncodeErr := ctx.JSON(list); jsonEncodeErr != nil {
		return fiber.ErrInternalServerError
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

func (s *Server) Move(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	metadata, exists := s.datastore.GetMetadataByID(ctx.Context(), id)
	if !exists {
		return fiber.NewError(http.StatusNotFound, "file not found")
	}

	newpath := ctx.FormValue("newpath")
	if newpath == "" {
		return fiber.NewError(http.StatusBadRequest, "newpath cant be empty")
	}

	valid := validatePath(newpath)
	if !valid {
		return fiber.NewError(http.StatusBadRequest, "the provided path is not valid")
	}

	// TODO: validate that the target path exists

	err := s.datastore.UpdateField(ctx.Context(), metadata.Id.Hex(), "path", newpath)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) Rename(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	metadata, exists := s.datastore.GetMetadataByID(ctx.Context(), id)
	if !exists {
		return fiber.NewError(http.StatusNotFound, "file not found")
	}

	newName := ctx.FormValue("new_name")
	if newName == "" {
		return fiber.NewError(http.StatusBadRequest, "new_name cant be empty")
	}

	newName = s.fixFilename(ctx.Context(), newName, metadata.Path)

	err := s.datastore.UpdateField(ctx.Context(), metadata.Id.Hex(), "name", newName)
	if err != nil {
		return err
	}

	return nil
}
