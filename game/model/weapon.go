package model

import (
	"image/color"
	"lintech/rego/iregoter"

	"github.com/harbdog/raycaster-go"

	"github.com/hajimehoshi/ebiten/v2"
)

type Weapon struct {
	*iregoter.Sprite
	firing             bool
	cooldown           int
	rateOfFire         float64
	projectileVelocity float64
	projectile         Projectile
}

func NewAnimatedWeapon(
	x, y, scale float64, animationRate int, img *ebiten.Image, columns, rows int, projectile Projectile, projectileVelocity, rateOfFire float64,
) *Weapon {
	mapColor := color.RGBA{0, 0, 0, 0}
	w := &Weapon{
		Sprite: iregoter.NewAnimatedSprite(x, y, scale, animationRate, img, mapColor, columns, rows, raycaster.AnchorCenter, 0, 0),
	}
	w.projectile = projectile
	w.projectileVelocity = projectileVelocity
	w.rateOfFire = rateOfFire

	return w
}

func (w *Weapon) Fire() bool {
	if w.cooldown <= 0 {
		// TODO: handle rate of fire greater than 60 per second?
		w.cooldown = int(1 / w.rateOfFire * float64(ebiten.TPS()))

		if !w.firing {
			w.firing = true
			w.Sprite.ResetAnimation()
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
	if w.firing && w.Sprite.LoopCounter() < 1 {
		w.Sprite.Update(nil)
	} else {
		w.firing = false
		w.Sprite.ResetAnimation()
	}
}
