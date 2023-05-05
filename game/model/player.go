package model

import (
	"image/color"
	"lintech/rego/game/loader"
	"log"
	"math"

	"github.com/harbdog/raycaster-go/geom"
)

const (
	MinimumVelocity = 1e-3
	MaximumVelocity = 1e-1
)

type Player struct {
	Reactor
	rgData RegoterData
	cfg    GameCfg

	health         int
	mouse          MousePosition
	CameraZ        float64
	Moved          bool
	weapon         RcTx
	weaponTemplate *WeaponTemplate
	weaponSet      []*WeaponTemplate

	// Movement in this tick
}

func (r *Player) ProcessMessage(m ReactorEventMessage) error {
	// log.Print(fmt.Sprintf("(%v) recv %T", r.thing.GetData().Entity.RgId, e))
	switch m.event.(type) {
	case EventUpdateTick:
		r.eventHandleUpdateTick(m.sender, m.event.(EventUpdateTick))
	case EventUpdateData:
		r.eventHandleUpdateData(m.sender, m.event.(EventUpdateData))
	case EventCfgChanged:
		r.eventHandleCfgChanged(m.sender, m.event.(EventCfgChanged))
	// case EventInput:
	// 	r.eventHandleInput(m.sender, m.event.(EventInput))
	default:
		r.eventHandleUnknown(m.sender, m.event)
	}
	return nil
}
func (r *Player) eventHandleCfgChanged(sender RcTx, e EventCfgChanged) {
	r.cfg = e.Cfg
}

func (r *Player) eventHandleUnknown(sender RcTx, e IReactorEvent) error {
	log.Fatal("Unknown event:", e)
	return nil
}

func NewPlayer(coreTx RcTx) RcTx {
	entity := Entity{
		RgId:            RgIdGenerator.GenId(),
		RgType:          RegoterEnumPlayer,
		Position:        Position{X: 8.5, Y: 3.5, Z: 0},
		Scale:           1,
		Angle:           geom.Radians(60.0),
		Pitch:           0,
		Velocity:        0,
		Resistance:      0.1,
		MapColor:        color.RGBA{0, 255, 0, 255},
		CollisionRadius: loader.ClipDistance,
		CollisionHeight: 0.5,
	}

	t := &Player{
		Reactor: NewReactor(),
		rgData: RegoterData{
			Entity: entity,
		},
		health:    fullHealth,
		CameraZ:   0.5,
		Moved:     false,
		weaponSet: NewWeapons(coreTx),
	}
	// t.rgData.DrawInfo = t.Weapon.di
	// if t.rgData.DrawInfo.Img == nil {
	// 	log.Fatal("Invalid nil Img for NewPlayer()")
	// }

	go t.Reactor.Run(t)
	m := ReactorEventMessage{t.tx, EventRegisterRegoter{t.tx, t.rgData}}
	coreTx <- m
	t.SelectWeapon(coreTx, 0)
	return t.tx
}

type playerSheet struct {
	x, y  float64
	angle float64
	pitch float64
}

func (p *Player) AddWeapon(w *WeaponTemplate) {
	p.weaponSet = append(p.weaponSet, w)
}

func (p *Player) SelectWeapon(coreTx RcTx, index int) RcTx {
	// TODO: add some kind of sheath/unsheath animation
	if index < 0 || index > len(p.weaponSet) {
		log.Fatalf("weaponIndex %v is out of range (0, %v)", index, len(p.weaponSet))
	}
	newTemplate := p.weaponSet[index]
	if newTemplate == nil || newTemplate == p.weaponTemplate {
		return p.weapon
	} else {
		if p.weapon != nil {
			p.HolsterWeapon()
		}
		p.weaponTemplate = newTemplate
		p.weapon = p.weaponTemplate.Spawn(coreTx)
		return p.weapon
	}
}
func (p *Player) HolsterWeapon() {
	m := ReactorEventMessage{p.tx, EventHolsterWeapon{}}
	p.weapon <- m
}

func (p *Player) fireWeapon() {
	m := ReactorEventMessage{p.tx, EventFireWeapon{}}
	p.weapon <- m

}

// func (p *Player) NextWeapon(reverse bool) *Weapon {
// 	_, weaponIndex := p.getSelectedWeapon()
// 	if weaponIndex < 0 {
// 		// check last weapon in event of unsheathing previously sheathed weapon
// 		weaponIndex = p.getWeaponIndex(p.LastWeapon)
// 		if weaponIndex < 0 {
// 			weaponIndex = 0
// 		}
// 		return p.SelectWeapon(weaponIndex)
// 	}

// 	weaponIndex++
// 	if weaponIndex >= len(p.weaponSet) {
// 		weaponIndex = 0
// 	}
// 	return p.SelectWeapon(weaponIndex)
// }

// func (p *Player) getWeaponIndex(w *Weapon) int {
// 	if w == nil {
// 		return -1
// 	}
// 	for index, wCheck := range p.weaponSet {
// 		if wCheck == w {
// 			return index
// 		}
// 	}
// 	return -1
// }

// func (p *Player) getSelectedWeapon() (*Weapon, int) {
// 	if p.weapon == nil {
// 		return nil, -1
// 	}

// 	return p.weapon, p.getWeaponIndex(p.weapon)
// }

func isMoving(m Movement) bool {
	if math.Abs(float64(m.Acceleration)) > MinimumVelocity ||
		math.Abs(float64(m.Velocity)) > MinimumVelocity ||
		m.MoveRotate != 0 || m.PitchRotate != 0 ||
		m.VissionRotate != 0 {
		return true
	} else {
		return false
	}
}

func (p *Player) eventHandleUpdateTick(sender RcTx, e IReactorEvent) {
	// Debug
	movement, action := handlePlayerInput(p.cfg, &p.mouse)
	movement.Velocity = p.rgData.Entity.Velocity
	if !action.KeyPressed {
		movement.MoveRotate = p.rgData.Entity.LastMoveRotate
	}

	if isMoving(movement) {
		// log.Printf("VissionRotate = %.3f", movement.VissionRotate)
		// log.Printf("Moverotate = %.3f", movement.MoveRotate)
		e := EventUpdatedMove{RgId: p.rgData.Entity.RgId, Move: movement}
		m := ReactorEventMessage{p.tx, e}
		sender <- m
	}

	if action.FireWeapon {
		p.fireWeapon()
	}
	// if info, ok := p.drawWeapon(screenSize); ok {
	// 	cu <- info
	// }

	// Todo
	// if c.IsHitIndicatorActive() {
	// 	screen.DrawImage(g.crosshairs.HitIndicator.Texture(), op)
	// 	g.crosshairs.Update()
	// 	cu <- info
	// }
}

func (p *Player) eventHandleUpdateData(sender RcTx, e EventUpdateData) {
	p.rgData.Entity = e.RgEntity
}

// // Move player by move speed in the forward/backward direction
// func (p *Player) Move(mSpeed float64) {
// 	p.movement.Acceleration = mSpeed
// }

// // Move player by strafe speed in the left/right direction
// func (p *Player) Strafe(sSpeed float64) {
// 	var strafeAngle float64 = geom.HalfPi
// 	if sSpeed < 0 {
// 		strafeAngle = -strafeAngle
// 	}
// 	p.movement.RotateSpeed = strafeAngle
// 	p.movement.Acceleration = sSpeed
// }

// // Rotate player heading angle by rotation speed
// func (p *Player) Rotate(rSpeed float64) {
// 	p.movement.RotateSpeed = rSpeed
// }

// // Update player pitch angle by pitch speed
// func (p *Player) Pitch(pSpeed float64) {
// 	// current raycasting method can only allow up to 22.5 degrees down, 45 degrees up
// 	p.movement.PitchSpeed = pSpeed
// }

func (p *Player) Stand() {
	p.CameraZ = 0.5
	p.rgData.Entity.Position.Z = 0
}

func (p *Player) IsStanding() bool {
	return p.rgData.Entity.Position.Z == 0 && p.CameraZ == 0.5
}

func (p *Player) Jump() {
	p.CameraZ = 0.9
	p.rgData.Entity.Position.Z = 0.4
	p.Moved = true
}

func (p *Player) Crouch() {
	p.CameraZ = 0.3
	p.rgData.Entity.Position.Z = 0
	p.Moved = true
}

func (p *Player) Prone() {
	p.CameraZ = 0.1
	p.rgData.Entity.Position.Z = 0
	p.Moved = true
}

// func (p *Player) fireWeapon() {
// 	w := p.Weapon
// 	if w == nil {
// 		p.NextWeapon(false)
// 		return
// 	}
// 	if w.OnCooldown() {
// 		return
// 	}

// 	// set weapon firing for animation to run
// 	w.Fire()

// 	// spawning projectile at player position just slightly below player's center point of view
// 	pX, pY, pZ := p.Entity.Position.X, p.Entity.Position.Y, geom.Clamp(p.CameraZ-0.1, 0.05, 0.95)
// 	// pitch, angle based on raycasted point at crosshairs
// 	var pAngle, pPitch float64
// 	// Todo
// 	convergenceDistance := p.camera.GetConvergenceDistance()
// 	convergencePoint := p.camera.GetConvergencePoint()
// 	if convergenceDistance <= 0 || convergencePoint == nil {
// 		pAngle, pPitch = p.Entity.Angle, p.Entity.Pitch
// 	} else {
// 		convergenceLine3d := &geom3d.Line3d{
// 			X1: pX, Y1: pY, Z1: pZ,
// 			X2: convergencePoint.X, Y2: convergencePoint.Y, Z2: convergencePoint.Z,
// 		}
// 		pAngle, pPitch = convergenceLine3d.Heading(), convergenceLine3d.Pitch()
// 	}

// 	projectile := w.SpawnProjectile(pX, pY, pZ, pAngle, pPitch, p.Entity)
// 	if projectile != nil {
// 		Todo
// 		g.addProjectile(projectile)
// 	}
// }

// func (p *Player) drawWeapon(sz ScreenSize) (RegoterUpdatedImg, bool) {
// 	// draw equipped weapon
// 	if p.Weapon != nil {
// 		w := p.Weapon
// 		op := &ebiten.DrawImageOptions{}
// 		op.Filter = ebiten.FilterNearest

// 		weaponScale := w.Scale() * p.cfg.RenderScale
// 		op.GeoM.Scale(weaponScale, weaponScale)
// 		op.GeoM.Translate(
// 			float64(sz.Width)/2-float64(w.W)*weaponScale/2,
// 			float64(sz.Height)-float64(w.H)*weaponScale+1,
// 		)

// 		// Todo
// 		// apply lighting setting
// 		//op.ColorScale.Scale(float32(g.maxLightRGB.R)/255, float32(g.maxLightRGB.G)/255, float32(g.maxLightRGB.B)/255, 1)

// 		img := w.Texture()
// 		changed := true
// 		info := RegoterUpdatedImg{ImgOp: op, Sprite: nil, Img: img,
// 			Visiable: true, Deleted: false, Changed: changed}
// 		return info, true
// 	} else {
// 		info := RegoterUpdatedImg{}
// 		return info, false
// 	}
// }

func (c *Player) SetConfig(cfg GameCfg) {
	c.cfg = cfg
}

func (c *Player) GetData() RegoterData {
	return c.rgData
}
