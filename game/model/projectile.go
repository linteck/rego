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
	lifespan    int
	harm        int
	effect      *EffectTemplate
	audioPlayer *RegoAudioPlayer
}

type Projectile struct {
	Reactor
	ProjectileTemplate
	unregistered bool
}

func (r *Projectile) ProcessMessage(m ReactorEventMessage) error {
	switch m.event.(type) {
	case EventUnregisterConfirmed:
		r.eventHandleUnregisterConfirmed(m.sender, m.event.(EventUnregisterConfirmed))
	default:
		if !r.unregistered {
			// log.Print(fmt.Sprintf("(%v) recv %T", r.thing.GetData().Entity.RgId, e))
			switch m.event.(type) {
			case EventUpdateTick:
				r.eventHandleUpdateTick(m.sender, m.event.(EventUpdateTick))
			case EventCollision:
				r.eventHandleCollision(m.sender, m.event.(EventCollision))
			case EventHealthChange:
				r.eventHandleHealthChange(m.sender, m.event.(EventHealthChange))
			case EventUnregisterConfirmed:
				r.eventHandleUnregisterConfirmed(m.sender, m.event.(EventUnregisterConfirmed))
			case EventCfgChanged:
				r.eventHandleCfgChanged(m.sender, m.event.(EventCfgChanged))
			default:
				r.eventHandleUnknown(m.sender, m.event)
			}
		}
	}
	return nil
}

func (r *Projectile) eventHandleUnknown(sender RcTx, e IReactorEvent) error {
	log.Fatalf("Unknown event: %T", e)
	return nil
}

func (r *Projectile) eventHandleHealthChange(sender RcTx, e EventHealthChange) {
}

func (c *Projectile) eventHandleCollision(sender RcTx, e EventCollision) {
	if e.collistion.peer == NULL_ID {
		log.Fatalf("Info: Try to find NULL_ID(%v) in core", NULL_ID)
	}
	if e.collistion.peer != WALL_ID {
		d := ReactorEventMessage{c.tx, EventDamagePeer{peer: e.collistion.peer, damage: c.harm}}
		sender <- d
	}

	m := ReactorEventMessage{c.tx, EventUnregisterRegoter{RgId: c.rgData.Entity.RgId}}
	sender <- m
	c.unregistered = true
	c.effect.Spawn(sender, e.collistion.position)
}

func (c *Projectile) eventHandleUpdateTick(sender RcTx, e EventUpdateTick) {
	c.lifespan -= 1

	if c.lifespan < 0 {
		m := ReactorEventMessage{c.tx, EventUnregisterRegoter{RgId: c.rgData.Entity.RgId}}
		sender <- m
		c.effect.Spawn(sender, e.RgEntity.Position)
	} else {
		m := ReactorEventMessage{c.tx, EventMovement{RgId: c.rgData.Entity.RgId,
			Move: Movement{Velocity: c.rgData.Entity.Velocity}}}
		sender <- m
	}

}

func (c *Projectile) eventHandleUnregisterConfirmed(sender RcTx, e EventUnregisterConfirmed) {
	c.running = false
}

func (p *ProjectileTemplate) Spawn(coreTx RcTx, pt *ProjectileTemplate,
	parentId ID, position Position, aimAngle float64, aimPitch float64) RcTx {
	p.playAudio()
	// Because weapon's location is higher than player feet.
	position.Z = 0.3
	return NewProjectile(coreTx, pt, parentId, position, aimAngle, aimPitch)
}

func (c *Projectile) eventHandleCfgChanged(sender RcTx, e EventCfgChanged) {
}

func NewProjectile(coreTx RcTx, pt *ProjectileTemplate, parentId ID, position Position,
	aimAngle float64, aimPitch float64) RcTx {

	p := &Projectile{
		Reactor:            NewReactor(),
		ProjectileTemplate: *pt,
	}
	// Don't use ID of Template
	p.rgData.Entity.RgId = <-IdGen
	p.rgData.Entity.ParentId = parentId
	p.rgData.Entity.Position = position
	p.rgData.Entity.Angle = aimAngle
	p.rgData.Entity.Pitch = aimPitch
	go p.Reactor.Run(p)
	m := ReactorEventMessage{p.tx, EventRegisterRegoter{p.tx, p.rgData}}
	coreTx <- m
	return p.tx
}

func NewProjectileTemplate(di DrawInfo,
	scale float64, collision CollisionSpace, velocity float64,
	effect *EffectTemplate, harm int, audioPlayer *RegoAudioPlayer,
) *ProjectileTemplate {
	//loadCrosshairsResource()
	entity := Entity{
		RgId:            <-IdGen,
		RgType:          RegoterEnumProjectile,
		RgName:          "Projectile",
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
		audioPlayer: audioPlayer,
		effect:      effect,
		harm:        harm,
		lifespan:    100,
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
	chargedBoltVelocity := 0.5 // Velocity (as distance travelled/second)
	audioPlayer := LoadAudioPlayer("blaster.mp3")
	chargedBoltProjectile := NewProjectileTemplate(di,
		chargedBoltScale, collision, chargedBoltVelocity, effect, 50, audioPlayer)

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
	redBoltVelocity := 0.5 // Velocity (as distance travelled/second)
	audioPlayer := LoadAudioPlayer("jab.wav")
	redBoltProjectile := NewProjectileTemplate(di,
		redBoltScale, collision, redBoltVelocity, effect, 30, audioPlayer)

	return redBoltProjectile
}

func (w *ProjectileTemplate) playAudio() {
	// Weapon postion does not update.
	// It is always with Player. So we jsut play audio with a fixed volume.
	w.audioPlayer.PlayWithVolume(0.3, true)
}
