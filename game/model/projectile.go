package model

import (
	"image/color"
	"lintech/rego/iregoter"
	"math"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"

	"github.com/hajimehoshi/ebiten/v2"
)

type Projectile struct {
	*iregoter.Sprite
	Ricochets    int
	Lifespan     float64
	ImpactEffect Effect
}

func NewProjectile(
	x, y, scale float64, img *ebiten.Image, mapColor color.RGBA,
	anchor raycaster.SpriteAnchor, collisionRadius, collisionHeight float64,
) *Projectile {
	p := &Projectile{
		Sprite:       iregoter.NewSprite(x, y, scale, img, mapColor, anchor, collisionRadius, collisionHeight),
		Ricochets:    0,
		Lifespan:     math.MaxFloat64,
		ImpactEffect: Effect{},
	}

	// projectiles should not be convergence capable by player focal point
	p.Focusable = false

	// projectiles self illuminate so they do not get dimmed in dark conditions
	p.Sprite.SetIllumination(5000)

	return p
}

func NewAnimatedProjectile(
	x, y, scale float64, animationRate int, img *ebiten.Image, mapColor color.RGBA, columns, rows int,
	anchor raycaster.SpriteAnchor, collisionRadius, collisionHeight float64,
) *Projectile {
	p := &Projectile{
		Sprite:       iregoter.NewAnimatedSprite(x, y, scale, animationRate, img, mapColor, columns, rows, anchor, collisionRadius, collisionHeight),
		Ricochets:    0,
		Lifespan:     math.MaxFloat64,
		ImpactEffect: Effect{},
	}

	// projectiles should not be convergence capable by player focal point
	p.Focusable = false

	// projectiles self illuminate so they do not get dimmed in dark conditions
	p.Sprite.SetIllumination(5000)

	return p
}

func (p *Projectile) Update(cu iregoter.ChanRegoterUpdate, rgEntity *iregoter.Entity,
	playEntiry *iregoter.Entity, HasCollision bool, screenSize iregoter.ScreenSize) {
	if HasCollision || rgEntity.PositionZ <= 0 {
		NewEffect(rgEntity.Position.X, rgEntity.Position.Y, rgEntity.PositionZ,
			rgEntity.Angle, rgEntity.Pitch)
	}
	s := p.Sprite
	scale := p.Sprite.Scale()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(
		float64(screenSize.Width)/2-float64(s.W)*scale/2,
		float64(screenSize.Height)/2-float64(s.H)*scale/2,
	)

	info := iregoter.RegoterUpdatedImg{ImgOp: op, Sprite: s, Img: s.Texture(),
		Visiable: true, Deleted: false, Changed: true}

	cu <- info
	close(cu)
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
	s := &iregoter.Sprite{}
	copier.Copy(p, w.projectile)
	copier.Copy(s, w.projectile.Sprite)

	p.Sprite = s
	p.Position = &geom.Vector2{X: x, Y: y}
	p.PositionZ = z
	p.Angle = angle
	p.Pitch = pitch

	// convert velocity from distance/second to distance per tick
	p.Velocity = w.projectileVelocity / float64(ebiten.TPS())

	// keep track of what spawned it
	p.Parent = spawnedBy

	return p
}
