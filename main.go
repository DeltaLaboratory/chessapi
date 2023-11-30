package main

import (
	"fmt"
	"os"
	"os/exec"

	"chessapi/server"
)

func main() {
	setStockfishPermissions()

	if server.NewServer().Run() != nil {
		panic("Failed to run server")
	}
}

func setStockfishPermissions() {
	if err := os.Chmod(fmt.Sprintf("%sstockfish", os.Getenv("KO_DATA_PATH")), 0777); err != nil {
		panic(err)
	}

	// check stockfish executable
	if stat, err := os.Stat(fmt.Sprintf("%sstockfish", os.Getenv("KO_DATA_PATH"))); err != nil {
		panic(err)
	} else {
		if stat.Mode()&0111 == 0 {
			panic("Stockfish is not executable")
		}

		if err = exec.Command(fmt.Sprintf("%sstockfish", os.Getenv("KO_DATA_PATH")), "--help").Run(); err != nil {
			fmt.Printf("Failed to run stockfish: %s\n", err.Error())
		}
	}
}
