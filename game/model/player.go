package model

import (
	"image/color"
	"lintech/rego/game/loader"
	"lintech/rego/iregoter"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go/geom"
)

const (
	MinimumVelocity = 1e-3
	MaximumVelocity = 1e-1
)

type resourcePlayer struct {
	texture *ebiten.Image
}

var playerResource *resourcePlayer

func loadResource() *resourcePlayer {
	if playerResource == nil {
		texture := loader.GetSpriteFromFile("crosshairs_sheet.png")
		playerResource = &resourcePlayer{texture: texture}
	}
	return playerResource
}

type Player struct {
	rgData iregoter.RegoterData
	cfg    iregoter.GameCfg

	health     int
	mouse      iregoter.MousePosition
	CameraZ    float64
	Moved      bool
	Weapon     *Weapon
	WeaponSet  []*Weapon
	LastWeapon *Weapon

	// Movement in this tick
	movement iregoter.RegoterMove
}

func NewPlayer(coreMsgbox chan<- iregoter.IRegoterEvent) *Regoter[*Player] {
	loadCrosshairsResource()
	entity := iregoter.Entity{
		RgId:            RgIdGenerator.GenId(),
		RgType:          iregoter.RegoterEnumPlayer,
		Position:        iregoter.Position{X: 8.5, Y: 3.5, Z: 0},
		Scale:           1,
		Angle:           iregoter.RotateAngle(geom.Radians(60.0)),
		Pitch:           0,
		Velocity:        0,
		MapColor:        color.RGBA{255, 0, 0, 255},
		CollisionRadius: loader.ClipDistance,
		CollisionHeight: 0.5,
		Collidable:      true}

	// di := iregoter.DrawInfo{
	// 	ImgLayer:    iregoter.ImgLayerSprite,
	// 	Img:         crosshairsResource.texture,
	// 	Columns:     8,
	// 	Rows:        8,
	// 	SpriteIndex: 55,
	// 	HitIndex:    57,
	// }
	t := &Player{
		rgData: iregoter.RegoterData{
			Entity: entity,
		},
		health:    fullHealth,
		CameraZ:   0.5,
		Moved:     false,
		WeaponSet: NewWeapons(),
	}
	t.SelectWeapon(0)
	t.rgData.DrawInfo = t.Weapon.di
	if t.rgData.DrawInfo.Img == nil {
		logger.Fatal("Invalid nil Img for NewPlayer()")
	}

	r := NewRegoter(coreMsgbox, t)
	return r
}

type playerSheet struct {
	x, y  float64
	angle iregoter.RotateAngle
	pitch iregoter.PitchAngle
}

func (p *Player) AddWeapon(w *Weapon) {
	p.WeaponSet = append(p.WeaponSet, w)
}

func (p *Player) SelectWeapon(weaponIndex int) *Weapon {
	// TODO: add some kind of sheath/unsheath animation
	if weaponIndex < 0 {
		// put away weapon
		if p.Weapon != nil {
			// store as last weapon
			p.LastWeapon = p.Weapon
		}
		p.Weapon = nil
		return nil
	}
	newWeapon := p.Weapon
	if weaponIndex < len(p.WeaponSet) {
		newWeapon = p.WeaponSet[weaponIndex]
	}
	if newWeapon != p.Weapon {
		// store as last weapon
		p.LastWeapon = p.Weapon
		p.Weapon = newWeapon
	}
	return p.Weapon
}

func (p *Player) NextWeapon(reverse bool) *Weapon {
	_, weaponIndex := p.getSelectedWeapon()
	if weaponIndex < 0 {
		// check last weapon in event of unsheathing previously sheathed weapon
		weaponIndex = p.getWeaponIndex(p.LastWeapon)
		if weaponIndex < 0 {
			weaponIndex = 0
		}
		return p.SelectWeapon(weaponIndex)
	}

	weaponIndex++
	if weaponIndex >= len(p.WeaponSet) {
		weaponIndex = 0
	}
	return p.SelectWeapon(weaponIndex)
}

func (p *Player) getWeaponIndex(w *Weapon) int {
	if w == nil {
		return -1
	}
	for index, wCheck := range p.WeaponSet {
		if wCheck == w {
			return index
		}
	}
	return -1
}

func (p *Player) getSelectedWeapon() (*Weapon, int) {
	if p.Weapon == nil {
		return nil, -1
	}

	return p.Weapon, p.getWeaponIndex(p.Weapon)
}

func (p *Player) Update(cu iregoter.RgTxMsgbox, rgEntity iregoter.Entity,
	playentity iregoter.Entity, rgState iregoter.RegoterState) {

	// Debug
	p.rgData.Entity = rgEntity
	p.movement = iregoter.RegoterMove{}
	p.handleInput(false, &p.mouse)
	// draw crosshairs
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest

	// Slow down Velocity
	if math.Abs(float64(p.movement.MoveSpeed)) < MinimumVelocity {
		if math.Abs(float64(p.rgData.Entity.Velocity)) > MinimumVelocity {
			p.movement.MoveSpeed = iregoter.Distance(-p.rgData.Entity.Velocity * 0.1)
		}
	}

	if rgEntity.Position.X != p.rgData.Entity.Position.X ||
		rgEntity.Position.Y != p.rgData.Entity.Position.Y ||
		rgEntity.Pitch != p.rgData.Entity.Pitch ||
		rgEntity.Angle != p.rgData.Entity.Angle {

		p.Moved = true
	} else {
		p.Moved = false
	}

	// de := iregoter.EventDebugPrint{DebugString: fmt.Sprintf("%+v", p.movement)}
	// cu <- de

	e := iregoter.RegoterEventUpdatedMove{RgId: p.rgData.Entity.RgId, Move: p.movement}
	cu <- e
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

// Move player by move speed in the forward/backward direction
func (p *Player) Move(mSpeed iregoter.Distance) {
	p.movement.MoveSpeed = mSpeed
}

// Move player by strafe speed in the left/right direction
func (p *Player) Strafe(sSpeed iregoter.Distance) {
	var strafeAngle iregoter.RotateAngle = geom.Pi / 8.0
	if sSpeed < 0 {
		strafeAngle = -strafeAngle
	}
	p.movement.RotateSpeed = strafeAngle
	p.movement.MoveSpeed = sSpeed
}

// Rotate player heading angle by rotation speed
func (p *Player) Rotate(rSpeed iregoter.RotateAngle) {
	p.movement.RotateSpeed = rSpeed
}

// Update player pitch angle by pitch speed
func (p *Player) Pitch(pSpeed iregoter.PitchAngle) {
	// current raycasting method can only allow up to 22.5 degrees down, 45 degrees up
	p.movement.PitchSpeed = pSpeed
}

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

// Todo
func (p *Player) fireWeapon() {}

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

// func (p *Player) drawWeapon(sz iregoter.ScreenSize) (iregoter.RegoterUpdatedImg, bool) {
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
// 		info := iregoter.RegoterUpdatedImg{ImgOp: op, Sprite: nil, Img: img,
// 			Visiable: true, Deleted: false, Changed: changed}
// 		return info, true
// 	} else {
// 		info := iregoter.RegoterUpdatedImg{}
// 		return info, false
// 	}
// }

func (c *Player) SetConfig(cfg iregoter.GameCfg) {
	c.cfg = cfg
}

func (c *Player) GetData() iregoter.RegoterData {
	return c.rgData
}
