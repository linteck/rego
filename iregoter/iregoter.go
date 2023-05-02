package iregoter

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type ID int

type ImgLayer int

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

type GameEventTick struct {
}

type GameEventDraw struct {
	Screen *ebiten.Image
}

type CoreEventDrawDone struct {
}

type ICoreEvent interface {
}

// type ScreenSize struct {
// 	Width  int
// 	Height int
// }

type CoreEventUpdateTick struct {
	RgEntity   Entity
	PlayEntiry Entity
	RgState    RegoterState
}

type IRegoterEvent interface {
}

type RegoterEventNewRegoter struct {
	Msgbox chan<- ICoreEvent
	RgData RegoterData
}

type RegoterEventRegoterUnregister struct {
	RgId ID
}

// type RegoterUpdatedImg struct {
// 	ImgLayer ImgLayer
// 	ImgOp    *ebiten.DrawImageOptions
// 	Img      *ebiten.Image
// 	Sprite   *Sprite
// 	Changed  bool
// 	Visiable bool
// 	Deleted  bool
// }

type RegoterUpdatedConfig struct {
}

// type RegoterEventUpdatedImg struct {
// 	RgId ID
// 	Info RegoterUpdatedImg
// }

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

type MousePosition struct {
	// Mouse
	X int
	Y int
}
type EventDebugPrint struct {
	DebugString string
}

type GameEventCfgChanged struct {
	Cfg GameCfg
}

type GameCfg struct {
	OsType    OsType
	MouseMode MouseMode
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
	RenderFloorTex bool
	// Debug option
	ShowSpriteBoxes bool
	Debug           bool
}

type CoreRxMsgbox <-chan IRegoterEvent
type CoreTxMsgbox chan<- ICoreEvent

type RgRxMsgbox <-chan ICoreEvent
type RgTxMsgbox chan<- IRegoterEvent

type RegoterState struct {
	Unregistered bool
	HasCollision bool
	HitHarm      int
}

type DrawInfo struct {
	ImgLayer      ImgLayer
	Img           *ebiten.Image
	Columns       int
	Rows          int
	SpriteIndex   int
	AnimationRate int
	HitIndex      int //Frame index when Sprite is hit
	Illumination  float64
}

type RegoterData struct {
	RgId     ID
	RgType   RegoterEnum
	Entity   Entity
	DrawInfo DrawInfo
}
