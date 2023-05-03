package model

import (
	"image/color"
	"lintech/rego/iregoter"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
)

const fullHealth = 100

type Enemy struct {
	rgData iregoter.RegoterData
	health int
}

func NewEnemy(coreMsgbox chan<- iregoter.IRegoterEvent,
	po iregoter.Position, di iregoter.DrawInfo, scale float64,
	cp iregoter.CollisionSpace, velocity float64,
) *Regoter[*Enemy] {
	//loadEnemyResource()
	entity := iregoter.Entity{
		RgId:            RgIdGenerator.GenId(),
		RgType:          iregoter.RegoterEnumSprite,
		Position:        po,
		Scale:           scale,
		MapColor:        color.RGBA{255, 0, 0, 255},
		Anchor:          raycaster.AnchorBottom,
		CollisionRadius: cp.CollisionRadius,
		CollisionHeight: cp.CollisionHeight,
		Velocity:        velocity,
		Angle:           0.25 * geom.Pi,
	}
	t := &Enemy{
		rgData: iregoter.RegoterData{
			Entity:   entity,
			DrawInfo: di,
		},
		health: fullHealth,
	}

	r := NewRegoter(coreMsgbox, t)
	return r
}

// func (c *Enemy) ActivateHitIndicator(hitTime int) {
// 	if c.HitIndicator != nil {
// 		c.hitTimer = hitTime
// 	}
// }

//	func (c *Enemy) IsHitIndicatorActive() bool {
//		return c.HitIndicator != nil && c.hitTimer > 0
//	}
func (c *Enemy) UpdateTick(cu iregoter.RgTxMsgbox) {

}

func (c *Enemy) UpdateData(cu iregoter.RgTxMsgbox, rgEntity iregoter.Entity,
	rgState iregoter.RegoterState) {

	c.rgData.Entity = rgEntity

	//log.Printf("enemy: %+v", rgEntity)
	c.health -= rgState.HitHarm
	if c.health <= 0 {
		// Send Unregister to show 'Die'
		cu <- iregoter.RegoterEventRegoterUnregister{RgId: c.rgData.Entity.RgId}
	}

	e := iregoter.RegoterEventUpdatedMove{RgId: c.rgData.Entity.RgId,
		Move: iregoter.RegoterMove{MoveRotate: 0, Acceleration: 0, PitchRotate: 0}}
	cu <- e
}

func (c *Enemy) SetConfig(cfg iregoter.GameCfg) {
}

func (c *Enemy) GetData() iregoter.RegoterData {
	return c.rgData
}
