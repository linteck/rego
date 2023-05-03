package game

import (
	"fmt"
	"image/color"
	"lintech/rego/game/loader"
	"lintech/rego/game/model"
	"lintech/rego/iregoter"
	"log"
	"math"
	"os"
	"time"

	"github.com/chen3feng/stl4go"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
)

var logger = log.New(os.Stdout, "Core ", 0)

type regoterInCore struct {
	tx     iregoter.CoreTxMsgbox
	rgType iregoter.RegoterEnum
	sprite *iregoter.Sprite
	state  iregoter.RegoterState
	entity iregoter.Entity
	di     iregoter.DrawInfo
}

type regoterSpace struct {
	entiry iregoter.CoreTxMsgbox
}

var imgLayerPriorities = [...]iregoter.ImgLayer{
	iregoter.ImgLayerPlayer,
	iregoter.ImgLayerSprite,
}

var allRegoterEnum = [...]iregoter.RegoterEnum{
	iregoter.RegoterEnumSprite,
	iregoter.RegoterEnumProjectile,
	iregoter.RegoterEnumEffect,
	iregoter.RegoterEnumCrosshair,
	iregoter.RegoterEnumPlayer,
}

type Core struct {
	rxBox    iregoter.CoreRxMsgbox
	txToGame iregoter.CoreTxMsgbox
	cfg      iregoter.GameCfg
	rgs      [len(allRegoterEnum)]*stl4go.SkipList[iregoter.ID, *regoterInCore]
	//imgs     [len(imgLayerPriorities)]*stl4go.DList[iregoter.RegoterUpdatedImg]

	// Camera
	camera        *raycaster.Camera
	scene         *ebiten.Image
	debugMessages *stl4go.DList[string]
	//--array of levels, levels refer to "floors" of the world--//
	mapObj       *loader.Map
	collisionMap []geom.Line
	mapWidth     int
	mapHeight    int
}

func (g *Core) eventHandleGameEventTick(e iregoter.GameEventTick) {
	for _, l := range g.rgs {
		l.ForEach(func(k iregoter.ID, v *regoterInCore) {
			e := iregoter.CoreEventUpdateTick{}
			v.tx <- e
		})
	}

	for _, l := range g.rgs {
		l.ForEach(func(k iregoter.ID, v *regoterInCore) {
			if v.sprite != nil {
				v.sprite.Update(g.camera.GetPosition())
			}
		})
	}

}

// func (g *Core) eventHandleGameEventTick(e iregoter.GameEventTick) {
// 	player := g.rgs[iregoter.RegoterEnumPlayer].Iterate().Value()
// 	for _, l := range g.rgs {
// 		l.ForEach(func(k iregoter.ID, v regoterInCore) {
// 			e := iregoter.CoreEventUpdateData{RgEntity: v.entity, PlayEntiry: player.entity,
// 				RgState: v.state}
// 			v.tx <- e
// 		})
// 	}
// 	// logger.Print(fmt.Sprintf("current rg num %v", g.rgs.Len()))
// }

func (g *Core) eventHandleGameEventCfgChanged(e iregoter.GameEventCfgChanged) {
	g.cfg = e.Cfg
	for _, l := range g.rgs {
		l.ForEach(func(k iregoter.ID, v *regoterInCore) {
			v.tx <- e
		})
	}
	// logger.Print(fmt.Sprintf("current rg num %v", g.rgs.Len()))
	g.applyConfig()
}

func (g *Core) eventHandleEventDebugPrint(e iregoter.EventDebugPrint) {
	g.debugMessages.PushBack(e.DebugString)
}

func createCoreSprite(rg *regoterInCore) *iregoter.Sprite {
	var sprite *iregoter.Sprite
	if rg.di.Img == nil {
		logger.Printf("Warning!!! Register Regoter without Img. Will not create Sprite for it.")
		return nil
	}
	if rg.di.AnimationRate != 0 {
		sprite = iregoter.NewAnimatedSprite(&rg.entity, rg.di.Img,
			rg.di.Columns, rg.di.Rows, rg.di.AnimationRate)
	} else {
		sprite = iregoter.NewSpriteFromSheet(&rg.entity, rg.di.Img,
			rg.di.Columns, rg.di.Rows, rg.di.SpriteIndex)
		sprite.SetAnimationFrame(rg.di.HitIndex)
	}
	sprite.SetIllumination(rg.di.Illumination)

	return sprite

}

func (g *Core) eventHandleRegisterRegoter(e iregoter.RegoterEventRegisterRegoter) {
	d := e.RgData
	rg := &regoterInCore{tx: e.Msgbox, rgType: d.Entity.RgType, entity: d.Entity, di: d.DrawInfo}
	if rg.di.AnimationRate != 0 && rg.di.SpriteIndex != 0 {
		logger.Fatal("This Regoter can not be both Animation and Sheet")
	}
	if rg.di.Img == nil && d.Entity.RgType != iregoter.RegoterEnumPlayer {
		logger.Fatal("Invalid nil Img for ", d.Entity.RgType, d.Entity.RgId)
	}
	rg.sprite = createCoreSprite(rg)
	rg.state.Unregistered = false
	rg.state.HasCollision = false
	g.rgs[rg.rgType].Insert(d.Entity.RgId, rg)
}

// func (g *Core) eventHandleUpdatedImg(e iregoter.RegoterEventUpdatedImg) {
// 	l := g.imgs[e.Info.ImgLayer]
// 	l.PushBack(e.Info)
// }

func (g *Core) findRegoter(id iregoter.ID) (*regoterInCore, bool) {
	for _, l := range g.rgs {
		r := l.Find(id)
		if r != nil {
			return *r, true
		}
	}
	return nil, false
}

func (g *Core) eventHandleUpdatedMove(e iregoter.RegoterEventUpdatedMove) {
	if p, ok := g.findRegoter(e.RgId); ok {
		moved := g.updatedMove(p, e)
		if moved && (p.rgType == iregoter.RegoterEnumPlayer) {
			g.updatePlayerCamera(&p.entity, moved, false)
		}
		e := iregoter.CoreEventUpdateData{RgEntity: p.entity, RgState: p.state}
		p.tx <- e
	}
}

func (g *Core) updatedMove(p *regoterInCore, e iregoter.RegoterEventUpdatedMove) bool {
	pe := &p.entity
	rgType := pe.RgType
	velocity := math.Max(pe.Velocity+e.Move.Acceleration, 0)
	velocity = math.Max(velocity*(1-pe.Resistance), 0)

	vAngle := simplifyAngle(pe.Angle + e.Move.VissionRotate)
	mAngle := simplifyAngle(vAngle + e.Move.MoveRotate)
	mPitch := simplifyAngle(pe.Pitch + e.Move.PitchRotate)

	moved := false
	if math.Abs(velocity) > model.MinimumVelocity {
		var checkAlternate bool
		var lineEnd *iregoter.Position
		if rgType == iregoter.RegoterEnumProjectile {
			trajectory := geom3d.Line3dFromAngle(pe.Position.X, pe.Position.Y, pe.Position.Z,
				mAngle, mPitch, velocity)
			lineEnd = &iregoter.Position{X: trajectory.X2, Y: trajectory.Y2, Z: trajectory.Z2}
			checkAlternate = false
		} else {
			moveLine := geom.LineFromAngle(pe.Position.X, pe.Position.Y, mAngle, velocity)
			lineEnd = &iregoter.Position{X: moveLine.X2, Y: moveLine.Y2, Z: pe.Position.Z}
			checkAlternate = true
		}
		newPos, hasCollision, _ := g.getValidMove(pe, lineEnd.X, lineEnd.Y, lineEnd.Z, checkAlternate)

		// Hit ground
		if pe.Position.Z < 0 {
			hasCollision = true
		}

		p.state.HasCollision = hasCollision

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
		pe.Velocity = limitVelocity(velocity, model.MaximumVelocity)
		moved = true
	}

	if pe.Velocity > model.MinimumVelocity {
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

func (g *Core) eventHandleRegoterUnregister(e iregoter.RegoterEventRegoterUnregister) {
	// Mark Deleted. Only delete it after drawing
	if rg, ok := g.findRegoter(e.RgId); ok {
		rg.state.Unregistered = true
	}
}

func (r *Core) eventHandleUnknown(e iregoter.IRegoterEvent) error {
	logger.Fatal(fmt.Sprintf("Unknown event: %T", e))
	return nil
}

func (g *Core) removeAllUnregisteredRogeter() {
	for _, l := range g.rgs {
		ids := make([]iregoter.ID, l.Len())
		index := 0
		l.ForEach(func(id iregoter.ID, val *regoterInCore) {
			if val.state.Unregistered {
				ids[index] = id
				index++
			}
		})
		for i := range ids[:index] {
			l.Remove(ids[i])
		}
	}
}

func (g *Core) process(e iregoter.IRegoterEvent) error {
	// logger.Print(fmt.Sprintf("core event recv %T", e))
	switch e.(type) {
	case iregoter.EventDebugPrint:
		g.eventHandleEventDebugPrint(e.(iregoter.EventDebugPrint))
	case iregoter.GameEventTick:
		g.eventHandleGameEventTick(e.(iregoter.GameEventTick))

	case iregoter.GameEventCfgChanged:
		g.eventHandleGameEventCfgChanged(e.(iregoter.GameEventCfgChanged))
	case iregoter.GameEventDraw:
		g.eventHandleGameEventDraw(e.(iregoter.GameEventDraw))

	case iregoter.RegoterEventRegisterRegoter:
		g.eventHandleRegisterRegoter(e.(iregoter.RegoterEventRegisterRegoter))
	case iregoter.RegoterEventRegoterUnregister:
		g.eventHandleRegoterUnregister(e.(iregoter.RegoterEventRegoterUnregister))
	// case iregoter.RegoterEventUpdatedImg:
	// 	g.eventHandleUpdatedImg(e.(iregoter.RegoterEventUpdatedImg))

	case iregoter.RegoterEventUpdatedMove:
		g.eventHandleUpdatedMove(e.(iregoter.RegoterEventUpdatedMove))
	default:
		g.eventHandleUnknown(e)
	}
	return nil
}

func (g *Core) Run() {
	for {
		select {
		case e := <-g.rxBox:
			err := g.process(e)
			if err != nil {
				logger.Fatal(err)
			}
		case <-time.After(time.Millisecond * 1000):
			logger.Print("Core has not received any message in 1 second")
			// err := g.update()
			// if err != nil {
			// 	fmt.Println(err)
			// }
		}
	}
}

func NewCore(cfg *iregoter.GameCfg) (iregoter.RgTxMsgbox, iregoter.RgRxMsgbox) {
	c := make(chan iregoter.IRegoterEvent)
	g := make(chan iregoter.ICoreEvent)

	var rgs [len(allRegoterEnum)]*stl4go.SkipList[iregoter.ID, *regoterInCore]
	for i := 0; i < len(rgs); i++ {
		rgs[i] = stl4go.NewSkipList[iregoter.ID, *regoterInCore]()
	}

	// var imgs [len(imgLayerPriorities)]*stl4go.DList[iregoter.RegoterUpdatedImg]
	// for i := 0; i < len(imgs); i++ {
	// 	imgs[i] = stl4go.NewDList[iregoter.RegoterUpdatedImg]()
	// }

	// load map
	mapObj := loader.NewMap()
	collisionMap := mapObj.GetCollisionLines(loader.ClipDistance)

	worldMap := mapObj.Level(0)
	mapWidth := len(worldMap)
	mapHeight := len(worldMap[0])

	// load content once when first run
	tex := loader.LoadContent(mapObj)
	if cfg.Debug {
		tex.FloorTex = loader.GetRGBAFromFile("grass_debug.png")
	} else {
		tex.FloorTex = loader.GetRGBAFromFile("grass.png")
	}

	// load texture handler
	//tex := NewTextureHandler(mapObj, 32)
	tex.RenderFloorTex = cfg.RenderFloorTex

	camera := raycaster.NewCamera(cfg.Width, cfg.Height, loader.TexWidth, mapObj, tex)
	debugMessages := stl4go.NewDList[string]()
	core := &Core{rxBox: c, txToGame: g, rgs: rgs,
		mapObj: mapObj, collisionMap: collisionMap,
		mapWidth: mapWidth, mapHeight: mapHeight, camera: camera, cfg: *cfg,
		debugMessages: debugMessages,
	}

	core.applyConfig()

	go core.Run()
	return c, g
}
func (core *Core) applyConfig() {
	//--init camera and renderer--//
	// use scale to keep the desired window width and height
	cfg := core.cfg
	//logger.Printf("%+v", cfg)
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
func (g *Core) updatePlayerCamera(pe *iregoter.Entity, moved bool, forceUpdate bool) {
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
