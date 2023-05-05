package game

import (
	"math/rand"

	_ "image/png"
)

// Game - This is the main type for your game.

// func (g *Game) updateSprites() {
// 	// Testing animated sprite movement
// 	for s := range g.sprites {
// 		if s.Velocity != 0 {
// 			vLine := geom.LineFromAngle(s.Position.X, s.Position.Y, s.Angle, s.Velocity)

// 			xCheck := vLine.X2
// 			yCheck := vLine.Y2
// 			zCheck := s.Position.Z

// 			newPos, isCollision, _ := g.getValidMove(s.Entity, xCheck, yCheck, zCheck, false)
// 			if isCollision {
// 				// for testing purposes, letting the sample sprite ping pong off walls in somewhat random direction
// 				s.Angle = randFloat(-math.Pi, math.Pi)
// 				s.Velocity = randFloat(0.01, 0.03)
// 			} else {
// 				s.Position = newPos
// 			}
// 		}
// 		s.Update(g.player.Position)
// 	}
// }

func randFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}
