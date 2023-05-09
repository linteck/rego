package model

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/spf13/viper"
)

// type gameTxMsgbox chan<- IRegoterEvent
// type gameRxMsgbox <-chan ICoreEvent

type Game struct {
	Reactor
	// Raycaster
	menu   *DemoMenu
	paused bool

	cfg         GameCfg
	coreTx      RcTx
	audioPlayer *RegoAudioPlayer
}

func createSpritesFunc() func(coreTx RcTx) {
	const max_gen_sprites = 1
	var gened_sprites = 0

	return func(coreTx RcTx) {
		if gened_sprites < max_gen_sprites {
			r := rand.Intn(10)
			if r == 1 {
				NewSorcerer(coreTx)
				NewWalker(coreTx)
				NewBat(coreTx)
				NewRock(coreTx)
				gened_sprites++
			}
		}
	}
}

var createSprites = createSpritesFunc()

// Update - Allows the game to run logic such as updating the world, gathering input, and playing audio.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	if !g.paused {
		m := ReactorEventMessage{g.tx, EventGameTick{}}
		g.coreTx <- m
	}
	// If we add Same Sprites in CreateGame(), they will show same frame of Animation at each Tick.
	// Becasue they have same count of Update Ticks.
	// So we need create Sprite inside Game.Update in different Ticks.
	createSprites(g.coreTx)
	g.handleInput()
	// update the menu (if active)
	g.menu.update()

	return nil
}

func (g *Game) Run() {
	if g.rx == nil || g.tx == nil {
		log.Fatal("Reactor channel is not initialized!")
	}
	g.paused = false
	log.Print("Start")
	// Debug
	//ebiten.SetTPS(1)
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	m := ReactorEventMessage{g.tx, EventDraw{Screen: screen}}
	g.coreTx <- m
	//While Core is drawing, we play background music
	g.playBackGroundAudio()
	<-g.rx

	// draw menu (if active)
	g.menu.draw(screen)

	//log.Print("Draw reply", r)
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	w, h := int(float64(g.cfg.ScreenWidth)), int(float64(g.cfg.ScreenHeight))
	g.menu.layout(w, h)
	return int(w), int(h)
}

func NewGame(coreTx RcTx, cfg GameCfg) *Game {
	//loadCrosshairsResource()
	t := &Game{
		Reactor:     NewReactor(),
		cfg:         cfg,
		coreTx:      coreTx,
		audioPlayer: LoadAudioPlayer("dark-castle-night.mp3"),
	}
	t.menu = t.createMenu()
	return t
}

// CreateGame - Allows the game to perform any initialization it needs to before starting to run.
// This is where it can query for any required services and load any non-graphic
// related content.  Calling base.Initialize will enumerate through any components
// and initialize them as well.
func CreateGame() *Game {
	fmt.Printf("Initializing Game\n")
	ebiten.SetWindowTitle("Rego Demo")
	// default TPS is 60
	// ebiten.SetMaxTPS(60)

	rand.Seed(time.Now().UnixNano())

	// initialize Game object
	cfg := initConfig()
	coreTx := NewCore(cfg)
	g := NewGame(coreTx, cfg)

	// create crosshairs and weapon
	NewCrosshairs(coreTx)
	NewPlayer(coreTx)

	// Todo
	// init the sprites
	// g.loadSprites()

	// init mouse look mode

	// init menu system

	return g
}

func (g *Game) playBackGroundAudio() {
	g.audioPlayer.PlayWithVolume(0.5, false)
}

func initConfig() GameCfg {
	viper.SetConfigName("demo-config")
	viper.SetConfigType("json")

	// special behavior needed for wasm play
	cfg := GameCfg{}
	switch runtime.GOOS {
	case "js":
		cfg.OsType = OsTypeBrowser
	default:
		cfg.OsType = OsTypeDesktop
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
	viper.SetDefault("screen.renderAudioDistance", 50)
	viper.SetDefault("screen.renderFloor", true)
	viper.SetDefault("screen.fovDegrees", 68)

	if cfg.OsType == OsTypeBrowser {
		viper.SetDefault("screen.width", 800)
		viper.SetDefault("screen.height", 600)
		viper.SetDefault("screen.renderScale", 0.5)
	} else {
		viper.SetDefault("screen.width", 1024)
		viper.SetDefault("screen.height", 768)
		viper.SetDefault("screen.renderScale", 1.0)
	}

	err := viper.ReadInConfig()
	if err != nil && cfg.Debug {
		fmt.Print(err)
	}

	// get config values
	cfg.ScreenWidth = viper.GetInt("screen.width")
	cfg.Width = cfg.ScreenWidth
	cfg.ScreenHeight = viper.GetInt("screen.height")
	cfg.Height = cfg.ScreenHeight
	cfg.FovDegrees = viper.GetFloat64("screen.fovDegrees")
	cfg.RenderScale = viper.GetFloat64("screen.renderScale")
	cfg.Fullscreen = viper.GetBool("screen.fullscreen")
	cfg.Vsync = viper.GetBool("screen.vsync")
	cfg.RenderDistance = viper.GetFloat64("screen.renderDistance")
	cfg.RenderAudioDistance = viper.GetFloat64("screen.renderAudioDistance")
	cfg.RenderFloorTex = viper.GetBool("screen.renderFloor")
	cfg.ShowSpriteBoxes = viper.GetBool("showSpriteBoxes")
	// cfg.ShowSpriteBoxes = true
	cfg.Debug = viper.GetBool("debug")
	//cfg.Debug = true
	if cfg.OsType == OsTypeBrowser {
		// web browser cannot start with cursor captured
	} else {
		ebiten.SetCursorMode(ebiten.CursorModeCaptured)
	}
	cfg.MouseMode = MouseModeLook
	return cfg
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

var mouse = MousePosition{math.MinInt32, math.MinInt32}

func (g *Game) handleInput() bool {
	menuKeyPressed := inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyF1)
	if menuKeyPressed {
		if g.menu.active {
			if g.cfg.OsType == OsTypeBrowser && inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
				// do not allow Esc key close menu in browser, since Esc key releases browser mouse capture
			} else {
				g.closeMenu()
				m := ReactorEventMessage{g.tx, EventCfgChanged{Cfg: g.cfg}}
				g.coreTx <- m
			}
		} else {
			g.openMenu()
		}
	}

	if g.cfg.OsType == OsTypeBrowser && ebiten.CursorMode() == ebiten.CursorModeVisible && !g.menu.active {
		// not working sometimes (https://developer.mozilla.org/en-US/docs/Web/API/Pointer_Lock_API#iframe_limitations):
		//   sm_exec.js:349 pointerlockerror event is fired. 'sandbox="allow-pointer-lock"' might be required at an iframe.
		//   This function on browsers must be called as a result of a gestural interaction or orientation change.
		//   localhost/:1 Uncaught (in promise) DOMException: The user has exited the lock before this request was completed.
		g.openMenu()
	}

	if g.paused {
		// currently only paused when menu is active, one could consider other pauses not the subject of this demo
		return g.menu.active
	}

	return g.menu.active

}
