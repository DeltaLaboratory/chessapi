package server

import "github.com/gofiber/fiber/v2/middleware/logger"

func (server *Server) route() {
	server.app.Use(logger.New())
	server.app.Get("/login/:id", server.Login)
	server.app.Get("/room/create/:name", server.CreateRoom)
	server.app.Get("/room/info/:id", server.InfoRoom)
	server.app.Get("/room/join/:id", server.JoinRoom)
	server.app.Get("/room/list", server.ListRoom)

	server.app.Get("/game/place/:id/:move", server.Place)
	server.app.Get("/game/turn/:id", server.CurrentTurn)
	server.app.Get("/game/timer/:id", server.RemainTime)
	server.app.Get("/game/board/:id", server.CurrentBoard)
	server.app.Get("/game/resign/:id", server.Resign)
}
