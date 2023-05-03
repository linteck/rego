package model

import (
	"image/color"
	"lintech/rego/game/loader"
	"lintech/rego/iregoter"

	"github.com/harbdog/raycaster-go"
)

type Effect struct {
	rgData    iregoter.RegoterData
	LoopCount int
}

func NewEffect(di iregoter.DrawInfo, scale float64, loopCount int) *Effect {
	//loadCrosshairsResource()
	entity := iregoter.Entity{
		RgId:            RgIdGenerator.GenId(),
		RgType:          iregoter.RegoterEnumSprite,
		Scale:           scale,
		Velocity:        0,
		MapColor:        color.RGBA{0, 0, 0, 0},
		Anchor:          raycaster.AnchorCenter,
		CollisionRadius: 0,
		CollisionHeight: 0,
	}
	t := &Effect{
		rgData: iregoter.RegoterData{
			Entity:   entity,
			DrawInfo: di,
		},
	}

	return t
}

func (ef *Effect) UpdateTick(cu iregoter.RgTxMsgbox) {
}

func (ef *Effect) UpdateData(cu iregoter.RgTxMsgbox, rgEntity iregoter.Entity,
	rgState iregoter.RegoterState) bool {

	if rgState.AnimationLoopCnt >= ef.LoopCount {
		e := iregoter.RegoterEventRegoterUnregister{RgId: ef.rgData.Entity.RgId}
		cu <- e
		return false
	} else {
		return true
	}
}

func (ef Effect) Spawn(coreMsgbox chan<- iregoter.IRegoterEvent,
	p iregoter.RegoterData) *Regoter[*Effect] {
	n := ef
	n.rgData.Entity.ParentId = p.Entity.RgId
	n.rgData.Entity.Position = p.Entity.Position
	r := NewRegoter(coreMsgbox, &n)
	return r
}

func (c *Effect) SetConfig(cfg iregoter.GameCfg) {
}

func (c *Effect) GetData() iregoter.RegoterData {
	return c.rgData
}

func NewRedExplosionEffect() *Effect {
	di := iregoter.DrawInfo{
		Img:           loader.GetSpriteFromFile("red_explosion_sheet.png"),
		AnimationRate: 1,
		Columns:       8,
		Rows:          3,
	}
	redExplosionEffect := NewEffect(di, 0.20, 2)
	return redExplosionEffect
}

func NewBlueExplosionEffect() *Effect {
	di := iregoter.DrawInfo{
		Img:           loader.GetSpriteFromFile("blue_explosion_sheet.png"),
		AnimationRate: 3,
		Columns:       5,
		Rows:          3,
	}
	blueExplosionEffect := NewEffect(di, 0.75, 1)
	return blueExplosionEffect
}
