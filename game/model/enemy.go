package model

import (
	"lintech/rego/game/loader"
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
	anchor raycaster.SpriteAnchor,
) *Regoter[*Enemy] {
	//loadEnemyResource()
	entity := iregoter.Entity{
		RgId:            RgIdGenerator.GenId(),
		RgType:          iregoter.RegoterEnumSprite,
		Position:        po,
		Scale:           scale,
		MapColor:        yellow,
		Anchor:          anchor,
		CollisionRadius: cp.CollisionRadius,
		CollisionHeight: cp.CollisionHeight,
		Velocity:        velocity,
		Angle:           rand.Float64() * geom.Pi2,
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
	rgState iregoter.RegoterState) bool {

	c.rgData.Entity = rgEntity
	c.hasCollision = rgState.HasCollision

	// log.Printf("Update Data %+v", c.rgData.Entity)
	//log.Printf("enemy: %+v", rgEntity)
	// c.health -= rgState.HitHarm
	if c.health <= 0 {
		// Send Unregister to show 'Die'
		cu <- iregoter.RegoterEventRegoterUnregister{RgId: c.rgData.Entity.RgId}
		return false
	}
	return true

}

func (c *Enemy) SetConfig(cfg iregoter.GameCfg) {
}

func (c *Enemy) GetData() iregoter.RegoterData {
	return c.rgData
}

var cnt = 1

func NewSorcerer(txToCore iregoter.RgTxMsgbox) {
	sorcImg := loader.GetSpriteFromFile("sorcerer_sheet.png")
	sorcWidth, sorcHeight := sorcImg.Bounds().Dx(), sorcImg.Bounds().Dy()
	sorcCols, sorcRows := 10, 1
	sorcScale := 1.0
	sorcVelocity := 0.02
	// in pixels, radius and height to use for collision testing
	sorcPxRadius, sorcPxHeight := 40.0, 120.0
	collisionRadius := (sorcScale * sorcPxRadius) / (float64(sorcWidth) / float64(sorcCols))
	collisionHeight := (sorcScale * sorcPxHeight) / (float64(sorcHeight) / float64(sorcRows))
	cnt += 1
	y := float64(2+cnt/100) * collisionRadius * 4
	x := float64(2+cnt%100) * collisionRadius * 4

	NewEnemy(txToCore,
		iregoter.Position{X: x, Y: y, Z: 0},
		iregoter.DrawInfo{
			Img:               sorcImg,
			ImgLayer:          iregoter.ImgLayerSprite,
			Columns:           sorcCols,
			Rows:              sorcRows,
			AnimationRate:     5,
			AnimationReversed: false,
		},
		sorcScale,
		iregoter.CollisionSpace{
			CollisionRadius: collisionRadius,
			CollisionHeight: collisionHeight,
		},
		sorcVelocity,
		raycaster.AnchorBottom,
	)
	// log.Printf("%v, %v", collisionRadius, collisionHeight)

}

func NewWalker(txToCore iregoter.RgTxMsgbox) {
	// animated walking 8-directional sprite character
	// [walkerTexFacingMap] player facing angle : texture row index
	var walkerTexFacingMap = map[float64]int{
		geom.Radians(315): 0,
		geom.Radians(270): 1,
		geom.Radians(225): 2,
		geom.Radians(180): 3,
		geom.Radians(135): 4,
		geom.Radians(90):  5,
		geom.Radians(45):  6,
		geom.Radians(0):   7,
	}
	walkerImg := loader.GetSpriteFromFile("outleader_walking_sheet.png")
	walkerWidth, walkerHeight := walkerImg.Bounds().Dx(), walkerImg.Bounds().Dy()
	walkerCols, walkerRows := 4, 8
	walkerScale := 0.75
	// in pixels, radius and height to use for collision testing
	walkerPxRadius, walkerPxHeight := 30.0, 80.0
	// convert pixel to grid using image pixel size
	walkerCollisionRadius := (walkerScale * walkerPxRadius) / (float64(walkerWidth) / float64(walkerCols))
	walkerCollisionHeight := (walkerScale * walkerPxHeight) / (float64(walkerHeight) / float64(walkerRows))
	// give sprite a sample velocity for movement
	walkerVelocity := 0.02

	cnt += 1
	y := float64(2+cnt/100)*walkerCollisionRadius*4 + walkerCollisionRadius*4
	x := float64(2+cnt%100) * walkerCollisionRadius * 4

	NewEnemy(txToCore,
		iregoter.Position{X: x, Y: y, Z: 0},
		iregoter.DrawInfo{
			Img:               walkerImg,
			ImgLayer:          iregoter.ImgLayerSprite,
			Columns:           walkerCols,
			Rows:              walkerRows,
			AnimationRate:     5,
			AnimationReversed: true,
			TexFacingMap:      &walkerTexFacingMap,
		},
		walkerScale,
		iregoter.CollisionSpace{
			CollisionRadius: walkerCollisionRadius,
			CollisionHeight: walkerCollisionHeight,
		},
		walkerVelocity,
		raycaster.AnchorBottom,
	)

	// log.Printf("%v, %v", walkerCollisionRadius, walkerCollisionHeight)
}

func NewBat(txToCore iregoter.RgTxMsgbox) {
	// animated flying 4-directional sprite creature
	// [batTexFacingMap] player facing angle : texture row index
	var batTexFacingMap = map[float64]int{
		geom.Radians(270): 1,
		geom.Radians(180): 2,
		geom.Radians(90):  3,
		geom.Radians(0):   0,
	}
	batImg := loader.GetSpriteFromFile("bat_sheet.png")
	batWidth, batHeight := batImg.Bounds().Dx(), batImg.Bounds().Dy()
	batCols, batRows := 3, 4
	batScale := 0.25
	// in pixels, radius and height to use for collision testing
	batPxRadius, batPxHeight := 14.0, 25.0
	// convert pixel to grid using image pixel size
	batCollisionRadius := (batScale * batPxRadius) / (float64(batWidth) / float64(batCols))
	batCollisionHeight := (batScale * batPxHeight) / (float64(batHeight) / float64(batRows))
	// raising Z-position of sprite model but using raycaster.AnchorTop to show below that position
	// give sprite a sample velocity for movement
	batVelocity := 0.03

	// if g.debug {
	// 	// just some debugging stuff
	// 	sorc.AddDebugLines(2, color.RGBA{0, 255, 0, 255})
	// 	walker.AddDebugLines(2, color.RGBA{0, 255, 0, 255})
	// 	batty.AddDebugLines(2, color.RGBA{0, 255, 0, 255})
	// 	chargedBoltProjectile.AddDebugLines(2, color.RGBA{0, 255, 0, 255})
	// 	redBoltProjectile.AddDebugLines(2, color.RGBA{0, 255, 0, 255})
	// }

	cnt += 1
	y := float64(2+cnt/100)*batCollisionRadius*4 + batCollisionRadius*40
	x := float64(2+cnt%100) * batCollisionRadius * 4

	NewEnemy(txToCore,
		iregoter.Position{X: x, Y: y, Z: 3},
		iregoter.DrawInfo{
			Img:               batImg,
			ImgLayer:          iregoter.ImgLayerSprite,
			Columns:           batCols,
			Rows:              batRows,
			AnimationRate:     5,
			AnimationReversed: true,
			TexFacingMap:      &batTexFacingMap,
		},
		batScale,
		iregoter.CollisionSpace{
			CollisionRadius: batCollisionRadius,
			CollisionHeight: batCollisionHeight,
		},
		batVelocity,
		raycaster.AnchorTop,
	)

	// log.Printf("%v, %v", batCollisionRadius, batCollisionHeight)
}

func NewRock(txToCore iregoter.RgTxMsgbox) {
	// rock that can be jumped over but not walked through
	rockImg := loader.GetSpriteFromFile("large_rock.png")
	rockWidth, rockHeight := rockImg.Bounds().Dx(), rockImg.Bounds().Dy()
	rockScale := 0.4
	rockPxRadius, rockPxHeight := 24.0, 35.0
	rockCollisionRadius := (rockScale * rockPxRadius) / float64(rockWidth)
	rockCollisionHeight := (rockScale * rockPxHeight) / float64(rockHeight)

	rockCols, rockRows := 1, 1
	rockVelocity := 0.0

	x := 8.0
	y := 5.5

	NewEnemy(txToCore,
		iregoter.Position{X: x, Y: y, Z: 0},
		iregoter.DrawInfo{
			Img:      rockImg,
			ImgLayer: iregoter.ImgLayerSprite,
			Columns:  rockCols,
			Rows:     rockRows,
		},
		rockScale,
		iregoter.CollisionSpace{
			CollisionRadius: rockCollisionRadius,
			CollisionHeight: rockCollisionHeight,
		},
		rockVelocity,
		raycaster.AnchorBottom,
	)
	// log.Printf("%v, %v", collisionRadius, collisionHeight)

}
