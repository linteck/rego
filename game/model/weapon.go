package model

import (
	"image/color"
	"lintech/rego/game/loader"
	"lintech/rego/iregoter"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go"
)

type Weapon struct {
	rgData     iregoter.RegoterData
	firing     bool
	cooldown   int
	rateOfFire float64
	projectile *Projectile
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

func NewWeaponChargedBolt() *Weapon {
	effect := NewBlueExplosionEffect()
	projectile := NewProjectileChargedBolt(effect)

	RoF := 2.0
	scale := 1.0
	di := iregoter.DrawInfo{
		Img:           loader.GetSpriteFromFile("hand_spell.png"),
		Columns:       3,
		Rows:          1,
		AnimationRate: 7,
	}
	weapon := NewWeapon(di, scale, projectile, RoF)
	return weapon
}

func NewWeaponRedBolt() *Weapon {
	effect := NewRedExplosionEffect()
	projectile := NewProjectileChargedBolt(effect)

	RoF := 6.0
	scale := 1.0
	di := iregoter.DrawInfo{
		Img:           loader.GetSpriteFromFile("hand_staff.png"),
		Columns:       3,
		Rows:          1,
		AnimationRate: 7,
	}
	weapon := NewWeapon(di, scale, projectile, RoF)
	return weapon
}

func NewWeapons() []*Weapon {
	weapons := []*Weapon{
		NewWeaponChargedBolt(), NewWeaponRedBolt()}

	return weapons
}

func NewWeapon(di iregoter.DrawInfo, scale float64,
	projectile *Projectile, rateOfFire float64,
) *Weapon {
	entity := iregoter.Entity{
		RgId:            RgIdGenerator.GenId(),
		RgType:          iregoter.RegoterEnumWeapon,
		Scale:           scale,
		Position:        iregoter.Position{X: 1, Y: 1, Z: 0},
		MapColor:        color.RGBA{0, 0, 0, 0},
		Anchor:          raycaster.AnchorCenter,
		CollisionRadius: 0,
		CollisionHeight: 0,
	}

	w := Weapon{
		rgData: iregoter.RegoterData{
			Entity:   entity,
			DrawInfo: di,
		},
		firing:     false,
		cooldown:   0,
		rateOfFire: rateOfFire,
		projectile: projectile,
	}

	return &w
}

func (w Weapon) Use(coreMsgbox chan<- iregoter.IRegoterEvent) *Regoter[*Weapon] {
	r := NewRegoter(coreMsgbox, &w)
	return r
}

func (w Weapon) Holster(coreMsgbox chan<- iregoter.IRegoterEvent) {
	//r := NewRegoter(coreMsgbox, &w)
}

func (c *Weapon) UpdateTick(cu iregoter.RgTxMsgbox) {

}

func (c *Weapon) UpdateData(cu iregoter.RgTxMsgbox, rgEntity iregoter.Entity,
	rgState iregoter.RegoterState) bool {
	return true
}

func (c *Weapon) SetConfig(cfg iregoter.GameCfg) {
}

func (c *Weapon) GetData() iregoter.RegoterData {
	return c.rgData
}

func (w *Weapon) Fire() bool {
	if w.cooldown <= 0 {
		// TODO: handle rate of fire greater than 60 per second?
		w.cooldown = int(1 / w.rateOfFire * float64(ebiten.TPS()))

		if !w.firing {
			w.firing = true
			//Todo
			// w.Sprite.ResetAnimation()
		}

		return true
	}
	return false
}

func (w *Weapon) OnCooldown() bool {
	return w.cooldown > 0
}

func (w *Weapon) ResetCooldown() {
	w.cooldown = 0
}

func (w *Weapon) Update() {
	if w.cooldown > 0 {
		w.cooldown -= 1
	}
	// if w.firing && w.Sprite.LoopCounter() < 1 {
	// 	w.Sprite.Update(nil)
	// } else {
	// 	w.firing = false
	// 	w.Sprite.ResetAnimation()
	// }
}
