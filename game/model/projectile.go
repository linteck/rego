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
		MapColor:        color.RGBA{0, 0, 0, 0},
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

func (c *Projectile) eventHandleUpdateTick(sender RcTx, e EventUpdateTick) {
	if e.RgState.HasCollision {
		c.effect.Spawn(sender, c.rgData)
		m := ReactorEventMessage{c.tx, EventUnregisterRegoter{RgId: c.rgData.Entity.RgId}}
		sender <- m
		c.running = false
	}

}

func (c *Projectile) eventHandleUpdateData(sender RcTx, e EventUpdateData) {
}

func (p *ProjectileTemplate) Spawn(coreTx RcTx, w RegoterData) RcTx {
	return NewProjectile(coreTx, p, w.Entity)
}

func (c *Projectile) eventHandleCfgChanged(sender RcTx, e EventCfgChanged) {
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
	chargedBoltVelocity := 6.0 // Velocity (as distance travelled/second)
	chargedBoltProjectile := NewProjectileTemplate(di,
		chargedBoltScale, collision, chargedBoltVelocity, effect, 1)

	return chargedBoltProjectile
}
