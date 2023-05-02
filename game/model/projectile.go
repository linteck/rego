package model

import (
	"image/color"
	"lintech/rego/iregoter"
	"math"

	"github.com/harbdog/raycaster-go"
	"github.com/jinzhu/copier"

	"github.com/hajimehoshi/ebiten/v2"
)

type Projectile struct {
	rgId         iregoter.ID
	entity       iregoter.Entity
	di           iregoter.DrawInfo
	cfg          iregoter.GameCfg
	Ricochets    int
	Lifespan     float64
	ImpactEffect Effect
}

func NewProjectile(
	scale float64, img *ebiten.Image, mapColor color.RGBA, columns, rows int,
	anchor raycaster.SpriteAnchor, collisionRadius, collisionHeight float64,
) *Projectile {
	id := RgIdGenerator.GenId()
	entity := iregoter.Entity{
		Scale:           scale,
		MapColor:        mapColor,
		CollisionRadius: collisionRadius,
		CollisionHeight: collisionHeight,
		Anchor:          anchor,
	}
	di := iregoter.DrawInfo{
		ImgLayer:     iregoter.ImgLayerSprite,
		Img:          img,
		Columns:      columns,
		Rows:         rows,
		Illumination: 5000,
	}
	p := &Projectile{
		rgId:         id,
		entity:       entity,
		di:           di,
		Ricochets:    0,
		Lifespan:     math.MaxFloat64,
		ImpactEffect: Effect{},
	}

	// // projectiles should not be convergence capable by player focal point
	// p.Focusable = false

	return p
}

func NewAnimatedProjectile(
	scale float64, animationRate int, img *ebiten.Image, mapColor color.RGBA, columns, rows int,
	anchor raycaster.SpriteAnchor, collisionRadius, collisionHeight float64,
) *Projectile {
	p := NewProjectile(scale, img, mapColor, columns, rows, anchor, collisionRadius, collisionHeight)
	p.di.AnimationRate = animationRate

	return p
}

func (p *Projectile) Update(cu iregoter.RgTxMsgbox, rgEntity *iregoter.Entity,
	playEntiry *iregoter.Entity, HasCollision bool) {
	if HasCollision || rgEntity.Position.Z <= 0 {
		NewEffect(rgEntity.Position.X, rgEntity.Position.Y, rgEntity.Position.Z,
			rgEntity.Angle, rgEntity.Pitch)
		//Todo Delete Projectile
	}
	// Send a Move event
}

// func (g *Game) addProjectile(projectile *model.Projectile) {
// 	g.projectiles[projectile] = struct{}{}
// }

// func (g *Game) deleteProjectile(projectile *model.Projectile) {
// 	delete(g.projectiles, projectile)
// }

func SpawnProjectile(w Weapon, x float64, y float64, z float64, angle iregoter.RotateAngle, pitch iregoter.PitchAngle,
	spawnedBy *iregoter.Entity) *Projectile {
	p := &Projectile{}
	copier.Copy(p, w.projectile)
	p.entity.Position.X = x
	p.entity.Position.Y = y
	p.entity.Position.Z = z
	p.entity.Angle = angle
	p.entity.Pitch = pitch

	// convert velocity from distance/second to distance per tick
	p.entity.Velocity = w.projectileVelocity / float64(ebiten.TPS())

	// keep track of what spawned it
	p.entity.ParentId = w.entity.RgId

	return p
}
