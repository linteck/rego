package model

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
	RgName          string
	Position        Position
	Scale           float64
	Anchor          raycaster.SpriteAnchor
	Angle           float64
	Pitch           float64
	Velocity        float64
	Resistance      float64
	LastMoveRotate  float64
	CollisionRadius float64
	CollisionHeight float64
	MapColor        color.RGBA
	ParentId        ID
}

func (e *Entity) Pos() *geom.Vector2 {
	return &geom.Vector2{X: e.Position.X, Y: e.Position.Y}
}

func (e *Entity) PosZ() float64 {
	return e.Position.Z
}
