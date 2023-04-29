package game

import (
	"embed"
	"fmt"
	"image/color"
	"lintech/rego/iregoter"
	"lintech/rego/regoter"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"

	"lintech/rego/game/loader"
	"lintech/rego/game/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/spf13/viper"
)

//go:embed resources
var embedded embed.FS

const (
	// distance to keep away from walls and obstacles to avoid clipping
	// TODO: may want a smaller distance to test vs. sprites
	clipDistance = 0.1
)

type osType int

const (
	osTypeDesktop osType = iota
	osTypeBrowser
)

type gameTxMsgbox chan<- iregoter.IRegoterEvent
type gameRxMsgbox <-chan iregoter.ICoreEvent

type Game struct {
	txToCore   gameTxMsgbox
	rxFromCore gameRxMsgbox

	// Raycaster
	menu   *DemoMenu
	paused bool

	//--create slicer and declare slices--//
	tex                *TextureHandler
	initRenderFloorTex bool

	// window resolution and scaling
	screenWidth  int
	screenHeight int
	renderScale  float64
	fullscreen   bool
	vsync        bool
	fovDegrees   float64
	fovDepth     float64

	//--viewport width / height--//
	width  int
	height int

	player *model.Player

	//--define camera and render scene--//
	camera *raycaster.Camera
	scene  *ebiten.Image

	mouseMode      MouseMode
	mouseX, mouseY int

	crosshairs *regoter.Regoter[*model.Crosshairs]

	// zoom settings
	zoomFovDepth float64

	renderDistance float64

	// lighting settings
	lightFalloff       float64
	globalIllumination float64
	minLightRGB        color.NRGBA
	maxLightRGB        color.NRGBA

	//--array of levels, levels refer to "floors" of the world--//
	mapObj       *model.Map
	collisionMap []geom.Line

	sprites     map[*model.Sprite]struct{}
	projectiles map[*model.Projectile]struct{}
	effects     map[*model.Effect]struct{}

	mapWidth, mapHeight int

	showSpriteBoxes bool
	osType          osType
	debug           bool
}

// Update - Allows the game to run logic such as updating the world, gathering input, and playing audio.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	e := iregoter.GameEventTick{}
	g.txToCore <- e

	// handle input (when paused making sure only to allow input for closing menu so it can be unpaused)
	g.handleInput()
	if !g.paused {
		// Perform logical updates
		w := g.player.Weapon
		if w != nil {
			w.Update()
		}
		g.updateProjectiles()
		g.updateSprites()

		// handle player camera movement
		g.updatePlayerCamera(false)
	}

	// update the menu (if active)
	g.menu.update()

	return nil
}

func (g *Game) Run() {
	g.paused = false
	logger.Print("Start")
	if err := ebiten.RunGame(g); err != nil {
		logger.Fatal(err)
	}
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	e := iregoter.GameEventDraw{Screen: screen}
	g.txToCore <- e
	<-g.rxFromCore
	//logger.Print("Draw reply", r)

	// Put projectiles together with sprites for raycasting both as sprites
	numSprites, numProjectiles, numEffects := len(g.sprites), len(g.projectiles), len(g.effects)
	raycastSprites := make([]raycaster.Sprite, numSprites+numProjectiles+numEffects)
	index := 0
	for sprite := range g.sprites {
		raycastSprites[index] = sprite
		index += 1
	}
	for projectile := range g.projectiles {
		raycastSprites[index] = projectile.Sprite
		index += 1
	}
	for effect := range g.effects {
		raycastSprites[index] = effect.Sprite
		index += 1
	}

	// Update camera (calculate raycast)
	g.camera.Update(raycastSprites)

	// Render raycast scene
	g.camera.Draw(g.scene)

	// draw equipped weapon
	if g.player.Weapon != nil {
		w := g.player.Weapon
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest

		weaponScale := w.Scale() * g.renderScale
		op.GeoM.Scale(weaponScale, weaponScale)
		op.GeoM.Translate(
			float64(g.width)/2-float64(w.W)*weaponScale/2,
			float64(g.height)-float64(w.H)*weaponScale+1,
		)

		// apply lighting setting
		op.ColorScale.Scale(float32(g.maxLightRGB.R)/255, float32(g.maxLightRGB.G)/255, float32(g.maxLightRGB.B)/255, 1)

		g.scene.DrawImage(w.Texture(), op)
	}

	if g.showSpriteBoxes {
		// draw sprite screen indicators to show we know where it was raycasted (must occur after camera.Update)
		for sprite := range g.sprites {
			drawSpriteBox(g.scene, sprite)
		}

		for sprite := range g.projectiles {
			drawSpriteBox(g.scene, sprite.Sprite)
		}

		for sprite := range g.effects {
			drawSpriteBox(g.scene, sprite.Sprite)
		}
	}

	// draw sprite screen indicator only for sprite at point of convergence
	convergenceSprite := g.camera.GetConvergenceSprite()
	if convergenceSprite != nil {
		for sprite := range g.sprites {
			if convergenceSprite == sprite {
				drawSpriteIndicator(g.scene, sprite)
				break
			}
		}
	}

	// draw raycasted scene
	op := &ebiten.DrawImageOptions{}
	if g.renderScale < 1 {
		op.Filter = ebiten.FilterNearest
		op.GeoM.Scale(1/g.renderScale, 1/g.renderScale)
	}
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

	// draw menu (if active)
	g.menu.draw(screen)

	// draw FPS/TPS counter debug display
	fps := fmt.Sprintf("FPS: %f\nTPS: %f/%v", ebiten.ActualFPS(), ebiten.ActualTPS(), ebiten.TPS())
	ebitenutil.DebugPrint(screen, fps)
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	w, h := int(float64(g.screenWidth)), int(float64(g.screenHeight))
	g.menu.layout(w, h)
	return int(w), int(h)
}

// NewGame - Allows the game to perform any initialization it needs to before starting to run.
// This is where it can query for any required services and load any non-graphic
// related content.  Calling base.Initialize will enumerate through any components
// and initialize them as well.
func NewGame() *Game {
	fmt.Printf("Initializing Game\n")
	ebiten.SetWindowTitle("Rego Demo")

	rand.Seed(time.Now().UnixNano())

	// initialize Game object
	g := new(Game)
	txToCore, rxFromCore := NewCore()
	g.txToCore = txToCore
	g.rxFromCore = rxFromCore

	// for i := 0; i < 10; i++ {
	// 	regoter.NewSpiteWalker(txToCore)
	// }

	g.initConfig()

	// default TPS is 60
	// ebiten.SetMaxTPS(60)

	// use scale to keep the desired window width and height
	g.setResolution(g.screenWidth, g.screenHeight)
	g.setRenderScale(g.renderScale)
	g.setFullscreen(g.fullscreen)
	g.setVsyncEnabled(g.vsync)

	// load map
	g.mapObj = model.NewMap()

	// load texture handler
	g.tex = NewTextureHandler(g.mapObj, 32)
	g.tex.renderFloorTex = g.initRenderFloorTex

	g.collisionMap = g.mapObj.GetCollisionLines(clipDistance)
	worldMap := g.mapObj.Level(0)
	g.mapWidth = len(worldMap)
	g.mapHeight = len(worldMap[0])

	// Todo
	// load content once when first run
	// g.loadContent()

	// create crosshairs and weapon
	g.crosshairs = model.NewCrosshairs(txToCore)

	// init player model
	angleDegrees := 60.0
	g.player = model.NewPlayer(8.5, 3.5, geom.Radians(angleDegrees), 0)
	g.player.CollisionRadius = clipDistance
	g.player.CollisionHeight = 0.5

	// Todo
	// init the sprites
	// g.loadSprites()

	if g.osType == osTypeBrowser {
		// web browser cannot start with cursor captured
	} else {
		ebiten.SetCursorMode(ebiten.CursorModeCaptured)
	}

	// init mouse look mode
	g.mouseMode = MouseModeLook
	g.mouseX, g.mouseY = math.MinInt32, math.MinInt32

	//--init camera and renderer--//
	g.camera = raycaster.NewCamera(g.width, g.height, loader.TexWidth, g.mapObj, g.tex)
	g.setRenderDistance(g.renderDistance)

	g.camera.SetFloorTexture(loader.GetTextureFromFile("floor.png"))
	g.camera.SetSkyTexture(loader.GetTextureFromFile("sky.png"))

	// initialize camera to player position
	g.updatePlayerCamera(true)
	g.setFovAngle(g.fovDegrees)
	g.fovDepth = g.camera.FovDepth()

	g.zoomFovDepth = 2.0

	// set demo lighting settings
	g.setLightFalloff(-200)
	g.setGlobalIllumination(500)
	minLightRGB := color.NRGBA{R: 76, G: 76, B: 76}
	maxLightRGB := color.NRGBA{R: 255, G: 255, B: 255}
	g.setLightRGB(minLightRGB, maxLightRGB)

	// init menu system
	g.menu = createMenu(g)

	return g
}

func (g *Game) initConfig() {
	viper.SetConfigName("demo-config")
	viper.SetConfigType("json")

	// special behavior needed for wasm play
	switch runtime.GOOS {
	case "js":
		g.osType = osTypeBrowser
	default:
		g.osType = osTypeDesktop
	}

	// setup environment variable with DEMO as prefix (e.g. "export DEMO_SCREEN_VSYNC=false")
	viper.SetEnvPrefix("demo")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	userHomePath, _ := os.UserHomeDir()
	if userHomePath != "" {
		userHomePath = userHomePath + "/.raycaster-go-demo"
		viper.AddConfigPath(userHomePath)
	}
	viper.AddConfigPath(".")

	// set default config values
	viper.SetDefault("debug", false)
	viper.SetDefault("showSpriteBoxes", false)
	viper.SetDefault("screen.fullscreen", false)
	viper.SetDefault("screen.vsync", true)
	viper.SetDefault("screen.renderDistance", -1)
	viper.SetDefault("screen.renderFloor", true)
	viper.SetDefault("screen.fovDegrees", 68)

	if g.osType == osTypeBrowser {
		viper.SetDefault("screen.width", 800)
		viper.SetDefault("screen.height", 600)
		viper.SetDefault("screen.renderScale", 0.5)
	} else {
		viper.SetDefault("screen.width", 1024)
		viper.SetDefault("screen.height", 768)
		viper.SetDefault("screen.renderScale", 1.0)
	}

	err := viper.ReadInConfig()
	if err != nil && g.debug {
		fmt.Print(err)
	}

	// get config values
	g.screenWidth = viper.GetInt("screen.width")
	g.screenHeight = viper.GetInt("screen.height")
	g.fovDegrees = viper.GetFloat64("screen.fovDegrees")
	g.renderScale = viper.GetFloat64("screen.renderScale")
	g.fullscreen = viper.GetBool("screen.fullscreen")
	g.vsync = viper.GetBool("screen.vsync")
	g.renderDistance = viper.GetFloat64("screen.renderDistance")
	g.initRenderFloorTex = viper.GetBool("screen.renderFloor")
	g.showSpriteBoxes = viper.GetBool("showSpriteBoxes")
	g.debug = viper.GetBool("debug")
}

func (g *Game) SaveConfig() error {
	userConfigPath, _ := os.UserHomeDir()
	if userConfigPath == "" {
		userConfigPath = "./"
	}
	userConfigPath += "/.raycaster-go-demo"

	userConfig := userConfigPath + "/demo-config.json"
	fmt.Print("Saving config file ", userConfig)

	if _, err := os.Stat(userConfigPath); os.IsNotExist(err) {
		err = os.MkdirAll(userConfigPath, os.ModePerm)
		if err != nil {
			fmt.Print(err)
			return err
		}
	}
	err := viper.WriteConfigAs(userConfig)
	if err != nil {
		fmt.Print(err)
	}

	return err
}
