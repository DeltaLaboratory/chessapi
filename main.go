package main

import (
	"chessapi/server"
)

func main() {
	if server.NewServer().Run() != nil {
		panic("Failed to run server")
	}
}
