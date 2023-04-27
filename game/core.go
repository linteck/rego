package game

import (
	"lintech/rego/iregoter"
	"log"
	"os"
	"time"

	"github.com/chen3feng/stl4go"
	"github.com/hajimehoshi/ebiten/v2"
)

var logger = log.New(os.Stdout, "Core ", 0)

type coreRxMsgbox <-chan iregoter.IRegoterEvent
type coreTxMsgbox chan<- iregoter.ICoreEvent

type rgInfo struct {
	tx       coreTxMsgbox
	position iregoter.RgPosition
	img      *ebiten.Image
}

type Core struct {
	rxBox    coreRxMsgbox
	txToGame chan<- iregoter.ICoreEvent
	rgs      *stl4go.SkipList[iregoter.ID, rgInfo]
}

func (g *Core) eventHandleGameEventTick(e iregoter.GameEventTick) {
	g.update()
}

func (g *Core) eventHandleNewRegoter(e iregoter.RegoterEventNewRegoter) {
	g.rgs.Insert(e.RgId, rgInfo{e.Msgbox, e.Position, e.Img})
}

func (g *Core) eventHandleUpdated(e iregoter.RegoterEventUpdated) {
	g.rgs.Find(e.RgId).position = e.Position
	g.rgs.Find(e.RgId).img = e.Img
}

func (g *Core) eventHandleRegoterDeleted(e iregoter.RegoterEventRegoterDeleted) {
	g.rgs.Remove(e.RgId)
}

func (r *Core) eventHandleUnknown(e iregoter.IRegoterEvent) error {
	logger.Fatal("Unknown event:", e)
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

	g.rgs.ForEach(func(k iregoter.ID, v rgInfo) {
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
	g.rgs.ForEach(func(k iregoter.ID, v rgInfo) {
		//logger.Print("Position(", v.position.X, ", ", v.position.Y, ")")
		//op.GeoM.Apply(float64(v.position.X), float64(v.position.Y))
		op.GeoM.Reset()
		op.GeoM.Translate(float64(v.position.X), float64(v.position.Y))

		//op.GeoM.Translate(screenWidth/2, screenHeight/2)
		e.Screen.DrawImage(v.img, op)
	})
	r := iregoter.CoreEventDrawDone{}
	g.txToGame <- r

}

func NewCore() (chan<- iregoter.IRegoterEvent, <-chan iregoter.ICoreEvent) {
	c := make(chan iregoter.IRegoterEvent)
	g := make(chan iregoter.ICoreEvent)
	rs := stl4go.NewSkipList[iregoter.ID, rgInfo]()
	core := &Core{c, g, rs}
	go core.Run()
	return c, g
}
