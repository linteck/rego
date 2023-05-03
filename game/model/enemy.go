package model

import (
	"lintech/rego/iregoter"
	"math/rand"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
)

const fullHealth = 100

type Enemy struct {
	rgData       iregoter.RegoterData
	health       int
	hasCollision bool
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
		MapColor:        yellow,
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
	// Debug
	movement := iregoter.RegoterMove{
		Velocity: c.rgData.Entity.Velocity,
	}
	if c.hasCollision {
		movement.VissionRotate = rand.Float64() * geom.Pi2
		// log.Printf("%+v", movement)
	}
	if isMoving(movement) {
		e := iregoter.RegoterEventUpdatedMove{RgId: c.rgData.Entity.RgId, Move: movement}
		cu <- e
	}
}

func (c *Enemy) UpdateData(cu iregoter.RgTxMsgbox, rgEntity iregoter.Entity,
	rgState iregoter.RegoterState) {

	c.rgData.Entity = rgEntity
	c.hasCollision = rgState.HasCollision

	// log.Printf("Update Data %+v", c.rgData.Entity)
	//log.Printf("enemy: %+v", rgEntity)
	c.health -= rgState.HitHarm
	if c.health <= 0 {
		// Send Unregister to show 'Die'
		cu <- iregoter.RegoterEventRegoterUnregister{RgId: c.rgData.Entity.RgId}
	}

}

func (c *Enemy) SetConfig(cfg iregoter.GameCfg) {
}

func (c *Enemy) GetData() iregoter.RegoterData {
	return c.rgData
}
