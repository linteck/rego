package model

import (
	"image/color"
	"lintech/rego/game/loader"
	"log"

	"github.com/harbdog/raycaster-go"
)

type ProjectileTemplate struct {
	rgData RegoterData
	// Ricochets    int
	lifespan int
	harm     int
	effect   *EffectTemplate
}

type Projectile struct {
	Reactor
	ProjectileTemplate
}

func (r *Projectile) ProcessMessage(m ReactorEventMessage) error {
	// log.Print(fmt.Sprintf("(%v) recv %T", r.thing.GetData().Entity.RgId, e))
	switch m.event.(type) {
	case EventUpdateTick:
		r.eventHandleUpdateTick(m.sender, m.event.(EventUpdateTick))
	case EventUpdateData:
		r.eventHandleUpdateData(m.sender, m.event.(EventUpdateData))
	case EventCfgChanged:
		r.eventHandleCfgChanged(m.sender, m.event.(EventCfgChanged))
	default:
		r.eventHandleUnknown(m.sender, m.event)
	}
	return nil
}

func (r *Projectile) eventHandleUnknown(sender RcTx, e IReactorEvent) error {
	log.Fatal("Unknown event:", e)
	return nil
}
func (c *Projectile) eventHandleUpdateTick(sender RcTx, e EventUpdateTick) {

	if e.RgState.HasCollision {
		c.effect.Spawn(sender)
		m := ReactorEventMessage{c.tx, EventUnregisterRegoter{RgId: c.rgData.Entity.RgId}}
		sender <- m
		c.running = false
	} else {
		m := ReactorEventMessage{c.tx, EventMovement{RgId: c.rgData.Entity.RgId,
			Move: Movement{Velocity: c.rgData.Entity.Velocity}}}
		sender <- m
	}

}

func (c *Projectile) eventHandleUpdateData(sender RcTx, e EventUpdateData) {
}

func (p *ProjectileTemplate) Spawn(coreTx RcTx, aim Entity) RcTx {
	return NewProjectile(coreTx, p, aim)
}

func (c *Projectile) eventHandleCfgChanged(sender RcTx, e EventCfgChanged) {
}

func NewProjectile(coreTx RcTx, pt *ProjectileTemplate, aim Entity) RcTx {
	p := &Projectile{
		Reactor:            NewReactor(),
		ProjectileTemplate: *pt,
	}
	// Don't use ID of Template
	p.rgData.Entity.RgId = RgIdGenerator.GenId()
	p.rgData.Entity.ParentId = aim.RgId
	p.rgData.Entity.Position = aim.Position
	p.rgData.Entity.Angle = aim.Angle
	p.rgData.Entity.Pitch = aim.Pitch
	go p.Reactor.Run(p)
	m := ReactorEventMessage{p.tx, EventRegisterRegoter{p.tx, p.rgData}}
	coreTx <- m
	return p.tx
}

func NewProjectileTemplate(di DrawInfo,
	scale float64, collision CollisionSpace, velocity float64,
	effect *EffectTemplate, harm int,
) *ProjectileTemplate {
	//loadCrosshairsResource()
	entity := Entity{
		RgId:            RgIdGenerator.GenId(),
		RgType:          RegoterEnumProjectile,
		Scale:           scale,
		Velocity:        velocity,
		MapColor:        color.RGBA{0, 0, 255, 200},
		Anchor:          raycaster.AnchorCenter,
		CollisionRadius: collision.CollisionRadius,
		CollisionHeight: collision.CollisionHeight,
	}
	t := &ProjectileTemplate{
		rgData: RegoterData{
			Entity:   entity,
			DrawInfo: di,
		},
		effect:   effect,
		harm:     harm,
		lifespan: 100000,
	}

	return t
}

func ProjectileChargedBolt(effect *EffectTemplate) *ProjectileTemplate {
	// preload projectile sprites
	chargedBoltImg := loader.GetSpriteFromFile("charged_bolt_sheet.png")
	chargedBoltWidth := chargedBoltImg.Bounds().Dx()
	chargedBoltCols, chargedBoltRows := 6, 1
	chargedBoltScale := 0.3
	di := DrawInfo{
		Img:           chargedBoltImg,
		Columns:       chargedBoltCols,
		Rows:          chargedBoltRows,
		AnimationRate: 1,
	}
	// in pixels, radius to use for collision testing
	chargedBoltPxRadius := 50.0
	chargedBoltCollisionRadius := (chargedBoltScale * chargedBoltPxRadius) / (float64(chargedBoltWidth) / float64(chargedBoltCols))
	chargedBoltCollisionHeight := 2 * chargedBoltCollisionRadius
	collision := CollisionSpace{chargedBoltCollisionRadius, chargedBoltCollisionHeight}
	// Debug
	// chargedBoltVelocity := 6.0 // Velocity (as distance travelled/second)
	chargedBoltVelocity := 0.2 // Velocity (as distance travelled/second)
	chargedBoltProjectile := NewProjectileTemplate(di,
		chargedBoltScale, collision, chargedBoltVelocity, effect, 1)

	return chargedBoltProjectile
}

func ProjectileRedBolt(effect *EffectTemplate) *ProjectileTemplate {
	// preload projectile sprites
	redBoltImg := loader.GetSpriteFromFile("red_bolt.png")
	redBoltWidth := redBoltImg.Bounds().Dx()
	redBoltCols, redBoltRows := 1, 1
	redBoltScale := 0.25
	di := DrawInfo{
		Img:           redBoltImg,
		Columns:       redBoltCols,
		Rows:          redBoltRows,
		AnimationRate: 1,
	}
	// in pixels, radius to use for collision testing
	redBoltPxRadius := 4.0
	redBoltCollisionRadius := (redBoltScale * redBoltPxRadius) / (float64(redBoltWidth) / float64(redBoltCols))
	redBoltCollisionHeight := 2 * redBoltCollisionRadius
	collision := CollisionSpace{redBoltCollisionRadius, redBoltCollisionHeight}
	redBoltVelocity := 6.0 // Velocity (as distance travelled/second)
	redBoltProjectile := NewProjectileTemplate(di,
		redBoltScale, collision, redBoltVelocity, effect, 1)

	return redBoltProjectile
}
