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
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
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
	iregoter.RegoterEnumPlayer,
	iregoter.RegoterEnumSprite,
	iregoter.RegoterEnumProjectile,
	iregoter.RegoterEnumEffect}

type Core struct {
	rxBox    iregoter.CoreRxMsgbox
	txToGame iregoter.CoreTxMsgbox
	cfg      iregoter.GameCfg
	rgs      [len(allRegoterEnum)]*stl4go.SkipList[iregoter.ID, regoterInCore]
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
	player := g.rgs[iregoter.RegoterEnumPlayer].Iterate().Value()
	for _, l := range g.rgs {
		l.ForEach(func(k iregoter.ID, v regoterInCore) {
			e := iregoter.CoreEventUpdateTick{RgEntity: v.entity, PlayEntiry: player.entity,
				RgState: v.state}
			v.tx <- e
		})
	}
	// logger.Print(fmt.Sprintf("current rg num %v", g.rgs.Len()))
}

func (g *Core) eventHandleGameEventCfgChanged(e iregoter.GameEventCfgChanged) {
	g.cfg = e.Cfg
	for _, l := range g.rgs {
		l.ForEach(func(k iregoter.ID, v regoterInCore) {
			v.tx <- e
		})
	}
	// logger.Print(fmt.Sprintf("current rg num %v", g.rgs.Len()))
	g.applyConfig()
}

func (g *Core) eventHandleEventDebugPrint(e iregoter.EventDebugPrint) {
	g.debugMessages.PushBack(e.DebugString)
}

func (g *Core) eventHandleNewRegoter(e iregoter.RegoterEventNewRegoter) {
	d := e.RgData
	rg := regoterInCore{tx: e.Msgbox, rgType: d.Entity.RgType, entity: d.Entity, di: d.DrawInfo}
	if rg.di.AnimationRate != 0 && rg.di.SpriteIndex != 0 {
		logger.Fatal("This Regoter can not be both Animation and Sheet")
	}
	if rg.di.Img == nil {
		logger.Fatal("Invalid nil Img for ", d.Entity.RgType, d.Entity.RgId)
	}
	var sprite *iregoter.Sprite
	if rg.di.AnimationRate != 0 {
		sprite = iregoter.NewAnimatedSprite(&rg.entity, rg.di.Img,
			rg.di.Columns, rg.di.Rows, rg.di.AnimationRate)
	} else {
		sprite = iregoter.NewSpriteFromSheet(&rg.entity, rg.di.Img,
			rg.di.Columns, rg.di.Rows, rg.di.SpriteIndex)
	}
	sprite.SetIllumination(rg.di.Illumination)
	rg.sprite = sprite
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
			return r, true
		}
	}
	return nil, false
}

func (g *Core) eventHandleUpdatedMove(e iregoter.RegoterEventUpdatedMove) {
	if p, ok := g.findRegoter(e.RgId); ok {
		if p.rgType == iregoter.RegoterEnumProjectile {
			g.eventHandleUpdated3DMove(p, e)
		} else {
			moved := g.eventHandleUpdated2DMove(p, e)
			if p.rgType == iregoter.RegoterEnumPlayer {
				g.updatePlayerCamera(&p.entity, moved, false)
			}
		}
	}
}

func (g *Core) eventHandleUpdated3DMove(p *regoterInCore, e iregoter.RegoterEventUpdatedMove) bool {
	pe := &p.entity

	mSpeed := float64(e.Move.MoveSpeed)
	velocity := pe.Velocity + mSpeed

	rotateSpeed := e.Move.RotateSpeed
	mAngle := float64(pe.Angle + rotateSpeed)
	mAngle = simplifyAngle(mAngle)

	pitchSpeed := e.Move.PitchSpeed
	mPitch := float64(pe.Pitch + pitchSpeed)
	mPitch = simplifyAngle(mPitch)

	moved := false
	if math.Abs(velocity) > model.MinimumVelocity {
		trajectory := geom3d.Line3dFromAngle(pe.Position.X, pe.Position.Y, pe.Position.Z,
			float64(mAngle), float64(mPitch), velocity)
		xCheck := trajectory.X2
		yCheck := trajectory.Y2
		zCheck := trajectory.Z2
		newPos, isCollision, collisions := g.getValidMove(&p.entity, xCheck, yCheck, zCheck, false)

		if isCollision && len(collisions) >= 1 {
			// for testing purposes, projectiles instantly get deleted when collision occurs
			// use the first collision point to place effect at
			// Bug: Why?
			newPos = collisions[0].collision
		}
		if pe.Position.Z < 0 {
			// Hit ground
			isCollision = true
		}

		p.state.HasCollision = isCollision

		if newPos.X != pe.Position.X || newPos.Y != pe.Position.Y || zCheck != pe.Position.Z {
			pe.Position.X = newPos.X
			pe.Position.Y = newPos.Y
			pe.Position.Z = zCheck
			moved = true
		}
	}

	if mAngle != float64(pe.Angle) || mPitch != float64(pe.Pitch) || velocity != pe.Velocity {
		pe.Angle = iregoter.RotateAngle(mAngle)
		pe.Pitch = iregoter.PitchAngle(geom.Clamp(mPitch, -math.Pi/8, math.Pi/4))
		pe.Velocity = limitVelocity(velocity, model.MaximumVelocity)
		moved = true
	}

	return moved
}

func (g *Core) eventHandleUpdated2DMove(p *regoterInCore, e iregoter.RegoterEventUpdatedMove) bool {
	pe := &p.entity
	mSpeed := float64(e.Move.MoveSpeed)
	velocity := pe.Velocity + mSpeed

	rotateSpeed := e.Move.RotateSpeed
	mAngle := float64(pe.Angle + rotateSpeed)
	mAngle = simplifyAngle(mAngle)

	pitchSpeed := e.Move.PitchSpeed
	mPitch := float64(pe.Pitch + pitchSpeed)
	mPitch = simplifyAngle(mPitch)

	moved := false
	if math.Abs(velocity) > model.MinimumVelocity {
		moveLine := geom.LineFromAngle(pe.Position.X, pe.Position.Y, float64(pe.Angle), velocity)

		newPos, collision, _ := g.getValidMove(pe, moveLine.X2, moveLine.Y2, pe.Position.Z, true)
		p.state.HasCollision = collision

		if newPos.X != pe.Position.X || newPos.Y != pe.Position.Y {
			moved = true
			pe.Position.X = newPos.X
			pe.Position.Y = newPos.Y
		}
	}
	if mAngle != float64(pe.Angle) || mPitch != float64(pe.Pitch) || velocity != pe.Velocity {
		pe.Angle = iregoter.RotateAngle(mAngle)
		pe.Pitch = iregoter.PitchAngle(geom.Clamp(mPitch, -math.Pi/8, math.Pi/4))
		pe.Velocity = limitVelocity(velocity, model.MaximumVelocity)
		moved = true
	}
	return moved
}

func limitVelocity(velocity float64, max float64) float64 {
	if velocity > max {
		return max
	}
	if velocity < -max {
		return -max
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

func (g *Core) eventHandleGameEventDraw(e iregoter.GameEventDraw) {

	// for _, l := range g.imgs {
	// 	l.ForEach(func(v iregoter.RegoterUpdatedImg) {
	// 		e.Screen.DrawImage(v.Img, v.ImgOp)
	// 	})
	// 	l.Clear()
	// }
	g.drawScreen(e.Screen)

	g.removeAllUnregisteredRogeter()
	r := iregoter.CoreEventDrawDone{}
	g.txToGame <- r

}

func (g *Core) removeAllUnregisteredRogeter() {
	for _, l := range g.rgs {
		ids := make([]iregoter.ID, l.Len())
		index := 0
		l.ForEach(func(id iregoter.ID, val regoterInCore) {
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
	switch e.(type) {
	case iregoter.EventDebugPrint:
		g.eventHandleEventDebugPrint(e.(iregoter.EventDebugPrint))
	case iregoter.GameEventTick:
		g.eventHandleGameEventTick(e.(iregoter.GameEventTick))
	case iregoter.GameEventCfgChanged:
		g.eventHandleGameEventCfgChanged(e.(iregoter.GameEventCfgChanged))
	case iregoter.GameEventDraw:
		g.eventHandleGameEventDraw(e.(iregoter.GameEventDraw))
	case iregoter.RegoterEventNewRegoter:
		g.eventHandleNewRegoter(e.(iregoter.RegoterEventNewRegoter))
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

func NewCore(cfg *iregoter.GameCfg) (chan<- iregoter.IRegoterEvent, <-chan iregoter.ICoreEvent) {
	c := make(chan iregoter.IRegoterEvent)
	g := make(chan iregoter.ICoreEvent)

	var rgs [len(allRegoterEnum)]*stl4go.SkipList[iregoter.ID, regoterInCore]
	for i := 0; i < len(rgs); i++ {
		rgs[i] = stl4go.NewSkipList[iregoter.ID, regoterInCore]()
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

func (g *Core) drawScreen(screen *ebiten.Image) {
	// Put projectiles together with sprites for raycasting both as sprites
	sl := g.rgs[iregoter.RegoterEnumSprite]
	numSprites := sl.Len()
	raycastSprites := make([]raycaster.Sprite, numSprites)

	index := 0
	sl.ForEach(func(i iregoter.ID, val regoterInCore) {
		raycastSprites[index] = val.sprite
		index += 1
	})

	// Debug
	// g.camera.SetHeadingAngle(geom.Pi / 2)
	// g.camera.SetPosition(&geom.Vector2{X: 10, Y: 10})
	// CameraZ := 0.5
	// g.camera.SetPositionZ(CameraZ)
	// End of Debug

	// Update camera (calculate raycast)
	g.camera.Update(raycastSprites)

	// Render raycast scene
	g.camera.Draw(g.scene)

	if g.cfg.ShowSpriteBoxes {
		// draw sprite screen indicators to show we know where it was raycasted (must occur after camera.Update)
		sl.ForEach(func(i iregoter.ID, val regoterInCore) {
			drawSpriteBox(g.scene, val.sprite)
		})
	}

	// Todo
	// // draw sprite screen indicator only for sprite at point of convergence
	// convergenceSprite := g.camera.GetConvergenceSprite()
	// if convergenceSprite != nil {
	// 	for sprite := range g.sprites {
	// 		if convergenceSprite == sprite {
	// 			drawSpriteIndicator(g.scene, sprite)
	// 			break
	// 		}
	// 	}
	// }

	// draw raycasted scene
	op := &ebiten.DrawImageOptions{}
	// Todo
	if g.cfg.RenderScale < 1 {
		op.Filter = ebiten.FilterNearest
		op.GeoM.Scale(1/g.cfg.RenderScale, 1/g.cfg.RenderScale)
	}
	screen.DrawImage(g.scene, op)

	g.drawDebugInfo(screen)
	g.drawMiniMap(screen)

}

func (g *Core) drawDebugInfo(screen *ebiten.Image) {
	// draw FPS/TPS counter debug display
	dbgMsg := fmt.Sprintf("FPS: %.1f\nTPS: %.1f/%v\n", ebiten.ActualFPS(), ebiten.ActualTPS(), ebiten.TPS())
	//ebitenutil.DebugPrint(screen, fps)
	g.debugMessages.ForEach(func(val string) {
		dbgMsg += (val + "\n")
	})
	g.debugMessages.Clear()
	ebitenutil.DebugPrint(screen, dbgMsg)
}

func (g *Core) drawMiniMap(screen *ebiten.Image) {
	// draw minimap
	mm := g.miniMap()
	mmImg := ebiten.NewImageFromImage(mm)
	if mmImg != nil {
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest

		op.GeoM.Scale(5.0, 5.0)
		op.GeoM.Translate(0, 50)
		screen.DrawImage(mmImg, op)
	}
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
	g.camera.SetHeadingAngle(float64(pe.Angle))
	g.camera.SetPitchAngle(float64(pe.Pitch))
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
	if angle > geom.Pi {
		angle = angle - geom.Pi2
	} else if angle <= -geom.Pi {
		angle = angle + geom.Pi2
	}
	return angle
}
