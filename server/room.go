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

		created: time.Now(),
		started: time.Now(),
		timer:   NewChessTime(time.Minute * 15),
	}

	server.roomStore.Store(roomId, room)

	fmt.Printf("Created room %s with name %s and user %s(%s)\n", roomId, roomname, userId.(string), userToken)

	if ctx.Query("stockfish") == "enable" {
		fmt.Printf("Stockfish enabled for room %s\n", roomId)
		go func() {
			room.Black = "stockfish"
			room.BlackId = "stockfish"

			eng, err := uci.New("stockfish")
			if err != nil {
				panic(err)
			}
			defer eng.Close()
			// initialize uci with new game
			if err := eng.Run(uci.CmdUCI, uci.CmdIsReady, uci.CmdUCINewGame); err != nil {
				panic(err)
			}
			fmt.Printf("Stockfish initialized\n")
			for {
				if room.game.Position().Turn() == chess.Black {
					fmt.Printf("Stockfish thinking...\n")
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

					room.timer.Update()
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
			rooms = remove(rooms, room)
		}
		if room.game.Outcome() != chess.NoOutcome {
			rooms = remove(rooms, room)
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

	created time.Time
	started time.Time
	timer   *ChessTime
}

type ChessTime struct {
	White time.Duration `json:"white"`
	Black time.Duration `json:"black"`

	placed time.Time
	turn   chess.Color
}

func (ct *ChessTime) Update() {
	ct.UpdateTime()
	ct.turn = ct.turn.Other()
}

func (ct *ChessTime) UpdateTime() {
	switch ct.turn {
	case chess.White:
		ct.White -= time.Since(ct.placed)
	case chess.Black:
		ct.Black -= time.Since(ct.placed)
	}
	ct.placed = time.Now()
}

func (ct *ChessTime) Worker() {
	for {
		ct.UpdateTime()
		time.Sleep(time.Millisecond * 500)
		if ct.White <= 0 || ct.Black <= 0 {
			break
		}
	}
}

func NewChessTime(dur time.Duration) *ChessTime {
	ct := &ChessTime{
		White:  dur,
		Black:  dur,
		turn:   chess.White,
		placed: time.Now(),
	}
	go ct.Worker()
	return ct
}

func remove[S interface{ ~[]E }, E comparable](slice S, s E) S {
	loc := slices.Index(slice, s)
	if loc == -1 {
		return slice
	}
	return append(slice[:loc], slice[loc+1:]...)
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		return "00:00"
	}
	return fmt.Sprintf("%02d:%02d", int(d.Minutes()), int(d.Seconds())%60)
}
