package iregoter

import (
	"image/color"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
)

type Position struct {
	X float64
	Y float64
	Z float64
}
type Entity struct {
	RgId            ID
	RgType          RegoterEnum
	Position        Position
	Scale           float64
	Anchor          raycaster.SpriteAnchor
	Angle           RotateAngle
	Pitch           PitchAngle
	Velocity        float64
	CollisionRadius float64
	CollisionHeight float64
	MapColor        color.RGBA
	Collidable      bool
}

func (e *Entity) Pos() *geom.Vector2 {
	return &geom.Vector2{X: e.Position.X, Y: e.Position.Y}
}

func (e *Entity) PosZ() float64 {
	return e.Position.Z
}
