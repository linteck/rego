package model

import (
	"image/color"
	"lintech/rego/game/loader"
	"lintech/rego/iregoter"

	"github.com/harbdog/raycaster-go"
)

type Crosshairs struct {
	rgData iregoter.RegoterData
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

func (c *Crosshairs) UpdateTick(cu iregoter.RgTxMsgbox) {

}

func (c *Crosshairs) UpdateData(cu iregoter.RgTxMsgbox, rgEntity iregoter.Entity,
	rgState iregoter.RegoterState) bool {
	return true
}

func (c *Crosshairs) SetConfig(cfg iregoter.GameCfg) {
}

func (c *Crosshairs) GetData() iregoter.RegoterData {
	return c.rgData
}
