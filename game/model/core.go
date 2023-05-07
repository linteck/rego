package model

import (
	"image/color"
	"lintech/rego/game/loader"
	"log"
	"math"

	"github.com/chen3feng/stl4go"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
)

type regoterInCore struct {
	tx     RcTx
	rgType RegoterEnum
	sprite *Sprite
	state  RegoterState
	entity Entity
	di     DrawInfo
}

var allRegoterEnum = [...]RegoterEnum{
	RegoterEnumSprite,
	RegoterEnumProjectile,
	RegoterEnumEffect,
	RegoterEnumCrosshair,
	RegoterEnumWeapon,
	RegoterEnumPlayer,
}

type Core struct {
	Reactor
	cfg GameCfg
	rgs [len(allRegoterEnum)]*stl4go.SkipList[ID, *regoterInCore]

	// Camera
	camera        *raycaster.Camera
	scene         *ebiten.Image
	tex           *loader.TextureHandler
	debugMessages *stl4go.DList[string]
	//--array of levels, levels refer to "floors" of the world--//
	mapObj       *loader.Map
	collisionMap []geom.Line
	mapWidth     int
	mapHeight    int
}

func (g *Core) ProcessMessage(m ReactorEventMessage) error {
	switch m.event.(type) {
	case EventDebugPrint:
		g.eventHandleEventDebugPrint(m.sender, m.event.(EventDebugPrint))

	case EventHolsterWeapon:
		g.eventHandleHolsterWeapon(m.sender, m.event.(EventHolsterWeapon))
	case EventGameTick:
		g.eventHandleGameEventTick(m.sender, m.event.(EventGameTick))

	case EventCfgChanged:
		g.eventHandleGameEventCfgChanged(m.sender, m.event.(EventCfgChanged))

	case EventDraw:
		g.eventHandleGameEventDraw(m.sender, m.event.(EventDraw))

	case EventRegisterRegoter:
		g.eventHandleRegisterRegoter(m.sender, m.event.(EventRegisterRegoter))

	case EventUnregisterRegoter:
		g.eventHandleRegoterUnregister(m.sender, m.event.(EventUnregisterRegoter))

	case EventMovement:
		g.eventHandleMovement(m.sender, m.event.(EventMovement))

	case EventDamage:
		g.eventHandleDamage(m.sender, m.event.(EventDamage))
	default:
		g.eventHandleUnknown(m.sender, m.event)
	}

	return nil
}

func (g *Core) getPlayer() *regoterInCore {
	// check sprite against player collision
	var player *regoterInCore = nil
	pl := g.rgs[RegoterEnumPlayer]
	if pl.Len() > 0 {
		player = pl.Iterate().Value()
	}
	return player
}

func (g *Core) eventHandleHolsterWeapon(sender RcTx, e EventHolsterWeapon) {
	pl := g.rgs[RegoterEnumWeapon]
	pl.Clear()
}

func (g *Core) eventHandleGameEventTick(sender RcTx, e EventGameTick) {
	player := g.getPlayer()
	for _, l := range g.rgs {
		l.ForEach(func(k ID, v *regoterInCore) {
			m := ReactorEventMessage{g.tx,
				EventUpdateTick{RgState: v.state, RgEntity: v.entity, PlayerEntity: player.entity}}
			v.tx <- m
		})
	}

	for _, l := range g.rgs {
		l.ForEach(func(k ID, v *regoterInCore) {
			if v.sprite != nil {
				if !v.state.AnimationRunning {
					v.sprite.ResetAnimation()
				} else {
					v.sprite.Update(g.camera.GetPosition())
				}
				if v.di.AnimationRate > 0 && v.sprite != nil {
					v.state.AnimationLoopCnt = v.sprite.LoopCounter()
					v.state.IsAnimationFirstFrame = v.sprite.IsLoopFirstFrame()
				}
			}
		})
	}

}

func (g *Core) eventHandleGameEventCfgChanged(sender RcTx, e EventCfgChanged) {
	if g.cfg != e.Cfg {
		g.cfg = e.Cfg
		for _, l := range g.rgs {
			l.ForEach(func(k ID, v *regoterInCore) {
				e := ReactorEventMessage{g.tx, EventCfgChanged{}}
				v.tx <- e
			})
		}
		// log.Print(fmt.Sprintf("current rg num %v", g.rgs.Len()))
		g.applyConfig()
	}
}

func (g *Core) eventHandleEventDebugPrint(sender RcTx, e EventDebugPrint) {
	g.debugMessages.PushBack(e.DebugString)
}

func createCoreSprite(rg *regoterInCore) *Sprite {
	var sprite *Sprite
	if rg.di.Img == nil {
		log.Printf("Warning!!! Register Regoter without Img. Will not create Sprite for it.")
		return nil
	}
	if rg.di.AnimationRate != 0 {
		sprite = NewAnimatedSprite(&rg.entity, rg.di.Img,
			rg.di.Columns, rg.di.Rows, rg.di.AnimationRate)
		sprite.SetAnimationReversed(rg.di.AnimationReversed)
		if rg.di.TexFacingMap != nil {
			sprite.SetTextureFacingMap(*rg.di.TexFacingMap)
		}
	} else {
		sprite = NewSpriteFromSheet(&rg.entity, rg.di.Img,
			rg.di.Columns, rg.di.Rows, rg.di.SpriteIndex)
		sprite.SetAnimationFrame(rg.di.HitIndex)
	}
	sprite.SetIllumination(rg.di.Illumination)

	return sprite

}

func (g *Core) eventHandleRegisterRegoter(sender RcTx, e EventRegisterRegoter) {
	d := e.RgData
	rg := &regoterInCore{tx: sender, rgType: d.Entity.RgType, entity: d.Entity, di: d.DrawInfo}
	if rg.di.AnimationRate != 0 && rg.di.SpriteIndex != 0 {
		log.Fatal("This Regoter can not be both Animation and Sheet")
	}
	if rg.di.Img == nil && d.Entity.RgType != RegoterEnumPlayer {
		log.Fatal("Invalid nil Img for ", d.Entity.RgType, d.Entity.RgId)
	}
	rg.state.AnimationRunning = true
	rg.sprite = createCoreSprite(rg)
	g.rgs[rg.rgType].Insert(d.Entity.RgId, rg)
}

// func (g *Core) eventHandleUpdatedImg(e RegoterEventUpdatedImg) {
// 	l := g.imgs[e.Info.ImgLayer]
// 	l.PushBack(e.Info)
// }

func (g *Core) findRegoter(id ID) (*regoterInCore, bool) {
	if id == NULL_ID {
		log.Fatalf("ID can not be NULL_ID(%v).", NULL_ID)
	}
	if id == WALL_ID {
		log.Printf("Info: Try to find WALL_ID(%v) in core", WALL_ID)
	} else {
		for _, l := range g.rgs {
			r := l.Find(id)
			if r != nil {
				return *r, true
			}
		}
	}
	return nil, false
}

func (g *Core) eventHandleMovement(sender RcTx, e EventMovement) {
	if p, ok := g.findRegoter(e.RgId); ok {
		moved := g.updatedMove(p, sender, e)
		if moved && (p.rgType == RegoterEnumPlayer) {
			g.updatePlayerCamera(&p.entity, moved, false)
		}
		if e.Command.StopAnimation {
			p.state.AnimationRunning = false
		}
		if e.Command.StartAnimation {
			p.state.AnimationRunning = true
		}
		e := EventUpdateData{RgEntity: p.entity, RgState: p.state}
		m := ReactorEventMessage{g.tx, e}
		p.tx <- m
	} else {
		log.Printf("Warning: Can not find Regoter(%v) in Event(%T).", e.RgId, e)
	}
}

func (g *Core) eventHandleDamage(sender RcTx, e EventDamage) {
	if e.peer == NULL_ID {
		log.Fatalf("ID can not be NULL_ID(%v).", NULL_ID)
	}
	if e.peer != WALL_ID {
		if p, ok := g.findRegoter(e.peer); ok {
			m := ReactorEventMessage{g.tx, EventDamage{peer: 0, damage: e.damage}}
			p.tx <- m
		} else {
			log.Printf("Warning: Can not find Regoter(%v) in Event(%T).", e.peer, e)
		}
	}
}

func (g *Core) updatedMove(p *regoterInCore, sender RcTx, e EventMovement) bool {
	pe := &p.entity
	rgType := pe.RgType
	velocity := math.Max(pe.Velocity+e.Move.Acceleration, 0)
	velocity = math.Max(velocity*(1-pe.Resistance), 0)

	vAngle := simplifyAngle(pe.Angle + e.Move.VissionRotate)
	mAngle := simplifyAngle(vAngle + e.Move.MoveRotate)
	mPitch := simplifyAngle(pe.Pitch + e.Move.PitchRotate)

	moved := false
	if math.Abs(velocity) > MinimumVelocity {
		var checkAlternate bool
		var lineEnd *Position
		if rgType == RegoterEnumProjectile {
			trajectory := geom3d.Line3dFromAngle(pe.Position.X, pe.Position.Y, pe.Position.Z,
				mAngle, mPitch, velocity)
			lineEnd = &Position{X: trajectory.X2, Y: trajectory.Y2, Z: trajectory.Z2}
			checkAlternate = false
		} else {
			moveLine := geom.LineFromAngle(pe.Position.X, pe.Position.Y, mAngle, velocity)
			lineEnd = &Position{X: moveLine.X2, Y: moveLine.Y2, Z: pe.Position.Z}
			checkAlternate = true
		}

		newPos, collisionEntity := g.getValidMove(pe, lineEnd.X, lineEnd.Y, lineEnd.Z, checkAlternate)
		if collisionEntity != nil {
			// Send EventCollistion to both Entities in collistion
			sender <- ReactorEventMessage{
				g.tx, EventCollision{collistion: *collisionEntity}}

			peerId := collisionEntity.peer
			if peerId != WALL_ID {
				if rg, ok := g.findRegoter(peerId); ok {
					collisionForPeer := EntityCollision{peer: pe.RgId, distance: collisionEntity.distance,
						position: collisionEntity.position}
					rg.tx <- ReactorEventMessage{
						g.tx, EventCollision{collistion: collisionForPeer}}
				} else {
					log.Printf("Warning: Can not find Peer Regoter(%v) in Event(%T).", peerId, e)
				}
			}
		} else {
			if lineEnd.Z < -1 {
				// Hit ground
				collision := EntityCollision{peer: WALL_ID, distance: 0,
					position: *lineEnd}
				sender <- ReactorEventMessage{
					g.tx, EventCollision{collistion: collision}}
			}
		}

		if newPos.X != pe.Position.X || newPos.Y != pe.Position.Y || lineEnd.Z != pe.Position.Z {
			pe.Position.X = newPos.X
			pe.Position.Y = newPos.Y
			pe.Position.Z = lineEnd.Z
			moved = true
		}
	}

	if vAngle != float64(pe.Angle) || mPitch != float64(pe.Pitch) || velocity != pe.Velocity {
		pe.Angle = vAngle
		pe.Pitch = geom.Clamp(mPitch, -math.Pi/8, math.Pi/4)
		pe.Velocity = limitVelocity(velocity, MaximumVelocity)
		moved = true
	}

	if pe.Velocity > MinimumVelocity {
		pe.LastMoveRotate = e.Move.MoveRotate
	} else {
		pe.LastMoveRotate = 0
	}

	return moved
}

func limitVelocity(velocity float64, max float64) float64 {
	if velocity > max {
		return max
	}
	if velocity < 0 {
		return 0
	}
	return velocity
}

func (g *Core) eventHandleRegoterUnregister(sender RcTx, e EventUnregisterRegoter) {
	for _, l := range g.rgs {
		r := l.Remove(e.RgId)
		if r {
			return
		}
	}
}

func (g *Core) eventHandleUnknown(sender RcTx, e IReactorEvent) error {
	log.Fatalf("Unknown event: %T", e)
	return nil
}

func NewCore(cfg GameCfg) RcTx {
	rc := NewReactorCore()
	var rgs [len(allRegoterEnum)]*stl4go.SkipList[ID, *regoterInCore]
	for i := 0; i < len(rgs); i++ {
		rgs[i] = stl4go.NewSkipList[ID, *regoterInCore]()
	}

	// load map
	mapObj := loader.NewMap()
	collisionMap := mapObj.GetCollisionLines(loader.ClipDistance)
	tex := loader.LoadContent(mapObj)

	worldMap := mapObj.Level(0)
	mapWidth := len(worldMap)
	mapHeight := len(worldMap[0])

	debugMessages := stl4go.NewDList[string]()
	core := &Core{Reactor: rc, rgs: rgs,
		mapObj: mapObj, collisionMap: collisionMap,
		mapWidth: mapWidth, mapHeight: mapHeight,
		debugMessages: debugMessages, tex: tex, cfg: cfg,
	}

	core.applyConfig()

	go core.Reactor.Run(core)
	return core.tx
}
func (core *Core) applyConfig() {
	cfg := core.cfg
	//--init camera and renderer--//
	// use scale to keep the desired window width and height
	// load content once when first run
	// load texture handler
	if cfg.Debug {
		core.tex.FloorTex = loader.GetRGBAFromFile("grass_debug.png")
	} else {
		core.tex.FloorTex = loader.GetRGBAFromFile("grass.png")
	}

	core.camera = raycaster.NewCamera(cfg.Width, cfg.Height, loader.TexWidth,
		core.mapObj, core.tex)
	//log.Printf("%+v", cfg)
	core.setResolution(cfg.ScreenWidth, cfg.ScreenHeight)
	core.setRenderScale(cfg.RenderScale)
	core.setFullscreen(cfg.Fullscreen)
	core.setVsyncEnabled(cfg.Vsync)

	core.setRenderDistance(cfg.RenderDistance)

	core.camera.SetFloorTexture(loader.GetTextureFromFile("floor.png"))
	core.camera.SetSkyTexture(loader.GetTextureFromFile("sky.png"))

	core.setFovAngle(cfg.FovDegrees)
	core.cfg.FovDepth = core.camera.FovDepth()

	core.cfg.ZoomFovDepth = 2.0

	// set demo lighting settings
	core.setLightFalloff(-200)
	core.setGlobalIllumination(500)
	minLightRGB := color.NRGBA{R: 76, G: 76, B: 76}
	maxLightRGB := color.NRGBA{R: 255, G: 255, B: 255}
	core.setLightRGB(minLightRGB, maxLightRGB)

}

// Update camera to match player position and orientation
func (g *Core) updatePlayerCamera(pe *Entity, moved bool, forceUpdate bool) {
	if !moved && !forceUpdate {
		// only update camera position if player moved or forceUpdate set
		return
	}

	g.camera.SetPosition(&geom.Vector2{X: pe.Position.X, Y: pe.Position.Y})
	CameraZ := 0.5
	g.camera.SetPositionZ(CameraZ)
	g.camera.SetHeadingAngle(pe.Angle)
	g.camera.SetPitchAngle(pe.Pitch)
}

func (g *Core) setFullscreen(fullscreen bool) {
	g.cfg.Fullscreen = fullscreen
	ebiten.SetFullscreen(fullscreen)
}

func (g *Core) setResolution(screenWidth, screenHeight int) {
	g.cfg.ScreenWidth, g.cfg.ScreenHeight = screenWidth, screenHeight
	ebiten.SetWindowSize(screenWidth, screenHeight)
	g.setRenderScale(g.cfg.RenderScale)
}

func (g *Core) setRenderScale(renderScale float64) {
	g.cfg.RenderScale = renderScale
	g.cfg.Width = int(math.Floor(float64(g.cfg.ScreenWidth) * g.cfg.RenderScale))
	g.cfg.Height = int(math.Floor(float64(g.cfg.ScreenHeight) * g.cfg.RenderScale))
	if g.camera != nil {
		g.camera.SetViewSize(g.cfg.Width, g.cfg.Height)
	}
	g.scene = ebiten.NewImage(g.cfg.Width, g.cfg.Height)
}

func (g *Core) setRenderDistance(renderDistance float64) {
	g.cfg.RenderDistance = renderDistance
	g.camera.SetRenderDistance(g.cfg.RenderDistance)
}

func (g *Core) setLightFalloff(lightFalloff float64) {
	g.cfg.LightFalloff = lightFalloff
	g.camera.SetLightFalloff(g.cfg.LightFalloff)
}

func (g *Core) setGlobalIllumination(globalIllumination float64) {
	g.cfg.GlobalIllumination = globalIllumination
	g.camera.SetGlobalIllumination(g.cfg.GlobalIllumination)
}

func (g *Core) setLightRGB(minLightRGB, maxLightRGB color.NRGBA) {
	g.cfg.MinLightRGB = minLightRGB
	g.cfg.MaxLightRGB = maxLightRGB
	g.camera.SetLightRGB(g.cfg.MinLightRGB, g.cfg.MaxLightRGB)
}

func (g *Core) setVsyncEnabled(enableVsync bool) {
	g.cfg.Vsync = enableVsync
	ebiten.SetVsyncEnabled(enableVsync)
}

func (g *Core) setFovAngle(fovDegrees float64) {
	g.cfg.FovDegrees = fovDegrees
	g.camera.SetFovAngle(fovDegrees, 1.0)
}

func (g *Core) setRenderFloorTexture(r bool) {
	//Todo
	//g.tex.RenderFloorTex = r
}

func simplifyAngle(angle float64) float64 {
	for angle <= -geom.Pi {
		angle = angle + geom.Pi2
	}
	for angle > geom.Pi {
		angle = angle - geom.Pi2
	}
	return angle
}
