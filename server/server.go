package server

import (
	"sync"

	"github.com/gofiber/fiber/v2"
)

type Server struct {
	app       *fiber.App
	userStore sync.Map
	roomStore sync.Map
	mu        sync.Mutex
}

func NewServer() *Server {
	return &Server{
		app:       fiber.New(fiber.Config{Immutable: true}),
		userStore: sync.Map{},
		roomStore: sync.Map{},
		mu:        sync.Mutex{},
	}
}

func (server *Server) Run() error {
	server.route()
	return server.app.Listen(":80")
}
