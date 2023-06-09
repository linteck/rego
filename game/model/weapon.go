package model

import (
	"image/color"
	"lintech/rego/game/loader"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/harbdog/raycaster-go"
)

type WeaponTemplate struct {
	rgData      RegoterData
	projectile  *ProjectileTemplate
	rateOfFire  float64
	audioPlayer *audio.Player
}

type Weapon struct {
	Reactor
	WeaponTemplate
	firing     bool
	fireWeapon ICooldownFlag
	// fireWeapon bool
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

func (r *Weapon) ProcessMessage(m ReactorEventMessage) error {
	// log.Print(fmt.Sprintf("(%v) recv %T", r.thing.GetData().Entity.RgId, e))
	switch m.event.(type) {
	case EventUpdateTick:
		r.eventHandleUpdateTick(m.sender, m.event.(EventUpdateTick))
	case EventUpdateData:
		r.eventHandleUpdateData(m.sender, m.event.(EventUpdateData))
	case EventCfgChanged:
		r.eventHandleCfgChanged(m.sender, m.event.(EventCfgChanged))
	case EventFireWeapon:
		r.eventHandleFireWeapon(m.sender, m.event.(EventFireWeapon))
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
	w.fireWeapon.cooldown()
	if w.fireWeapon.get() {
		w.projectile.Spawn(sender, w.WeaponTemplate.projectile, e.PlayerEntity.RgId,
			e.PlayerEntity.Position, e.PlayerEntity.Angle, e.PlayerEntity.Pitch)
		if err := w.audioPlayer.Rewind(); err != nil {
			log.Printf("Warning: Audioplayer Rewind fail!")
		} else {
			w.audioPlayer.Play()
		}
		if !e.RgState.AnimationRunning {
			startAnimation := ReactorEventMessage{w.tx, EventMovement{
				RgId:    w.rgData.Entity.RgId,
				Command: Command{StartAnimation: true}}}
			sender <- startAnimation
		}
	} else {
		if e.RgState.AnimationRunning && e.RgState.AnimationLoopCnt >= 1 {
			stopAnimation := ReactorEventMessage{w.tx, EventMovement{
				RgId:    w.rgData.Entity.RgId,
				Command: Command{StopAnimation: true}}}
			sender <- stopAnimation
		}
	}
	return nil
}

func (r *Weapon) eventHandleCfgChanged(sender RcTx, e EventCfgChanged) {
}
func (w *Weapon) eventHandleFireWeapon(sender RcTx, e EventFireWeapon) error {
	w.fireWeapon.set()
	return nil
}

func (r *Weapon) eventHandleUnknown(sender RcTx, e IReactorEvent) error {
	log.Fatalf("Unknown event: %T", e)
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
	audioPlayer := loader.LoadAudioWav("blaster.mp3")
	t := NewWeaponTemplate(coreTx, di, scale, projectile, RoF, audioPlayer)
	return t
}

func NewWeaponRedBolt(coreTx RcTx) *WeaponTemplate {
	effect := NewRedExplosionEffect()
	projectile := ProjectileRedBolt(effect)

	RoF := 6.0
	scale := 1.0
	di := DrawInfo{
		Img:           loader.GetSpriteFromFile("hand_staff.png"),
		Columns:       3,
		Rows:          1,
		AnimationRate: 7,
	}
	audioPlayer := loader.LoadAudioWav("jab.wav")
	t := NewWeaponTemplate(coreTx, di, scale, projectile, RoF, audioPlayer)
	return t
}

func NewWeapons(coreTx RcTx) []*WeaponTemplate {
	weapons := []*WeaponTemplate{
		NewWeaponChargedBolt(coreTx), NewWeaponRedBolt(coreTx)}

	return weapons
}

func NewWeaponTemplate(coreTx RcTx, di DrawInfo, scale float64,
	projectile *ProjectileTemplate, rateOfFire float64, audioPlayer *audio.Player,
) *WeaponTemplate {
	entity := Entity{
		RgId:            <-IdGen,
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
		projectile:  projectile,
		rateOfFire:  rateOfFire,
		audioPlayer: audioPlayer,
	}

	return &w
}

func NewWeapon(coreTx RcTx, tp *WeaponTemplate) RcTx {
	cooldownInit := int(float64(ebiten.TPS())/float64(tp.rateOfFire)) + 1
	w := &Weapon{
		Reactor:        NewReactor(),
		WeaponTemplate: *tp,
		fireWeapon:     &cooldownFlag{counterInit: cooldownInit},
	}
	// Don't use ID of Template
	w.rgData.Entity.RgId = <-IdGen
	go w.Reactor.Run(w)
	m := ReactorEventMessage{w.tx, EventRegisterRegoter{w.tx, w.rgData}}
	coreTx <- m

	return w.tx
}

func (t *WeaponTemplate) Spawn(coreTx RcTx) RcTx {
	return NewWeapon(coreTx, t)
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
