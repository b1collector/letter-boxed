package main

import (
	"fmt"
	"log"
	"os"

	"github.com/b1collector/letter-boxed/internal/letterbox"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Required arguments missing.")
	}
	letters := os.Args[1]

	board := letterbox.NewGameBoard(letters)
	board.LoadWordList("dictionary.txt")
	solution := board.GenerateSolution()
	fmt.Printf("%v\n", solution)
}
