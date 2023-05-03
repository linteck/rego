package model

import (
	"lintech/rego/iregoter"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go"
)

type Effect struct {
	*iregoter.Sprite
	LoopCount int
}

func NewAnimatedEffect(
	x, y, scale float64, animationRate int, img *ebiten.Image, columns, rows int, anchor raycaster.SpriteAnchor, loopCount int,
) *Effect {
	// mapColor := color.RGBA{0, 0, 0, 0}
	e := &Effect{
		// Sprite:    iregoter.NewAnimatedSprite(x, y, scale, animationRate, img, mapColor, columns, rows, anchor, 0, 0),
		LoopCount: loopCount,
	}

	// // effects should not be convergence capable by player focal point
	// e.Sprite.Focusable = false

	// effects self illuminate so they do not get dimmed in dark conditions
	// e.Sprite.SetIllumination(5000)

	return e
}

func NewEffect(x float64, y float64, z float64, angle float64, pitch float64) *Effect {
	e := &Effect{}
	// s := &iregoter.Sprite{}
	// copier.Copy(e, p.ImpactEffect)
	// copier.Copy(s, p.ImpactEffect.Sprite)

	// e.Sprite = s
	// e.Position = &geom.Vector2{X: x, Y: y}
	// e.Position.Z = z
	// e.Angle = angle
	// e.Pitch = pitch

	// // keep track of what spawned it
	// e.Parent = p.Parent

	return e
}
