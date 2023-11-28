package server

import (
	"fmt"
	"slices"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/notnil/chess"
	"github.com/notnil/chess/uci"
)

func (server *Server) CreateRoom(ctx *fiber.Ctx) error {
	roomname := ctx.Params("name")
	userToken := ctx.Query("token")

	userId, ok := server.userStore.Load(userToken)
	if !ok {
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	if roomname == "" {
		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	roomId := uuid.New().String()
	room := &Room{
		Id:      roomId,
		Name:    roomname,
		White:   userId.(string),
		WhiteId: userToken,
		game:    chess.NewGame(),

		created:    time.Now(),
		started:    time.Now(),
		lastPlaced: time.Time{},
		timerWhite: time.Minute * 10,
		timerBlack: time.Minute * 10,
	}

	server.roomStore.Store(roomId, room)

	fmt.Printf("Created room %s with name %s and user %s(%s)\n", roomId, roomname, userId.(string), userToken)

	if ctx.Query("stockfish") == "enable" {
		fmt.Printf("Stockfish enabled for room %s\n", roomId)
		go func() {
			time.Sleep(time.Second * 5)

			room.Black = "stockfish"
			room.BlackId = "stockfish"

			eng, err := uci.New("/data/stockfish")
			if err != nil {
				panic(err)
			}
			defer eng.Close()
			// initialize uci with new game
			if err := eng.Run(uci.CmdUCI, uci.CmdIsReady, uci.CmdUCINewGame); err != nil {
				panic(err)
			}
			for {
				if room.game.Position().Turn() == chess.Black {
					cmdPos := uci.CmdPosition{Position: room.game.Position()}
					cmdGo := uci.CmdGo{MoveTime: time.Second}
					if err := eng.Run(cmdPos, cmdGo); err != nil {
						panic(err)
					}
					move := eng.SearchResults().BestMove
					if err := room.game.Move(move); err != nil {
						panic(err)
					}
					fmt.Printf("Stockfish move %s\n", move)

					room.timerBlack -= time.Since(room.lastPlaced)
					room.lastPlaced = time.Now()
				}
				time.Sleep(time.Second)
			}
		}()
	}

	return ctx.Status(fiber.StatusOK).JSON(CreateRoomResponse{
		Id: roomId,
	})
}

func (server *Server) InfoRoom(ctx *fiber.Ctx) error {
	roomId := ctx.Params("id")
	room, ok := server.roomStore.Load(roomId)

	if !ok {
		fmt.Printf("Room %s not found\n", roomId)
		return ctx.SendStatus(fiber.StatusNotFound)
	}

	return ctx.Status(fiber.StatusOK).JSON(room)
}

func (server *Server) JoinRoom(ctx *fiber.Ctx) error {
	server.mu.Lock()
	defer server.mu.Unlock()

	roomId := ctx.Params("id")
	userToken := ctx.Query("token")

	userId, ok := server.userStore.Load(userToken)
	if !ok {
		fmt.Printf("User %s not found\n", userToken)
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	room, ok := server.roomStore.Load(roomId)
	if !ok {
		fmt.Printf("Room %s not found\n", roomId)
		return ctx.SendStatus(fiber.StatusNotFound)
	}

	userIdT := userId.(string)
	roomT := room.(*Room)

	if roomT.WhiteId == userToken || roomT.BlackId == userToken {
		fmt.Printf("User %s already in room %s\n", userIdT, roomId)
		return ctx.SendStatus(fiber.StatusConflict)
	}

	if roomT.White == "" {
		roomT.White = userIdT
		roomT.WhiteId = userToken
	} else if roomT.Black == "" {
		roomT.Black = userIdT
		roomT.BlackId = userToken
	} else {
		fmt.Printf("Room %s is full\n", roomId)
		return ctx.SendStatus(fiber.StatusConflict)
	}

	fmt.Printf("User %s joined room %s\n", userIdT, roomId)

	return ctx.Status(fiber.StatusOK).JSON(roomT)
}

func (server *Server) ListRoom(ctx *fiber.Ctx) error {
	var rooms []*Room

	offset, _ := ctx.ParamsInt("offset")

	server.mu.Lock()
	defer server.mu.Unlock()

	server.roomStore.Range(func(key, value interface{}) bool {
		rooms = append(rooms, value.(*Room))
		return true
	})

	slices.SortStableFunc(rooms, func(i, j *Room) int {
		if i.created.Before(j.created) {
			return -1
		}
		if i.created.After(j.created) {
			return 1
		}
		return 0
	})

	if offset > len(rooms) {
		offset = len(rooms)
	}

	if len(rooms[offset:]) == 0 {
		return ctx.Status(fiber.StatusOK).SendString("[]")
	}

	for _, room := range rooms[offset:] {
		if room.BlackId != "" {
			room.Name = fmt.Sprintf("%s (playing)", room.Name)
		}
		if room.game.Outcome() != chess.NoOutcome {
			room.Name = fmt.Sprintf("%s (finished)", room.Name)
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(rooms[offset:])
}

type CreateRoomResponse struct {
	Id string `json:"room_id"`
}

type Room struct {
	Id      string `json:"room_id"`
	Name    string `json:"name"`
	White   string `json:"white"`
	Black   string `json:"black"`
	WhiteId string `json:"white_id"`
	BlackId string `json:"black_id"`

	game *chess.Game

	created    time.Time
	started    time.Time
	lastPlaced time.Time
	timerWhite time.Duration
	timerBlack time.Duration
}
