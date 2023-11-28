package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (server *Server) Login(ctx *fiber.Ctx) error {
	server.mu.Lock()
	defer server.mu.Unlock()

	id := ctx.Params("id")

	if id == "" {
		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	hash := uuid.New().String()

	server.userStore.Store(hash, id)

	return ctx.Status(fiber.StatusOK).JSON(LoginResponse{
		Token: hash,
	})
}

type LoginResponse struct {
	Token string `json:"token"`
}
