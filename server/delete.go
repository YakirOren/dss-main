package server

import (
	"github.com/gofiber/fiber/v2"
	"net/http"
)

func (s *Server) Delete(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	success := s.datastore.Delete(ctx.Context(), id)
	if !success {
		return fiber.NewError(http.StatusBadRequest, "could not delete")
	}

	ctx.Status(http.StatusOK)

	return nil
}
