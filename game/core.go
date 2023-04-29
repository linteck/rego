package game

import (
	"fmt"
	"lintech/rego/iregoter"
	"log"
	"os"
	"time"

	"github.com/chen3feng/stl4go"
)

var logger = log.New(os.Stdout, "Core ", 0)

type coreRxMsgbox <-chan iregoter.IRegoterEvent
type coreTxMsgbox chan<- iregoter.ICoreEvent

type regoterInCore struct {
	tx coreTxMsgbox
}

var imgLayerPriorities = [...]iregoter.ImgLayer{
	iregoter.ImgLayerSprite,
	iregoter.ImgLayerSpriteHint,
	iregoter.ImgLayerProjectile}

type Core struct {
	rxBox    coreRxMsgbox
	txToGame chan<- iregoter.ICoreEvent
	rgs      *stl4go.SkipList[iregoter.ID, regoterInCore]
	imgs     [len(imgLayerPriorities)]*stl4go.DList[iregoter.RegoterUpdatedInfo]
}

func (g *Core) eventHandleGameEventTick(e iregoter.GameEventTick) {
	g.update()
}

func (g *Core) eventHandleNewRegoter(e iregoter.RegoterEventNewRegoter) {
	g.rgs.Insert(e.RgId, regoterInCore{e.Msgbox})
}

func (g *Core) eventHandleUpdated(e iregoter.RegoterEventUpdated) {
	l := g.imgs[e.Info.ImgLayer]
	l.PushBack(e.Info)
}

func (g *Core) eventHandleRegoterDeleted(e iregoter.RegoterEventRegoterDeleted) {
	g.rgs.Remove(e.RgId)
}

func (r *Core) eventHandleUnknown(e iregoter.IRegoterEvent) error {
	logger.Fatal(fmt.Sprintf("Unknown event: %T", e))
	return nil
}

func (g *Core) eventHandleGameEventDraw(e iregoter.GameEventDraw) {

	for p := range imgLayerPriorities {
		g.imgs[p].ForEach(func(v iregoter.RegoterUpdatedInfo) {
			e.Screen.DrawImage(v.Img, v.ImgOp)
		})
		g.imgs[p].Clear()
	}
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

func NewCore() (chan<- iregoter.IRegoterEvent, <-chan iregoter.ICoreEvent) {
	c := make(chan iregoter.IRegoterEvent)
	g := make(chan iregoter.ICoreEvent)
	rs := stl4go.NewSkipList[iregoter.ID, regoterInCore]()
	var imgs [len(imgLayerPriorities)]*stl4go.DList[iregoter.RegoterUpdatedInfo]
	for i := 0; i < len(imgs); i++ {
		imgs[i] = stl4go.NewDList[iregoter.RegoterUpdatedInfo]()
	}
	core := &Core{c, g, rs, imgs}
	go core.Run()
	return c, g
}
