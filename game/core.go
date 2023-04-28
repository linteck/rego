package game

import (
	"fmt"
	"lintech/rego/iregoter"
	"log"
	"os"
	"time"

	"github.com/chen3feng/stl4go"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var logger = log.New(os.Stdout, "Core ", 0)

type coreRxMsgbox <-chan iregoter.IRegoterEvent
type coreTxMsgbox chan<- iregoter.ICoreEvent

type regoterInCore struct {
	tx          coreTxMsgbox
	updatedInfo iregoter.RegoterUpdatedInfo
}

type Core struct {
	rxBox    coreRxMsgbox
	txToGame chan<- iregoter.ICoreEvent
	rgs      *stl4go.SkipList[iregoter.ID, regoterInCore]
}

func (g *Core) eventHandleGameEventTick(e iregoter.GameEventTick) {
	g.update()
}

func (g *Core) eventHandleNewRegoter(e iregoter.RegoterEventNewRegoter) {
	g.rgs.Insert(e.RgId, regoterInCore{e.Msgbox, e.Info})
}

func (g *Core) eventHandleUpdated(e iregoter.RegoterEventUpdated) {
	g.rgs.Find(e.RgId).updatedInfo = e.Info
}

func (g *Core) eventHandleRegoterDeleted(e iregoter.RegoterEventRegoterDeleted) {
	g.rgs.Remove(e.RgId)
}

func (r *Core) eventHandleUnknown(e iregoter.IRegoterEvent) error {
	logger.Fatal(fmt.Sprintf("Unknown event: %T", e))
	return nil
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
	case iregoter.RegoterEventUpdated:
		g.eventHandleUpdated(e.(iregoter.RegoterEventUpdated))
	default:
		g.eventHandleUnknown(e)
	}
	return nil
}

func (g *Core) update() error {
	e := iregoter.CoreEventTick{}

	g.rgs.ForEach(func(k iregoter.ID, v regoterInCore) {
		v.tx <- e
	})
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

func (g *Core) eventHandleGameEventDraw(e iregoter.GameEventDraw) {

	op := &ebiten.DrawImageOptions{}
	// op.GeoM.Translate(-float64(frameWidth)/2, -float64(frameHeight)/2)
	// op.GeoM.Translate(screenWidth/2, screenHeight/2)
	g.rgs.ForEach(func(k iregoter.ID, v regoterInCore) {
		//logger.Print("Position(", v.position.X, ", ", v.position.Y, ")")
		//op.GeoM.Apply(float64(v.position.X), float64(v.position.Y))
		op.GeoM.Reset()
		p := v.updatedInfo.Position
		op.GeoM.Translate(float64(p.X), float64(p.Y))

		//op.GeoM.Translate(screenWidth/2, screenHeight/2)
		e.Screen.DrawImage(v.updatedInfo.Img, op)
	})
	// draw FPS/TPS counter debug display
	fps := fmt.Sprintf("FPS: %f\nTPS: %f/%v", ebiten.ActualFPS(), ebiten.ActualTPS(), ebiten.TPS())
	ebitenutil.DebugPrint(e.Screen, fps)

	r := iregoter.CoreEventDrawDone{}
	g.txToGame <- r

}

func NewCore() (chan<- iregoter.IRegoterEvent, <-chan iregoter.ICoreEvent) {
	c := make(chan iregoter.IRegoterEvent)
	g := make(chan iregoter.ICoreEvent)
	rs := stl4go.NewSkipList[iregoter.ID, regoterInCore]()
	core := &Core{c, g, rs}
	go core.Run()
	return c, g
}
