package game

import (
	"lintech/rego/game/model"
	"log"
	"testing"
	"time"
)

func TestGame(t *testing.T) {
	// run the game
	go func() {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		g := model.CreateGame(10)
		g.Run()
	}()
	time.Sleep(10 * time.Second)
}

func BenchmarkGame(b *testing.B) {
	// run the game
	go func() {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		g := model.CreateGame(100)
		g.Run()
	}()
	time.Sleep(100 * time.Second)

}
