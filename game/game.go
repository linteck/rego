package game

import (
	"lintech/rego/iregoter"
	"lintech/rego/regoter"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 320
	screenHeight = 240
)

type gameTxMsgbox chan<- iregoter.IRegoterEvent
type gameRxMsgbox <-chan iregoter.ICoreEvent

type Game struct {
	txToCore   gameTxMsgbox
	rxFromCore gameRxMsgbox
}

func (g *Game) Update() error {
	e := iregoter.GameEventTick{}
	g.txToCore <- e
	//time.Sleep(1 * time.Second)
	return nil
}

func (g *Game) Run() {
	logger.Print("Start")
	if err := ebiten.RunGame(g); err != nil {
		logger.Fatal(err)
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	e := iregoter.GameEventDraw{Screen: screen}
	g.txToCore <- e
	<-g.rxFromCore
	//logger.Print("Draw reply", r)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func NewGame() {
	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("Animation (Ebitengine Demo)")

	txToCore, rxFromCore := NewCore()
	for i := 0; i < 10; i++ {
		regoter.NewSpiteWalker(txToCore)
	}

	g := &Game{txToCore, rxFromCore}
	g.Run()
}
