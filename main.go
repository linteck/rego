package main

import "lintech/rego/game/model"

func main() {
	// run the game
	g := model.CreateGame()
	g.Run()
}
