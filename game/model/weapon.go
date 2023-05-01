package model

import (
	"image/color"
	"lintech/rego/game/loader"
	"lintech/rego/iregoter"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go"
)

type Weapon struct {
	di                 iregoter.DrawInfo
	entity             iregoter.Entity
	firing             bool
	cooldown           int
	rateOfFire         float64
	projectileVelocity float64
	projectile         Projectile
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
	// preload projectile sprites
	chargedBoltImg := loader.GetSpriteFromFile("charged_bolt_sheet.png")
	chargedBoltWidth := chargedBoltImg.Bounds().Dx()
	chargedBoltCols, chargedBoltRows := 6, 1
	chargedBoltScale := 0.3
	// in pixels, radius to use for collision testing
	chargedBoltPxRadius := 50.0
	chargedBoltCollisionRadius := (chargedBoltScale * chargedBoltPxRadius) / (float64(chargedBoltWidth) / float64(chargedBoltCols))
	chargedBoltCollisionHeight := 2 * chargedBoltCollisionRadius
	chargedBoltProjectile := NewAnimatedProjectile(
		chargedBoltScale, 1, chargedBoltImg, blueish,
		chargedBoltCols, chargedBoltRows, raycaster.AnchorCenter, chargedBoltCollisionRadius, chargedBoltCollisionHeight)

	// preload effect sprites
	blueExplosionImg := loader.GetSpriteFromFile("blue_explosion_sheet.png")
	blueExplosionEffect := NewAnimatedEffect(
		0, 0, 0.75, 3, blueExplosionImg, 5, 3, raycaster.AnchorCenter, 1,
	)
	chargedBoltProjectile.ImpactEffect = *blueExplosionEffect

	// create weapons
	chargedBoltRoF := 2.5      // Rate of Fire (as RoF/second)
	chargedBoltVelocity := 6.0 // Velocity (as distance travelled/second)
	weaponImg := loader.GetSpriteFromFile("hand_spell.png")
	chargedBoltWeapon := NewWeapon(1.0, 7, weaponImg, 3, 1, *chargedBoltProjectile,
		chargedBoltVelocity, chargedBoltRoF)
	return chargedBoltWeapon
}

func NewWeaponRedBolt() *Weapon {
	redBoltImg := loader.GetSpriteFromFile("red_bolt.png")
	redBoltWidth := redBoltImg.Bounds().Dx()
	redBoltScale := 0.25
	// in pixels, radius to use for collision testing
	redBoltPxRadius := 4.0
	redBoltCollisionRadius := (redBoltScale * redBoltPxRadius) / float64(redBoltWidth)
	redBoltCollisionHeight := 2 * redBoltCollisionRadius
	redBoltProjectile := NewProjectile(
		redBoltScale, redBoltImg, reddish, 1, 1,
		raycaster.AnchorCenter, redBoltCollisionRadius, redBoltCollisionHeight,
	)

	redExplosionImg := loader.GetSpriteFromFile("red_explosion_sheet.png")
	redExplosionEffect := NewAnimatedEffect(
		0, 0, 0.20, 1, redExplosionImg, 8, 3, raycaster.AnchorCenter, 1,
	)
	redBoltProjectile.ImpactEffect = *redExplosionEffect

	staffBoltRoF := 6.0
	staffBoltVelocity := 24.0
	weaponImg := loader.GetSpriteFromFile("hand_staff.png")
	staffBoltWeapon := NewWeapon(1.0, 7, weaponImg, 3, 1, *redBoltProjectile, staffBoltVelocity, staffBoltRoF)
	return staffBoltWeapon
}

func NewWeapons() []*Weapon {
	weapons := []*Weapon{
		NewWeaponChargedBolt(), NewWeaponRedBolt()}

	return weapons
}

func NewWeapon(scale float64,
	animationRate int, img *ebiten.Image, columns, rows int, projectile Projectile, projectileVelocity, rateOfFire float64,
) *Weapon {
	entity := iregoter.Entity{
		Scale: scale,
	}
	di := iregoter.DrawInfo{
		ImgLayer:      iregoter.ImgLayerSprite,
		Img:           img,
		Columns:       columns,
		Rows:          rows,
		AnimationRate: animationRate}
	w := Weapon{
		di:                 di,
		entity:             entity,
		firing:             false,
		cooldown:           0,
		rateOfFire:         rateOfFire,
		projectileVelocity: projectileVelocity,
		projectile:         projectile,
	}
	if w.di.Img == nil {
		logger.Fatal("Invalid nil Img for NewWeapon()")
	}
	return &w
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
