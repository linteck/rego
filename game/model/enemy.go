package model

import (
	"lintech/rego/game/loader"
	"log"
	"math/rand"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
)

const fullHealth = 100

type Enemy struct {
	Reactor
	rgData           RegoterData
	cfg              GameCfg
	unregistered     bool
	health           int
	collistionRotate float64
	harm             int
	audioPlayer      *RegoAudioPlayer
}

func (r *Enemy) ProcessMessage(m ReactorEventMessage) error {
	// log.Print(fmt.Sprintf("(%v) recv %T", r.thing.GetData().Entity.RgId, e))
	switch m.event.(type) {
	case EventUnregisterConfirmed:
		r.eventHandleUnregisterConfirmed(m.sender, m.event.(EventUnregisterConfirmed))
	default:
		if !r.unregistered {
			switch m.event.(type) {
			case EventCollision:
				r.eventHandleCollision(m.sender, m.event.(EventCollision))
			case EventHealthChange:
				r.eventHandleHealthChange(m.sender, m.event.(EventHealthChange))
			case EventUpdateTick:
				r.eventHandleUpdateTick(m.sender, m.event.(EventUpdateTick))
			case EventCfgChanged:
				r.eventHandleCfgChanged(m.sender, m.event.(EventCfgChanged))
			default:
				r.eventHandleUnknown(m.sender, m.event)
			}
		}
	}
	return nil
}

func (r *Enemy) eventHandleUnknown(sender RcTx, e IReactorEvent) error {
	log.Fatalf("Unknown event: %T", e)
	return nil
}

func (r *Enemy) eventHandleCfgChanged(sender RcTx, e EventCfgChanged) {
	r.cfg = e.Cfg
}

func (r *Enemy) eventHandleHealthChange(sender RcTx, e EventHealthChange) {
	r.health -= e.change
	if r.health < 0 {
		m := ReactorEventMessage{r.tx, EventUnregisterRegoter{RgId: r.rgData.Entity.RgId}}
		sender <- m
		r.unregistered = true
	}

}

func (c *Enemy) eventHandleCollision(sender RcTx, e EventCollision) {
	if e.collistion.peer == NULL_ID {
		log.Fatalf("Info: Try to find NULL_ID(%v) in core", NULL_ID)
	}
	if e.collistion.peer != WALL_ID && c.harm != 0 {
		m := ReactorEventMessage{c.tx, EventDamagePeer{peer: e.collistion.peer, damage: c.harm}}
		sender <- m
	}
	c.collistionRotate = rand.Float64() * geom.Pi2
}

func NewEnemy(coreTx RcTx,
	po Position, di DrawInfo, scale float64,
	cp CollisionSpace, velocity float64, harm int,
	anchor raycaster.SpriteAnchor, audioPlayer *RegoAudioPlayer,
) RcTx {
	//loadEnemyResource()
	entity := Entity{
		RgId:            <-IdGen,
		RgType:          RegoterEnumSprite,
		RgName:          "Enemy",
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
		Reactor: NewReactor(),
		rgData: RegoterData{
			Entity:   entity,
			DrawInfo: di,
		},
		health:      fullHealth,
		harm:        harm,
		audioPlayer: audioPlayer,
	}

	go t.Reactor.Run(t)
	m := ReactorEventMessage{t.tx, EventRegisterRegoter{t.tx, t.rgData}}
	coreTx <- m
	return t.tx
}

// func (c *Enemy) ActivateHitIndicator(hitTime int) {
// 	if c.HitIndicator != nil {
// 		c.hitTimer = hitTime
// 	}
// }

//	func (c *Enemy) IsHitIndicatorActive() bool {
//		return c.HitIndicator != nil && c.hitTimer > 0
//	}
func (c *Enemy) eventHandleUpdateTick(sender RcTx, e EventUpdateTick) {
	c.rgData.Entity = e.RgEntity
	movement := Movement{
		Velocity: c.rgData.Entity.Velocity,
	}
	if c.collistionRotate != 0 {
		movement.VissionRotate += c.collistionRotate
		c.collistionRotate = 0
	}
	if isMoving(movement) {
		v := EventMovement{RgId: c.rgData.Entity.RgId, Move: movement}
		m := ReactorEventMessage{c.tx, v}
		sender <- m
		c.playAudio(e)
	}
}

func (c *Enemy) eventHandleUnregisterConfirmed(sender RcTx, e EventUnregisterConfirmed) {
	c.running = false
}

func (c *Enemy) SetConfig(cfg GameCfg) {
}

func (c *Enemy) GetData() RegoterData {
	return c.rgData
}

var cnt = 1

func NewSorcerer(conrTx RcTx) {
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

	NewEnemy(conrTx,
		Position{X: x, Y: y, Z: 0},
		DrawInfo{
			Img:               sorcImg,
			ImgLayer:          ImgLayerSprite,
			Columns:           sorcCols,
			Rows:              sorcRows,
			AnimationRate:     5,
			AnimationReversed: false,
		},
		sorcScale,
		CollisionSpace{
			CollisionRadius: collisionRadius,
			CollisionHeight: collisionHeight,
		},
		sorcVelocity,
		10,
		raycaster.AnchorBottom,
		LoadAudioPlayer("swinging-whoosh.mp3"),
	)
	// log.Printf("%v, %v", collisionRadius, collisionHeight)

}

func NewWalker(coreTx RcTx) {
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

	NewEnemy(coreTx,
		Position{X: x, Y: y, Z: 0},
		DrawInfo{
			Img:               walkerImg,
			ImgLayer:          ImgLayerSprite,
			Columns:           walkerCols,
			Rows:              walkerRows,
			AnimationRate:     5,
			AnimationReversed: true,
			TexFacingMap:      &walkerTexFacingMap,
		},
		walkerScale,
		CollisionSpace{
			CollisionRadius: walkerCollisionRadius,
			CollisionHeight: walkerCollisionHeight,
		},
		walkerVelocity,
		5,
		raycaster.AnchorBottom,
		LoadAudioPlayer("werewolf.wav"),
	)

	// log.Printf("%v, %v", walkerCollisionRadius, walkerCollisionHeight)
}

func NewBat(coreTx RcTx) {
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

	NewEnemy(coreTx,
		Position{X: x, Y: y, Z: 3},
		DrawInfo{
			Img:               batImg,
			ImgLayer:          ImgLayerSprite,
			Columns:           batCols,
			Rows:              batRows,
			AnimationRate:     5,
			AnimationReversed: true,
			TexFacingMap:      &batTexFacingMap,
		},
		batScale,
		CollisionSpace{
			CollisionRadius: batCollisionRadius,
			CollisionHeight: batCollisionHeight,
		},
		batVelocity,
		3,
		raycaster.AnchorTop,
		LoadAudioPlayer("cat.wav"),
	)

	// log.Printf("%v, %v", batCollisionRadius, batCollisionHeight)
}

func NewRock(coreTx RcTx) {
	// rock that can be jumped over but not walked through
	rockImg := loader.GetSpriteFromFile("large_rock.png")
	rockWidth, rockHeight := rockImg.Bounds().Dx(), rockImg.Bounds().Dy()
	rockScale := 0.4
	rockPxRadius, rockPxHeight := 24.0, 35.0
	rockCollisionRadius := (rockScale * rockPxRadius) / float64(rockWidth)
	rockCollisionHeight := (rockScale * rockPxHeight) / float64(rockHeight)

	rockCols, rockRows := 1, 1
	rockVelocity := 0.0

	y := float64(2+cnt/100)*rockCollisionRadius*4 + rockCollisionRadius*60
	x := float64(2+cnt%100) * rockCollisionRadius * 4

	NewEnemy(coreTx,
		Position{X: x, Y: y, Z: 0},
		DrawInfo{
			Img:      rockImg,
			ImgLayer: ImgLayerSprite,
			Columns:  rockCols,
			Rows:     rockRows,
		},
		rockScale,
		CollisionSpace{
			CollisionRadius: rockCollisionRadius,
			CollisionHeight: rockCollisionHeight,
		},
		rockVelocity,
		0,
		raycaster.AnchorBottom,
		nil,
	)
	// log.Printf("%v, %v", collisionRadius, collisionHeight)

}

func (c *Enemy) playAudio(e EventUpdateTick) {
	if c.audioPlayer != nil && e.RgState.IsAnimationFirstFrame {
		c.audioPlayer.Play(e.RgEntity.Position, e.PlayerEntity.Position, c.cfg.RenderAudioDistance)
	}
}
