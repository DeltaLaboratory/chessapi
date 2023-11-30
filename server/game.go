package server

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/notnil/chess"
)

func (server *Server) Place(ctx *fiber.Ctx) error {
	userId, ok := server.userStore.Load(ctx.Query("token"))
	if !ok {
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	room, ok := server.roomStore.Load(ctx.Params("id"))
	if !ok {
		return ctx.SendStatus(fiber.StatusNotFound)
	}

	roomT := room.(*Room)

	player := "w"
	if roomT.BlackId == ctx.Query("token") {
		player = "b"
	}

	fmt.Printf("Player %s(%s) move %s\n", userId, player, ctx.Params("move"))

	if roomT.game.Position().Turn().String() != player {
		fmt.Printf("Player %s(%s) move %s - not it's turn\n", userId, player, ctx.Params("move"))
		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	if err := roomT.game.MoveStr(ctx.Params("move")); err != nil {
		fmt.Printf("Player %s(%s) move %s - %s\n", userId, player, ctx.Params("move"), err.Error())
		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	return ctx.Status(fiber.StatusOK).JSON(PlaceResult{
		Outcome: roomT.game.Outcome().String(),
		Method:  roomT.game.Method().String(),
		FEN:     roomT.game.FEN(),
	})
}

func (server *Server) CurrentTurn(ctx *fiber.Ctx) error {
	room, ok := server.roomStore.Load(ctx.Params("id"))
	if !ok {
		return ctx.SendStatus(fiber.StatusNotFound)
	}

	roomT := room.(*Room)

	return ctx.Status(fiber.StatusOK).JSON(CurrentTurnResponse{
		Turn: roomT.game.Position().Turn().String(),
	})
}

func (server *Server) CurrentBoard(ctx *fiber.Ctx) error {
	room, ok := server.roomStore.Load(ctx.Params("id"))
	if !ok {
		return ctx.SendStatus(fiber.StatusNotFound)
	}

	roomT := room.(*Room)

	return ctx.Status(fiber.StatusOK).JSON(CurrentBoardResponse{
		FEN:     roomT.game.FEN(),
		Outcome: roomT.game.Outcome().String(),
		Method:  roomT.game.Method().String(),
	})
}

func (server *Server) RemainTime(ctx *fiber.Ctx) error {
	_, ok := server.userStore.Load(ctx.Query("token"))
	if !ok {
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	room, ok := server.roomStore.Load(ctx.Params("id"))
	if !ok {
		return ctx.SendStatus(fiber.StatusNotFound)
	}

	roomT := room.(*Room)

	if roomT.WhiteId == ctx.Query("token") {
		return ctx.Status(fiber.StatusOK).JSON(RemainTimeResponse{
			Player:   roomT.timer.White.String(),
			Opponent: roomT.timer.Black.String(),
		})
	}
	if roomT.BlackId == ctx.Query("token") {
		return ctx.Status(fiber.StatusOK).JSON(RemainTimeResponse{
			Player:   roomT.timer.Black.String(),
			Opponent: roomT.timer.White.String(),
		})
	}
	return ctx.SendStatus(fiber.StatusBadRequest)
}

func (server *Server) Resign(ctx *fiber.Ctx) error {
	userId, ok := server.userStore.Load(ctx.Query("token"))
	if !ok {
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	room, ok := server.roomStore.Load(ctx.Params("id"))
	if !ok {
		return ctx.SendStatus(fiber.StatusNotFound)
	}

	roomT := room.(*Room)

	player := "w"
	if roomT.BlackId == ctx.Query("token") {
		player = "b"
	}

	fmt.Printf("Player %s(%s) resign\n", userId, player)

	switch player {
	case "w":
		roomT.game.Resign(chess.White)
	case "b":
		roomT.game.Resign(chess.Black)
	}

	return ctx.Status(fiber.StatusOK).JSON(PlaceResult{
		Outcome: roomT.game.Outcome().String(),
		Method:  roomT.game.Method().String(),
		FEN:     roomT.game.FEN(),
	})
}

type PlaceResult struct {
	Outcome string `json:"outcome"`
	Method  string `json:"method"`
	FEN     string `json:"fen"`
}

type CurrentTurnResponse struct {
	Turn string `json:"turn"`
}

type CurrentBoardResponse struct {
	FEN     string `json:"fen"`
	Outcome string `json:"outcome"`
	Method  string `json:"method"`
}

type RemainTimeResponse struct {
	Player   string `json:"player"`
	Opponent string `json:"opponent"`
}
