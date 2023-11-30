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
	if err := os.Chmod(fmt.Sprintf("%s/stockfish", os.Getenv("KO_DATA_PATH")), 0777); err != nil {
		fmt.Printf("Failed to chmod stockfish: %s\n", err.Error())
	}

	// check stockfish executable
	if stat, err := os.Stat(fmt.Sprintf("%s/stockfish", os.Getenv("KO_DATA_PATH"))); err != nil {
		fmt.Printf("Failed to stat stockfish: %s\n", err.Error())
	} else {
		if stat.Mode()&0111 == 0 {
			fmt.Printf("Stockfish is not executable\n")
		}

		if err = exec.Command(fmt.Sprintf("%s/stockfish", os.Getenv("KO_DATA_PATH")), "--help").Run(); err != nil {
			fmt.Printf("Failed to run stockfish: %s\n", err.Error())
		}
	}
}
