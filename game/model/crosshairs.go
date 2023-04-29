package model

import (
	"image/color"
	"lintech/rego/game/loader"
	"lintech/rego/iregoter"
	"lintech/rego/regoter"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go"
	"github.com/jinzhu/copier"
)

type resourceCrosshairs struct {
	texture *ebiten.Image
}

type Crosshairs struct {
	*Sprite
	hitTimer     int
	HitIndicator *Crosshairs
}

type SpriteSheet struct {
	x, y, scale                             float64
	img                                     *ebiten.Image
	columns, rows, crosshairIndex, hitIndex int
}

var crosshairsResource *resourceCrosshairs

func resource() *resourceCrosshairs {
	if crosshairsResource == nil {
		crosshairsResource.texture = loader.GetSpriteFromFile("crosshairs_sheet.png")
	}
	return crosshairsResource
}

//g.tex.textures[16] = getSpriteFromFile("crosshairs_sheet.png")

func NewCrosshairs(coreMsgbox chan<- iregoter.IRegoterEvent) *regoter.Regoter[*Crosshairs] {
	s := SpriteSheet{1, 1, 2.0, crosshairsResource.texture, 8, 8, 55, 57}

	mapColor := color.RGBA{0, 0, 0, 0}
	t := &Crosshairs{
		Sprite: NewSpriteFromSheet(s.x, s.y, s.scale, s.img, mapColor, s.columns, s.rows, s.crosshairIndex, raycaster.AnchorCenter, 0, 0),
	}

	hitIndicator := &Crosshairs{}
	copier.Copy(hitIndicator, t)
	hitIndicator.Sprite.SetAnimationFrame(s.hitIndex)
	t.HitIndicator = hitIndicator

	r := regoter.NewRegoter(coreMsgbox, t)
	return r
}

func (c *Crosshairs) ActivateHitIndicator(hitTime int) {
	if c.HitIndicator != nil {
		c.hitTimer = hitTime
	}
}

func (c *Crosshairs) IsHitIndicatorActive() bool {
	return c.HitIndicator != nil && c.hitTimer > 0
}

func (c *Crosshairs) Update(v iregoter.Vision, cu iregoter.ChanRegoterUpdate) {
	if c.HitIndicator != nil && c.hitTimer > 0 {
		// TODO: prefer to use timer rather than frame update counter?
		c.hitTimer -= 1
	}
	// draw crosshairs
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest

	s := c.Sprite
	crosshairScale := c.Sprite.Scale()
	op.GeoM.Scale(crosshairScale, crosshairScale)
	op.GeoM.Translate(
		float64(v.ScreenSize.Width)/2-float64(s.W)*crosshairScale/2,
		float64(v.ScreenSize.Height)/2-float64(s.H)*crosshairScale/2,
	)
	img := s.Texture()

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
