package model

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"lintech/rego/iregoter"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/images"
)

var logger = log.New(os.Stderr, "Regoter ", 0)

type IThing interface {
	GetData() iregoter.RegoterData
	UpdateTick(c iregoter.RgTxMsgbox)
	UpdateData(c iregoter.RgTxMsgbox, rgEntity iregoter.Entity,
		state iregoter.RegoterState) bool
	SetConfig(cfg iregoter.GameCfg)
}

type Regoter[T IThing] struct {
	rxBox  iregoter.RgRxMsgbox
	txChan iregoter.RgTxMsgbox
	thing  T
}

func (r *Regoter[T]) process(e iregoter.ICoreEvent) (bool, error) {
	// logger.Print(fmt.Sprintf("(%v) recv %T", r.thing.GetData().Entity.RgId, e))
	running := true
	switch e.(type) {
	case iregoter.CoreEventUpdateTick:
		r.eventHandleUpdateTick(e.(iregoter.CoreEventUpdateTick))
	case iregoter.CoreEventUpdateData:
		running = r.eventHandleUpdateData(e.(iregoter.CoreEventUpdateData))
	case iregoter.GameEventCfgChanged:
		r.eventHandleCfgChanged(e.(iregoter.GameEventCfgChanged))
	default:
		r.eventHandleUnknown(e)
	}
	return running, nil
}

// Update the position and status of Regoter
// And send new Position and status to IGame
func (r *Regoter[T]) eventHandleUpdateData(e iregoter.CoreEventUpdateData) bool {
	return r.thing.UpdateData(r.txChan, e.RgEntity, e.RgState)
}

func (r *Regoter[T]) eventHandleUpdateTick(e iregoter.CoreEventUpdateTick) error {
	r.thing.UpdateTick(r.txChan)
	return nil
}

func (r *Regoter[T]) eventHandleCfgChanged(e iregoter.GameEventCfgChanged) error {
	r.thing.SetConfig(e.Cfg)
	return nil
}

func (r *Regoter[T]) eventHandleUnknown(e iregoter.ICoreEvent) error {
	logger.Fatal("Unknown event:", e)
	return nil
}

func (r *Regoter[T]) Run() {
	running := true
	var err error
	for running {
		e := <-r.rxBox
		running, err = r.process(e)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func NewRegoter[T IThing](coreMsgbox chan<- iregoter.IRegoterEvent, t T) *Regoter[T] {
	// NOTE:  If there are about 10,000 Regoters,
	// 			  this Core routine may not able to recv msg quick enough.
	//			 	So Regoter will block on sending data to Core and can not receive data.
	// .      At same time Core are going to send data to this Regoter.
	//        It will be Dead Lock!!!
	// 			  So we set Regoter chan buffer size to 100 and keep Core buffer size at 1.
	//				So Core will not be blocked on Sending. And Regoter need wait Core.
	c := make(chan iregoter.ICoreEvent, 10)
	r := &Regoter[T]{c, coreMsgbox, t}
	go func() {
		d := t.GetData()
		e := iregoter.RegoterEventRegisterRegoter{Msgbox: c, RgData: d}
		r.txChan <- e
		r.Run()
	}()
	return r
}

func LoadImage() *ebiten.Image {
	// Decode an image from the image file's byte slice.
	img, _, err := image.Decode(bytes.NewReader(images.Runner_png))
	if err != nil {
		logger.Fatal("Load Image: ", err)
	}
	return ebiten.NewImageFromImage(img)
}
