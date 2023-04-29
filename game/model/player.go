package model

import (
	"image/color"
	"lintech/rego/game/loader"
	"lintech/rego/iregoter"
	"lintech/rego/regoter"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
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
	*Entity
	CameraZ    float64
	Moved      bool
	Weapon     *Weapon
	WeaponSet  []*Weapon
	LastWeapon *Weapon
}

type playerSheet struct {
	x, y, angle, pitch float64
}

func NewPlayer(coreMsgbox chan<- iregoter.IRegoterEvent) *regoter.Regoter[*Player] {
	// init player model
	angleDegrees := 60.0

	s := playerSheet{8.5, 3.5, geom.Radians(angleDegrees), 0}
	p := &Player{
		Entity: &Entity{
			Position:  &geom.Vector2{X: s.x, Y: s.y},
			PositionZ: 0,
			Angle:     s.angle,
			Pitch:     s.pitch,
			Velocity:  0,
			MapColor:  color.RGBA{255, 0, 0, 255},
		},
		CameraZ:   0.5,
		Moved:     false,
		WeaponSet: []*Weapon{},
	}

	p.CollisionRadius = loader.ClipDistance
	p.CollisionHeight = 0.5

	r := regoter.NewRegoter(coreMsgbox, p)
	return r
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

func (c *Player) Update(v iregoter.Vision, cu iregoter.ChanRegoterUpdate) {
	// draw crosshairs
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest

	changed := true
	info := iregoter.RegoterUpdatedInfo{ImgOp: op, Img: img,
		Visiable: true, Deleted: false, Changed: changed}
	cu <- info

	// Todo
	// if c.IsHitIndicatorActive() {
	// 	screen.DrawImage(g.crosshairs.HitIndicator.Texture(), op)
	// 	g.crosshairs.Update()
	// 	cu <- info
	// }
	close(cu)
}

// Move player by move speed in the forward/backward direction
func (g *Game) Move(mSpeed float64) {
	moveLine := geom.LineFromAngle(g.player.Position.X, g.player.Position.Y, g.player.Angle, mSpeed)

	newPos, _, _ := g.getValidMove(g.player.Entity, moveLine.X2, moveLine.Y2, g.player.PositionZ, true)
	if !newPos.Equals(g.player.Pos()) {
		g.player.Position = newPos
		g.player.Moved = true
	}
}

// Move player by strafe speed in the left/right direction
func (g *Game) Strafe(sSpeed float64) {
	strafeAngle := geom.HalfPi
	if sSpeed < 0 {
		strafeAngle = -strafeAngle
	}
	strafeLine := geom.LineFromAngle(g.player.Position.X, g.player.Position.Y, g.player.Angle-strafeAngle, math.Abs(sSpeed))

	newPos, _, _ := g.getValidMove(g.player.Entity, strafeLine.X2, strafeLine.Y2, g.player.PositionZ, true)
	if !newPos.Equals(g.player.Pos()) {
		g.player.Position = newPos
		g.player.Moved = true
	}
}

// Rotate player heading angle by rotation speed
func (g *Game) Rotate(rSpeed float64) {
	g.player.Angle += rSpeed

	pi2 := geom.Pi2
	if g.player.Angle >= pi2 {
		g.player.Angle = pi2 - g.player.Angle
	} else if g.player.Angle <= -pi2 {
		g.player.Angle = g.player.Angle + pi2
	}

	g.player.Moved = true
}

// Update player pitch angle by pitch speed
func (g *Game) Pitch(pSpeed float64) {
	// current raycasting method can only allow up to 22.5 degrees down, 45 degrees up
	g.player.Pitch = geom.Clamp(pSpeed+g.player.Pitch, -math.Pi/8, math.Pi/4)
	g.player.Moved = true
}

func (g *Game) Stand() {
	g.player.CameraZ = 0.5
	g.player.PositionZ = 0
	g.player.Moved = true
}

func (g *Game) IsStanding() bool {
	return g.player.PositionZ == 0 && g.player.CameraZ == 0.5
}

func (g *Game) Jump() {
	g.player.CameraZ = 0.9
	g.player.PositionZ = 0.4
	g.player.Moved = true
}

func (g *Game) Crouch() {
	g.player.CameraZ = 0.3
	g.player.PositionZ = 0
	g.player.Moved = true
}

func (g *Game) Prone() {
	g.player.CameraZ = 0.1
	g.player.PositionZ = 0
	g.player.Moved = true
}

func (g *Game) fireWeapon() {
	w := g.player.Weapon
	if w == nil {
		g.player.NextWeapon(false)
		return
	}
	if w.OnCooldown() {
		return
	}

	// set weapon firing for animation to run
	w.Fire()

	// spawning projectile at player position just slightly below player's center point of view
	pX, pY, pZ := g.player.Position.X, g.player.Position.Y, geom.Clamp(g.player.CameraZ-0.1, 0.05, 0.95)
	// pitch, angle based on raycasted point at crosshairs
	var pAngle, pPitch float64
	convergenceDistance := g.camera.GetConvergenceDistance()
	convergencePoint := g.camera.GetConvergencePoint()
	if convergenceDistance <= 0 || convergencePoint == nil {
		pAngle, pPitch = g.player.Angle, g.player.Pitch
	} else {
		convergenceLine3d := &geom3d.Line3d{
			X1: pX, Y1: pY, Z1: pZ,
			X2: convergencePoint.X, Y2: convergencePoint.Y, Z2: convergencePoint.Z,
		}
		pAngle, pPitch = convergenceLine3d.Heading(), convergenceLine3d.Pitch()
	}

	projectile := w.SpawnProjectile(pX, pY, pZ, pAngle, pPitch, g.player.Entity)
	if projectile != nil {
		g.addProjectile(projectile)
	}
}
