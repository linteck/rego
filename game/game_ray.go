package game

import (
	"lintech/rego/iregoter"
	"math/rand"
	"os"

	"image/color"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Game - This is the main type for your game.

func drawSpriteBox(screen *ebiten.Image, sprite *iregoter.Sprite) {
	r := sprite.ScreenRect()
	if r == nil {
		return
	}

	minX, minY := float32(r.Min.X), float32(r.Min.Y)
	maxX, maxY := float32(r.Max.X), float32(r.Max.Y)

	vector.StrokeRect(screen, minX, minY, maxX-minX, maxY-minY, 1, color.RGBA{255, 0, 0, 255}, false)
}

func drawSpriteIndicator(screen *ebiten.Image, sprite *iregoter.Sprite) {
	r := sprite.ScreenRect()
	if r == nil {
		return
	}

	dX, dY := float32(r.Dx())/8, float32(r.Dy())/8
	midX, minY := float32(r.Max.X)-float32(r.Dx())/2, float32(r.Min.Y)-dY

	vector.StrokeLine(screen, midX, minY+dY, midX-dX, minY, 1, color.RGBA{0, 255, 0, 255}, false)
	vector.StrokeLine(screen, midX, minY+dY, midX+dX, minY, 1, color.RGBA{0, 255, 0, 255}, false)
	vector.StrokeLine(screen, midX-dX, minY, midX+dX, minY, 1, color.RGBA{0, 255, 0, 255}, false)
}

// func (g *Game) updateSprites() {
// 	// Testing animated sprite movement
// 	for s := range g.sprites {
// 		if s.Velocity != 0 {
// 			vLine := geom.LineFromAngle(s.Position.X, s.Position.Y, s.Angle, s.Velocity)

// 			xCheck := vLine.X2
// 			yCheck := vLine.Y2
// 			zCheck := s.PositionZ

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

func exit(rc int) {
	// TODO: any cleanup?
	os.Exit(rc)
}
