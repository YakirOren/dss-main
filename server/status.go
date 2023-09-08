package server

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"net/http"
)

const (
	Done     = "done"
	Progress = "in progress"
)

func (s *Server) Status(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	metadata, exists := s.datastore.GetMetadataByID(ctx.Context(), id)

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
