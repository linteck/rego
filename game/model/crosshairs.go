package model

import (
	"image/color"
	"lintech/rego/game/loader"
	"lintech/rego/iregoter"

	"github.com/harbdog/raycaster-go"
)

type Crosshairs struct {
	rgData iregoter.RegoterData
	health int
}

func NewCrosshairs(coreMsgbox chan<- iregoter.IRegoterEvent) *Regoter[*Crosshairs] {
	//loadCrosshairsResource()
	entity := iregoter.Entity{
		RgId:            RgIdGenerator.GenId(),
		RgType:          iregoter.RegoterEnumCrosshair,
		Position:        iregoter.Position{X: 5, Y: 5, Z: 0},
		Scale:           2,
		MapColor:        color.RGBA{255, 0, 0, 255},
		Anchor:          raycaster.AnchorCenter,
		CollisionRadius: 0,
		CollisionHeight: 0,
	}
	di := iregoter.DrawInfo{
		ImgLayer:    iregoter.ImgLayerSprite,
		Img:         loader.GetSpriteFromFile("crosshairs_sheet.png"),
		Columns:     8,
		Rows:        8,
		SpriteIndex: 55,
		HitIndex:    57,
	}
	t := &Crosshairs{
		rgData: iregoter.RegoterData{
			Entity:   entity,
			DrawInfo: di,
		},
	}

	r := NewRegoter(coreMsgbox, t)
	return r
}

// func (c *Crosshairs) ActivateHitIndicator(hitTime int) {
// 	if c.HitIndicator != nil {
// 		c.hitTimer = hitTime
// 	}
// }

// func (c *Crosshairs) IsHitIndicatorActive() bool {
// 	return c.HitIndicator != nil && c.hitTimer > 0
// }

func (c *Crosshairs) Update(cu iregoter.RgTxMsgbox, rgEntity iregoter.Entity,
	playEntiry iregoter.Entity, rgState iregoter.RegoterState) {

	c.rgData.Entity = rgEntity

	c.health -= rgState.HitHarm
	if c.health <= 0 {
		// Send Unregister to show 'Die'
		cu <- iregoter.RegoterEventRegoterUnregister{RgId: c.rgData.Entity.RgId}
	}
	// draw crosshairs
	// op := &ebiten.DrawImageOptions{}
	// op.Filter = ebiten.FilterNearest

	// s := c.Sprite
	// crosshairScale := c.Sprite.Scale()
	// op.GeoM.Scale(crosshairScale, crosshairScale)
	// op.GeoM.Translate(
	// 	float64(screenSize.Width)/2-float64(s.W)*crosshairScale/2,
	// 	float64(screenSize.Height)/2-float64(s.H)*crosshairScale/2,
	// )

	// changed := true
	// info := iregoter.RegoterUpdatedImg{ImgOp: op, Sprite: s,
	// 	Visiable: true, Deleted: false, Changed: changed}
	// e := iregoter.RegoterEventUpdatedImg{RgId: 0, Info: info}
	// cu <- e
	// Todo
	// if c.IsHitIndicatorActive() {
	// 	screen.DrawImage(g.crosshairs.HitIndicator.Texture(), op)
	// 	g.crosshairs.Update()
	// 	cu <- info
	// }
}

func (c *Crosshairs) SetConfig(cfg iregoter.GameCfg) {
}

func (c *Crosshairs) GetData() iregoter.RegoterData {
	return c.rgData
}
