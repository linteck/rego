package iregoter

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type ID int

type ImgLayer int

type ChanRegoterUpdate chan<- RegoterUpdatedImg

const (
	ImgLayerSprite ImgLayer = iota
	ImgLayerPlayer
)

type RegoterEnum int

const (
	RegoterEnumPlayer RegoterEnum = iota
	RegoterEnumSprite
	RegoterEnumProjectile
	RegoterEnumEffect
)

type RgPosition struct {
	X int
	Y int
}

type ICore interface {
}

type GameEventTick struct {
	ScreenSize ScreenSize
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

type CoreEventUpdateTick struct {
	RgEntity     *Entity
	PlayEntiry   *Entity
	HasCollision bool
	ScreenSize   ScreenSize
}

type IRegoterEvent interface {
}

type RegoterEventNewRegoter struct {
	Msgbox chan<- ICoreEvent
	RgId   ID
	RgType RegoterEnum
	Entity *Entity
}

type RegoterEventRegoterDeleted struct {
	RgId ID
}

type RegoterUpdatedImg struct {
	ImgLayer ImgLayer
	ImgOp    *ebiten.DrawImageOptions
	Img      *ebiten.Image
	Sprite   *Sprite
	Changed  bool
	Visiable bool
	Deleted  bool
}

type RegoterUpdatedPlayerImg struct {
	ImgOp *ebiten.DrawImageOptions
	Img   *ebiten.Image
}

type RegoterUpdatedMenuImg struct {
}

type RegoterEventUpdatedImg struct {
	RgId ID
	Info RegoterUpdatedImg
}

type RegoterEventUpdatedPlayerImg struct {
	Info RegoterUpdatedPlayerImg
}

type RegoterEventUpdatedMenuImg struct {
}

type RegoterUpdatedMove struct {
}

type RotateAngle float64
type Distance float64
type PitchAngle float64

type RegoterMove struct {
	RotateSpeed RotateAngle
	MoveSpeed   Distance
	PitchSpeed  PitchAngle
}

type RegoterEventUpdatedMove struct {
	RgId ID
	Move RegoterMove
}

type OsType int

const (
	OsTypeDesktop OsType = iota
	OsTypeBrowser
)

type MouseMode int

const (
	MouseModeLook MouseMode = iota
	MouseModeMove
	MouseModeCursor
)

type MouseInfo struct {
	MouseMode MouseMode
	// Mouse
	MouseX int
	MouseY int
}

type GameCfg struct {
	OsType OsType
	// window resolution and scaling
	ScreenWidth  int
	ScreenHeight int
	//--viewport width / height--//
	Width  int
	Height int

	RenderScale float64
	Fullscreen  bool
	Vsync       bool
	FovDegrees  float64
	FovDepth    float64
	// zoom settings
	ZoomFovDepth   float64
	RenderDistance float64
	// lighting settings
	LightFalloff       float64
	GlobalIllumination float64
	MinLightRGB        color.NRGBA
	MaxLightRGB        color.NRGBA
	//
	InitRenderFloorTex bool
	// Debug option
	ShowSpriteBoxes bool
	Debug           bool
}
