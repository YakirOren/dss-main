package server

import (
	"encoding/json"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

const (
	Done     = "done"
	Progress = "in progress"
)

type Status struct {
	State             string `json:"State"`
	TotalFragments    int    `json:"TotalFragments"`
	UploadedFragments int    `json:"UploadedFragments"`
}

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

	marshal, err := json.Marshal(Status{
		State:             state,
		TotalFragments:    metadata.TotalFragments,
		UploadedFragments: len(metadata.Fragments),
	})
	if err != nil {
		return err
	}

	if err = ctx.Send(marshal); err != nil {
		return fiber.ErrInternalServerError
	}

	return nil
}
