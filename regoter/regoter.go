package regoter

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"lintech/rego/iregoter"
	"log"
	"os"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/images"
)

var logger = log.New(os.Stderr, "Regoter ", 0)

type rgRxMsgbox <-chan iregoter.ICoreEvent
type rgTxMsgbox chan<- iregoter.IRegoterEvent

type IThing interface {
	Update() (iregoter.RgPosition, *ebiten.Image)
}

type Regoter[T IThing] struct {
	id     iregoter.ID
	rxBox  rgRxMsgbox
	txChan rgTxMsgbox
	thing  T
}

type idGenerator struct {
	id iregoter.ID
	mu sync.Mutex
}

var idg = idGenerator{id: 0}

func (g *idGenerator) genId() iregoter.ID {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.id += 1
	return g.id
}

func (g *idGenerator) currentId() iregoter.ID {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.id
}

func (r *Regoter[T]) process(e iregoter.ICoreEvent) error {
	//logger.Print(fmt.Sprintf("(%v) recv %T", r.id, e))
	switch e.(type) {
	case iregoter.CoreEventTick:
		r.eventHandleUpdate(e.(iregoter.CoreEventTick))
	default:
		r.eventHandleUnknown(e)
	}
	return nil
}

// Update the position and status of Regoter
// And send new Position and status to IGame
func (r *Regoter[T]) eventHandleUpdate(e iregoter.CoreEventTick) error {
	p, img := r.thing.Update()
	u := iregoter.RegoterEventUpdated{RgId: r.id, Position: p, Img: img}
	r.txChan <- u
	// FOR DEBUG: Test concurrent
	// if idg.currentId() < 10000 {
	// 	Create(r.txChan, r.position)
	// }
	return nil
}

func (r *Regoter[T]) eventHandleUnknown(e iregoter.ICoreEvent) error {
	logger.Fatal("Unknown event:", e)
	return nil
}

func (r *Regoter[T]) Run() {
	for {
		e := <-r.rxBox
		err := r.process(e)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func NewRegoter[T IThing](coreMsgbox chan<- iregoter.IRegoterEvent, t T) *Regoter[T] {
	id := idg.genId()
	// NOTE:  If there are about 10,000 Regoters,
	// 			  this Core routine may not able to recv msg quick enough.
	//			 	So Regoter will block on sending data to Core and can not receive data.
	// .      At same time Core are going to send data to this Regoter.
	//        It will be Dead Lock!!!
	// 			  So we set Regoter chan buffer size to 100 and keep Core buffer size at 1.
	//				So Core will not be blocked on Sending. And Regoter need wait Core.
	c := make(chan iregoter.ICoreEvent, 100)
	r := &Regoter[T]{id, c, coreMsgbox, t}
	go func() {

		p, img := r.thing.Update()
		r.txChan <- iregoter.RegoterEventNewRegoter{RgId: r.id, Msgbox: c, Position: p, Img: img}
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
