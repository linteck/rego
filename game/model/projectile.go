package model

import (
	"image/color"
	"lintech/rego/game/loader"
	"lintech/rego/iregoter"

	"github.com/harbdog/raycaster-go"
)

type Projectile struct {
	rgData iregoter.RegoterData
	// Ricochets    int
	lifespan int
	harm     int
	effect   *Effect
}

func NewProjectile(di iregoter.DrawInfo,
	scale float64, collision iregoter.CollisionSpace, velocity float64,
	effect *Effect, harm int,
) *Projectile {
	//loadCrosshairsResource()
	entity := iregoter.Entity{
		RgId:            RgIdGenerator.GenId(),
		RgType:          iregoter.RegoterEnumProjectile,
		Scale:           scale,
		Velocity:        velocity,
		MapColor:        color.RGBA{0, 0, 0, 0},
		Anchor:          raycaster.AnchorCenter,
		CollisionRadius: collision.CollisionRadius,
		CollisionHeight: collision.CollisionHeight,
	}
	t := &Projectile{
		rgData: iregoter.RegoterData{
			Entity:   entity,
			DrawInfo: di,
		},
		effect:   effect,
		harm:     harm,
		lifespan: 100000,
	}

	return t
}

func (c *Projectile) UpdateTick(cu iregoter.RgTxMsgbox) {

}

func (c *Projectile) UpdateData(cu iregoter.RgTxMsgbox, rgEntity iregoter.Entity,
	rgState iregoter.RegoterState) bool {

	c.rgData.Entity = rgEntity
	if rgState.HasCollision || rgEntity.Position.Z <= 0 {
		c.effect.Spawn(cu, c.rgData)
		return false
	}
	return true
}

func (p Projectile) Spawn(coreMsgbox chan<- iregoter.IRegoterEvent,
	w iregoter.RegoterData) *Regoter[*Projectile] {
	n := p
	n.rgData.Entity.ParentId = w.Entity.RgId
	n.rgData.Entity.Position = w.Entity.Position
	n.rgData.Entity.Angle = w.Entity.Angle
	n.rgData.Entity.Pitch = w.Entity.Pitch
	r := NewRegoter(coreMsgbox, &n)
	return r
}

func (c *Projectile) SetConfig(cfg iregoter.GameCfg) {
}

func (c *Projectile) GetData() iregoter.RegoterData {
	return c.rgData
}

func NewProjectileChargedBolt(effect *Effect) *Projectile {
	// preload projectile sprites
	chargedBoltImg := loader.GetSpriteFromFile("charged_bolt_sheet.png")
	chargedBoltWidth := chargedBoltImg.Bounds().Dx()
	chargedBoltCols, chargedBoltRows := 6, 1
	chargedBoltScale := 0.3
	di := iregoter.DrawInfo{
		Img:           chargedBoltImg,
		Columns:       chargedBoltCols,
		Rows:          chargedBoltRows,
		AnimationRate: 1,
	}
	// in pixels, radius to use for collision testing
	chargedBoltPxRadius := 50.0
	chargedBoltCollisionRadius := (chargedBoltScale * chargedBoltPxRadius) / (float64(chargedBoltWidth) / float64(chargedBoltCols))
	chargedBoltCollisionHeight := 2 * chargedBoltCollisionRadius
	collision := iregoter.CollisionSpace{chargedBoltCollisionRadius, chargedBoltCollisionHeight}
	chargedBoltVelocity := 6.0 // Velocity (as distance travelled/second)
	chargedBoltProjectile := NewProjectile(di,
		chargedBoltScale, collision, chargedBoltVelocity, effect, 1)

	return chargedBoltProjectile
}
