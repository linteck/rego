package main

import (
	"lintech/rego/game/model"
	"log"
	_ "runtime/pprof"
)

func main() {
	// run the game
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	g := model.CreateGame(1)
	g.Run()
}
