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

	mouseInfo MousePosition
	cfg       GameCfg
	coreTx    RcTx
}

func (r *Game) processMessage() {
	var err error
	moreMsg := true
	for moreMsg {
		select {
		case msg := <-r.rx:
			err = r.process(msg)
			if err != nil {
				fmt.Println(err)
			}
		default:
			moreMsg = false
		}
	}
}

func (r *Game) process(m ReactorEventMessage) error {
	// logger.Print(fmt.Sprintf("(%v) recv %T", r.thing.GetData().Entity.RgId, e))
	switch m.event.(type) {
	default:
		r.eventHandleUnknown(m.sender, m.event)
	}
	return nil
}

func (g *Game) eventHandleUnknown(sender RcTx, e IReactorEvent) error {
	logger.Fatal(fmt.Sprintf("Unknown event: %T", e))
	return nil
}

// Update - Allows the game to run logic such as updating the world, gathering input, and playing audio.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	// g.processMessage()
	g.handleInput(g.mouseInfo)
	if !g.paused {
		m := ReactorEventMessage{g.tx, EventGameTick{}}
		g.coreTx <- m
	}
	// update the menu (if active)
	g.menu.update()

	return nil
}

func (g *Game) Run() {
	g.paused = false
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	logger.Print("Start")
	// Debug
	//ebiten.SetTPS(1)
	if err := ebiten.RunGame(g); err != nil {
		logger.Fatal(err)
	}
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	m := ReactorEventMessage{g.tx, EventDraw{Screen: screen}}
	g.coreTx <- m
	<-g.rx

	// draw menu (if active)
	g.menu.draw(screen)

	//logger.Print("Draw reply", r)
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	w, h := int(float64(g.cfg.ScreenWidth)), int(float64(g.cfg.ScreenHeight))
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
	// default TPS is 60
	// ebiten.SetMaxTPS(60)

	rand.Seed(time.Now().UnixNano())

	// initialize Game object
	g := new(Game)
	g.initConfig()

	coreTx := NewCore(g.cfg)

	// Todo

	// create crosshairs and weapon
	NewCrosshairs(coreTx)
	NewPlayer(coreTx)
	for i := 0; i < 1; i++ {
		// NewSorcerer(coreTx)
		// NewWalker(coreTx)
		// NewBat(coreTx)
		NewRock(coreTx)
	}

	// Todo
	// init the sprites
	// g.loadSprites()

	if g.cfg.OsType == OsTypeBrowser {
		// web browser cannot start with cursor captured
	} else {
		ebiten.SetCursorMode(ebiten.CursorModeCaptured)
	}

	// init mouse look mode
	g.cfg.MouseMode = MouseModeLook
	g.mouseInfo.X, g.mouseInfo.Y = math.MinInt32, math.MinInt32

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
		g.cfg.OsType = OsTypeBrowser
	default:
		g.cfg.OsType = OsTypeDesktop
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

	if g.cfg.OsType == OsTypeBrowser {
		viper.SetDefault("screen.width", 800)
		viper.SetDefault("screen.height", 600)
		viper.SetDefault("screen.renderScale", 0.5)
	} else {
		viper.SetDefault("screen.width", 1024)
		viper.SetDefault("screen.height", 768)
		viper.SetDefault("screen.renderScale", 1.0)
	}

	err := viper.ReadInConfig()
	if err != nil && g.cfg.Debug {
		fmt.Print(err)
	}

	// get config values
	g.cfg.ScreenWidth = viper.GetInt("screen.width")
	g.cfg.Width = g.cfg.ScreenWidth
	g.cfg.ScreenHeight = viper.GetInt("screen.height")
	g.cfg.Height = g.cfg.ScreenHeight
	g.cfg.FovDegrees = viper.GetFloat64("screen.fovDegrees")
	g.cfg.RenderScale = viper.GetFloat64("screen.renderScale")
	g.cfg.Fullscreen = viper.GetBool("screen.fullscreen")
	g.cfg.Vsync = viper.GetBool("screen.vsync")
	g.cfg.RenderDistance = viper.GetFloat64("screen.renderDistance")
	g.cfg.RenderFloorTex = viper.GetBool("screen.renderFloor")
	g.cfg.ShowSpriteBoxes = viper.GetBool("showSpriteBoxes")
	g.cfg.ShowSpriteBoxes = true
	g.cfg.Debug = viper.GetBool("debug")
	g.cfg.Debug = true
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

func (g *Game) handleInput(si MousePosition) bool {
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