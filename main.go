package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"chessapi/server"
)

func main() {
	// download stockfish binary
	downloadStockfish()

	if server.NewServer().Run() != nil {
		panic("Failed to run server")
	}
}

func downloadStockfish() {
	// https://bin.deltalab.group/stockfish-ubuntu-x86-64-avx2

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
