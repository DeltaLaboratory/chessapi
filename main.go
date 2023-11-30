package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	"chessapi/server"
)

func main() {
	// download stockfish binary
	downloadStockfish()
	setStockfishPermissions()

	if server.NewServer().Run() != nil {
		panic("Failed to run server")
	}
}

func downloadStockfish() {
	// https://bin.deltalab.group/stockfish-ubuntu-x86-64-avx2

	if _, err := os.Stat("/data/stockfish"); err == nil {
		fmt.Printf("Stockfish already downloaded\n")
		return
	}

	fmt.Printf("Downloading stockfish...\n")

	req, err := http.NewRequest("GET", "https://bin.deltalab.group/stockfish-ubuntu-x86-64-avx2", nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		panic("Failed to download stockfish")
	}

	// Write the body to file
	stockfish, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile("/data/stockfish", stockfish, 0755); err != nil {
		panic(err)
	}

	fmt.Printf("Downloaded stockfish\n")
}

func setStockfishPermissions() {
	if err := os.Chmod("/data/stockfish", 0777); err != nil {
		panic(err)
	}

	// check stockfish executable
	if stat, err := os.Stat("/data/stockfish"); err != nil {
		panic(err)
	} else {
		if stat.Mode()&0111 == 0 {
			panic("Stockfish is not executable")
		}

		// add stockfish to path
		if err := os.Setenv("PATH", fmt.Sprintf("%s:/data", os.Getenv("PATH"))); err != nil {
			panic(err)
		}

		if err = exec.Command("stockfish", "--help").Run(); err != nil {
			fmt.Printf("Failed to run stockfish: %s\n", err.Error())
		}
	}

	// check stockfish can be called with --help

	fmt.Printf("Stockfish permissions set\n")
}
