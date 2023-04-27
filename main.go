package main

import (
	"lintech/rego/game"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	//debug.SetMaxThreads(20000)
	// run the game
	game.NewGame()
}
