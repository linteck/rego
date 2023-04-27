package iregoter

import "github.com/hajimehoshi/ebiten/v2"

type ID int
type RgPosition struct {
	X int
	Y int
}

type ICore interface {
}

type GameEventTick struct {
}

type GameEventDraw struct {
	Screen *ebiten.Image
}

type CoreEventDrawDone struct {
}

type ICoreEvent interface {
}

type CoreEventTick struct {
}

type IRegoterEvent interface {
}

type RegoterEventNewRegoter struct {
	Msgbox   chan<- ICoreEvent
	RgId     ID
	Position RgPosition
	Img      *ebiten.Image
}

type RegoterEventRegoterDeleted struct {
	RgId ID
}

type RegoterEventUpdated struct {
	RgId     ID
	Position RgPosition
	Img      *ebiten.Image
}
