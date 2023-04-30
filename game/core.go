package game

import (
	"fmt"
	"image/color"
	"lintech/rego/game/loader"
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

type coreRxMsgbox <-chan iregoter.IRegoterEvent
type coreTxMsgbox chan<- iregoter.ICoreEvent

type regoterInCore struct {
	tx      coreTxMsgbox
	rgType  iregoter.RegoterEnum
	entity  *iregoter.Entity
	deleted bool
}

type regoterSpace struct {
	entiry coreTxMsgbox
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
	rxBox    coreRxMsgbox
	txToGame chan<- iregoter.ICoreEvent
	rgs      [len(allRegoterEnum)]*stl4go.SkipList[iregoter.ID, regoterInCore]
	imgs     [len(imgLayerPriorities)]*stl4go.DList[iregoter.RegoterUpdatedImg]
	cfg      *iregoter.GameCfg

	// Camera
	camera *raycaster.Camera
	scene  *ebiten.Image
	//--array of levels, levels refer to "floors" of the world--//
	mapObj       *loader.Map
	collisionMap []geom.Line
	mapWidth     int
	mapHeight    int
}

func (g *Core) eventHandleGameEventTick(e iregoter.GameEventTick) {
	g.update(e.ScreenSize)
}

func (g *Core) eventHandleNewRegoter(e iregoter.RegoterEventNewRegoter) {
	g.rgs[e.RgType].Insert(e.RgId, regoterInCore{tx: e.Msgbox, rgType: e.RgType, entity: e.Entity})
}

func (g *Core) eventHandleUpdatedImg(e iregoter.RegoterEventUpdatedImg) {
	l := g.imgs[e.Info.ImgLayer]
	l.PushBack(e.Info)
}

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
				g.updatePlayerCamera(p.entity, moved, false)
			}
		}
	}
}

func (g *Core) eventHandleUpdated3DMove(p *regoterInCore, e iregoter.RegoterEventUpdatedMove) {
	pe := p.entity
	mSpeed := float64(e.Move.MoveSpeed)
	rotateSpeed := e.Move.RotateSpeed
	mAngle := float64(pe.Angle + rotateSpeed)
	pitchSpeed := e.Move.PitchSpeed

	pi2 := geom.Pi2
	if mAngle >= pi2 {
		mAngle = mAngle - pi2
	} else if mAngle <= 0.0 {
		mAngle = mAngle + pi2
	}

	trajectory := geom3d.Line3dFromAngle(pe.Position.X, pe.Position.Y, pe.PositionZ,
		float64(pe.Angle+rotateSpeed), float64(pe.Pitch+pitchSpeed), pe.Velocity+mSpeed)
	xCheck := trajectory.X2
	yCheck := trajectory.Y2
	zCheck := trajectory.Z2
	newPos, isCollision, collisions := g.getValidMove(p.entity, xCheck, yCheck, zCheck, false)

	if isCollision || pe.PositionZ <= 0 {
		// for testing purposes, projectiles instantly get deleted when collision occurs
		p.deleted = true
		if len(collisions) >= 1 {
			// use the first collision point to place effect at
			newPos = collisions[0].collision
		}
	}
	pe.Position = newPos
	pe.PositionZ = zCheck

}

func (g *Core) eventHandleUpdated2DMove(p *regoterInCore, e iregoter.RegoterEventUpdatedMove) bool {
	pe := p.entity
	mSpeed := float64(e.Move.MoveSpeed)
	rotateSpeed := e.Move.RotateSpeed
	mAngle := float64(pe.Angle + rotateSpeed)
	pitchSpeed := e.Move.PitchSpeed

	pi2 := geom.Pi2
	if mAngle >= pi2 {
		mAngle = mAngle - pi2
	} else if mAngle <= 0.0 {
		mAngle = mAngle + pi2
	}

	moveLine := geom.LineFromAngle(pe.Position.X, pe.Position.Y, mAngle, mSpeed)
	newPos, _, _ := g.getValidMove(pe, moveLine.X2, moveLine.Y2, pe.PositionZ, true)
	if !newPos.Equals(pe.Pos()) {
		pe.Position = newPos
		pe.Angle = iregoter.RotateAngle(mAngle)
		pe.Pitch = iregoter.PitchAngle(geom.Clamp(float64(pe.Pitch+pitchSpeed), -math.Pi/8, math.Pi/4))
		return true
	}
	return false
}

func (g *Core) eventHandleRegoterDeleted(e iregoter.RegoterEventRegoterDeleted) {
	for _, l := range g.rgs {
		l.Remove(e.RgId)
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

	r := iregoter.CoreEventDrawDone{}
	g.txToGame <- r

}

func (g *Core) process(e iregoter.IRegoterEvent) error {
	//logger.Print(fmt.Sprintf(" recv %T", e))
	switch e.(type) {
	case iregoter.GameEventTick:
		g.eventHandleGameEventTick(e.(iregoter.GameEventTick))
	case iregoter.GameEventDraw:
		g.eventHandleGameEventDraw(e.(iregoter.GameEventDraw))
	case iregoter.RegoterEventNewRegoter:
		g.eventHandleNewRegoter(e.(iregoter.RegoterEventNewRegoter))
	case iregoter.RegoterEventRegoterDeleted:
		g.eventHandleRegoterDeleted(e.(iregoter.RegoterEventRegoterDeleted))
	case iregoter.RegoterEventUpdatedImg:
		g.eventHandleUpdatedImg(e.(iregoter.RegoterEventUpdatedImg))
	case iregoter.RegoterEventUpdatedMove:
		g.eventHandleUpdatedMove(e.(iregoter.RegoterEventUpdatedMove))
	default:
		g.eventHandleUnknown(e)
	}
	return nil
}

func (g *Core) update(screenSize iregoter.ScreenSize) error {
	player := g.rgs[iregoter.RegoterEnumPlayer].Iterate().Value()
	for _, l := range g.rgs {
		l.ForEach(func(k iregoter.ID, v regoterInCore) {
			e := iregoter.CoreEventUpdateTick{RgEntity: v.entity, PlayEntiry: player.entity,
				HasCollision: false, ScreenSize: screenSize}
			v.tx <- e
		})
	}
	// logger.Print(fmt.Sprintf("current rg num %v", g.rgs.Len()))
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

	var imgs [len(imgLayerPriorities)]*stl4go.DList[iregoter.RegoterUpdatedImg]
	for i := 0; i < len(imgs); i++ {
		imgs[i] = stl4go.NewDList[iregoter.RegoterUpdatedImg]()
	}

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
	core := &Core{rxBox: c, txToGame: g, rgs: rgs, imgs: imgs,
		mapObj: mapObj, collisionMap: collisionMap,
		mapWidth: mapWidth, mapHeight: mapHeight, camera: camera, cfg: cfg}

	//--init camera and renderer--//
	// use scale to keep the desired window width and height
	core.setResolution(cfg.ScreenWidth, cfg.ScreenHeight)
	core.setRenderScale(cfg.RenderScale)
	core.setFullscreen(cfg.Fullscreen)
	core.setVsyncEnabled(cfg.Vsync)

	core.setRenderDistance(cfg.RenderDistance)

	camera.SetFloorTexture(loader.GetTextureFromFile("floor.png"))
	camera.SetSkyTexture(loader.GetTextureFromFile("sky.png"))

	core.setFovAngle(cfg.FovDegrees)
	core.cfg.FovDepth = camera.FovDepth()

	core.cfg.ZoomFovDepth = 2.0

	// set demo lighting settings
	core.setLightFalloff(-200)
	core.setGlobalIllumination(500)
	minLightRGB := color.NRGBA{R: 76, G: 76, B: 76}
	maxLightRGB := color.NRGBA{R: 255, G: 255, B: 255}
	core.setLightRGB(minLightRGB, maxLightRGB)

	go core.Run()
	return c, g
}

func (g *Core) drawScreen(screen *ebiten.Image) {
	// Put projectiles together with sprites for raycasting both as sprites
	sl := g.imgs[iregoter.ImgLayerSprite]
	numSprites := sl.Len()
	raycastSprites := make([]raycaster.Sprite, numSprites)

	index := 0
	sl.ForEach(func(val iregoter.RegoterUpdatedImg) {
		raycastSprites[index] = val.Sprite
		index += 1
	})

	// Update camera (calculate raycast)
	g.camera.Update(raycastSprites)

	// Render raycast scene
	g.camera.Draw(g.scene)

	pl := g.imgs[iregoter.ImgLayerPlayer]
	pl.ForEach(func(val iregoter.RegoterUpdatedImg) {
		g.scene.DrawImage(val.Img, val.ImgOp)
	})

	if g.cfg.ShowSpriteBoxes {
		// draw sprite screen indicators to show we know where it was raycasted (must occur after camera.Update)
		sl.ForEach(func(val iregoter.RegoterUpdatedImg) {
			drawSpriteBox(g.scene, val.Sprite)
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
	// if g.renderScale < 1 {
	// 	op.Filter = ebiten.FilterNearest
	// 	op.GeoM.Scale(1/g.renderScale, 1/g.renderScale)
	// }
	screen.DrawImage(g.scene, op)

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

	// draw FPS/TPS counter debug display
	fps := fmt.Sprintf("FPS: %f\nTPS: %f/%v", ebiten.ActualFPS(), ebiten.ActualTPS(), ebiten.TPS())
	ebitenutil.DebugPrint(screen, fps)
}

// Update camera to match player position and orientation
func (g *Core) updatePlayerCamera(pe *iregoter.Entity, moved bool, forceUpdate bool) {
	if !moved && !forceUpdate {
		// only update camera position if player moved or forceUpdate set
		return
	}

	g.camera.SetPosition(pe.Position.Copy())
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
