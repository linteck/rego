package model

import (
	"fmt"
	"image/color"
	"lintech/rego/game/loader"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go"
)

type WeaponTemplate struct {
	rgData     RegoterData
	projectile *ProjectileTemplate
	cooldown   int
	rateOfFire float64
}

type Weapon struct {
	Reactor
	WeaponTemplate
	firing     bool
	fireWeapon bool
}

var (
	// colors for minimap representation
	blueish = color.RGBA{62, 62, 100, 96}
	reddish = color.RGBA{180, 62, 62, 96}
	brown   = color.RGBA{47, 40, 30, 196}
	green   = color.RGBA{27, 37, 7, 196}
	orange  = color.RGBA{69, 30, 5, 196}
	yellow  = color.RGBA{255, 200, 0, 196}
)

func (r *Weapon) Run() {
	r.running = true
	var err error
	for r.running {
		msg := <-r.rx
		err = r.process(msg)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (r *Weapon) process(m ReactorEventMessage) error {
	// logger.Print(fmt.Sprintf("(%v) recv %T", r.thing.GetData().Entity.RgId, e))
	switch m.event.(type) {
	case EventUpdateTick:
		r.eventHandleUpdateTick(m.sender, m.event.(EventUpdateTick))
	case EventUpdateData:
		r.eventHandleUpdateData(m.sender, m.event.(EventUpdateData))
	case EventCfgChanged:
		r.eventHandleCfgChanged(m.sender, m.event.(EventCfgChanged))
	case EventFireWeapon:
		r.eventHandleFireWeapon(m.sender, m.event.(EventInput))
	default:
		r.eventHandleUnknown(m.sender, m.event)
	}
	return nil
}

// Update the position and status of Regoter
// And send new Position and status to IGame
func (r *Weapon) eventHandleUpdateData(sender RcTx, e EventUpdateData) {
}

func (w *Weapon) eventHandleUpdateTick(sender RcTx, e EventUpdateTick) error {
	if w.fireWeapon {
		if w.cooldown > 0 {
			w.cooldown -= 1
		} else {
			w.cooldown = int(1 / w.rateOfFire * float64(ebiten.TPS()))
			w.projectile.Spawn(sender, w.rgData)
			w.fireWeapon = false
		}
	}
	return nil
}

func (r *Weapon) eventHandleCfgChanged(sender RcTx, e EventCfgChanged) error {
	return nil
}

func (r *Weapon) eventHandleFireWeapon(sender RcTx, e EventInput) error {
	r.fireWeapon = true
	return nil
}

func (r *Weapon) eventHandleUnknown(sender RcTx, e IReactorEvent) error {
	logger.Fatal("Unknown event:", e)
	return nil
}

func NewWeaponChargedBolt(coreTx RcTx) *WeaponTemplate {
	effect := NewBlueExplosionEffect()
	projectile := ProjectileChargedBolt(effect)

	RoF := 2.0
	scale := 1.0
	di := DrawInfo{
		Img:           loader.GetSpriteFromFile("hand_spell.png"),
		Columns:       3,
		Rows:          1,
		AnimationRate: 7,
	}
	t := NewWeaponTemplate(coreTx, di, scale, projectile, RoF)
	return t
}

func NewWeaponRedBolt(coreTx RcTx) *WeaponTemplate {
	effect := NewRedExplosionEffect()
	projectile := ProjectileChargedBolt(effect)

	RoF := 6.0
	scale := 1.0
	di := DrawInfo{
		Img:           loader.GetSpriteFromFile("hand_staff.png"),
		Columns:       3,
		Rows:          1,
		AnimationRate: 7,
	}
	t := NewWeaponTemplate(coreTx, di, scale, projectile, RoF)
	return t
}

func NewWeapons(coreTx RcTx) []*WeaponTemplate {
	weapons := []*WeaponTemplate{
		NewWeaponChargedBolt(coreTx), NewWeaponRedBolt(coreTx)}

	return weapons
}

func NewWeaponTemplate(coreTx RcTx, di DrawInfo, scale float64,
	projectile *ProjectileTemplate, rateOfFire float64,
) *WeaponTemplate {
	entity := Entity{
		RgId:            RgIdGenerator.GenId(),
		RgType:          RegoterEnumWeapon,
		Scale:           scale,
		Position:        Position{X: 1, Y: 1, Z: 0},
		MapColor:        color.RGBA{0, 0, 0, 0},
		Anchor:          raycaster.AnchorCenter,
		CollisionRadius: 0,
		CollisionHeight: 0,
	}

	w := WeaponTemplate{
		rgData: RegoterData{
			Entity:   entity,
			DrawInfo: di,
		},
		cooldown:   0,
		rateOfFire: rateOfFire,
		projectile: projectile,
	}

	return &w
}

func NewWeapon(coreTx RcTx, tp *WeaponTemplate) RcTx {
	w := Weapon{
		Reactor:        NewReactor(),
		WeaponTemplate: *tp,
	}
	go w.Run()
	m := ReactorEventMessage{w.tx, EventRegisterRegoter{w.tx, w.rgData}}
	coreTx <- m

	return w.tx
}

func (t *WeaponTemplate) Spawn(coreTx RcTx) RcTx {
	return NewWeapon(coreTx, t)
}

func (w *Weapon) PullTrigger(pulledTrigger bool) {
	w.fireWeapon = pulledTrigger || w.fireWeapon
}

// func (w *Weapon) Update() {
// 	if w.cooldown > 0 {
// 		w.cooldown -= 1
// 	}
// 	// if w.firing && w.Sprite.LoopCounter() < 1 {
// 	// 	w.Sprite.Update(nil)
// 	// } else {
// 	// 	w.firing = false
// 	// 	w.Sprite.ResetAnimation()
// 	// }
// }
