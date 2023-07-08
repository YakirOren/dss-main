package server

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

const (
	Done     = "done"
	Progress = "in progress"
)

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

	state := Done
	if metadata.TotalFragments != len(metadata.Fragments) {
		state = Progress
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
