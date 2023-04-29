package iregoter

import "github.com/hajimehoshi/ebiten/v2"

type ID int

type ImgLayer int

type ChanRegoterUpdate chan<- RegoterUpdatedInfo

const (
	ImgLayerSprite ImgLayer = iota
	ImgLayerSpriteHint
	ImgLayerProjectile
)

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

type ScreenSize struct {
	Width  int
	Height int
}

type Vision struct {
	ScreenSize ScreenSize
}

type CoreEventTick struct {
	Vision Vision
}

type IRegoterEvent interface {
}

type RegoterEventNewRegoter struct {
	Msgbox chan<- ICoreEvent
	RgId   ID
}

type RegoterEventRegoterDeleted struct {
	RgId ID
}

type RegoterUpdatedInfo struct {
	ImgLayer ImgLayer
	ImgOp    *ebiten.DrawImageOptions
	Img      *ebiten.Image
	Changed  bool
	Visiable bool
	Deleted  bool
}
type RegoterEventUpdated struct {
	RgId ID
	Info RegoterUpdatedInfo
}
