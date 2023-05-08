package model

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

type IReactorEvent interface{}
type RcRx <-chan ReactorEventMessage
type RcTx chan<- ReactorEventMessage

type IEventHandler interface {
	handleEvent(RcTx, IReactorEvent)
}

type Reactor struct {
	rx      RcRx
	tx      RcTx
	running bool
}

type ReactorEventMessage struct {
	sender RcTx
	event  IReactorEvent
}

type ID int

type ImgLayer int

const (
	ImgLayerSprite ImgLayer = iota
	ImgLayerPlayer
)

type RegoterEnum int

const (
	RegoterEnumSprite RegoterEnum = iota
	RegoterEnumProjectile
	RegoterEnumEffect
	RegoterEnumCrosshair
	RegoterEnumWeapon
	RegoterEnumPlayer
)

type ICoreEvent interface {
}

//	type ScreenSize struct {
//		Width  int
//		Height int
//	}

type IRegoterEvent interface {
}

type RegoterUpdatedConfig struct {
}

// type RegoterEventUpdatedImg struct {
// 	RgId ID
// 	Info RegoterUpdatedImg
// }

type Movement struct {
	MoveRotate    float64
	PitchRotate   float64
	Acceleration  float64
	Velocity      float64
	VissionRotate float64
}

type Command struct {
	StartAnimation bool
	StopAnimation  bool
}

type Action struct {
	FireWeapon bool
	nextWeapon bool
	KeyPressed bool
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
	ZoomFovDepth        float64
	RenderDistance      float64
	RenderAudioDistance float64
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

// type CoreRxMsgbox <-chan IRegoterEvent
// type CoreTxMsgbox chan<- ICoreEvent

// type RgRxMsgbox <-chan ICoreEvent
// type RgTxMsgbox chan<- IRegoterEvent

type RegoterState struct {
	AnimationLoopCnt      int
	IsAnimationFirstFrame bool
	AnimationRunning      bool
}

type DrawInfo struct {
	ImgLayer          ImgLayer
	Img               *ebiten.Image
	TexFacingMap      *map[float64]int
	AnimationReversed bool
	Columns           int
	Rows              int
	SpriteIndex       int
	AnimationRate     int
	HitIndex          int //Frame index when Sprite is hit
	Illumination      float64
}

type RegoterData struct {
	Entity   Entity
	DrawInfo DrawInfo
}

type CollisionSpace struct {
	CollisionRadius float64
	CollisionHeight float64
}

type EntityCollision struct {
	// entity     *Entity
	position Position
	peer     ID
	distance float64
}

type EventCollision struct {
	collistion EntityCollision
}

// Events
type EventDebugPrint struct {
	DebugString string
}

type EventDamagePeer struct {
	peer   ID
	damage int
}

type EventHealthChange struct {
	change int
}

type EventCfgChanged struct {
	Cfg GameCfg
}

type EventDraw struct {
	Screen *ebiten.Image
}

type EventDrawDone struct {
}

type EventHolsterWeapon struct {
}

type EventFireWeapon struct {
}

type EventUpdateTick struct {
	RgEntity     Entity
	RgState      RegoterState
	PlayerEntity Entity
}

type EventGameTick struct{}

type EventRegisterRegoter struct {
	tx     RcTx
	RgData RegoterData
}

type EventUnregisterRegoter struct {
	RgId ID
}

type EventUnregisterConfirmed struct {
}

type EventMovement struct {
	RgId    ID
	Move    Movement
	Command Command
}

// type EventInput struct {
// 	input Movement
// }

func NewReactor() Reactor {
	// NOTE:  If there are about 10,000 Regoters,
	// 			  this Core routine may not able to recv msg quick enough.
	//			 	So Regoter will block on sending data to Core and can not receive data.
	// .      At same time Core are going to send data to this Regoter.
	//        It will be Dead Lock!!!
	// 			  So we set Regoter chan buffer size to 100 and keep Core buffer size at 1.
	//				So Core will not be blocked on Sending. And Regoter need wait Core.
	c := make(chan ReactorEventMessage, 100)
	rc := Reactor{c, c, true}
	return rc
}

type IProcessMessage interface {
	ProcessMessage(m ReactorEventMessage) error
}

func (r *Reactor) Run(t IProcessMessage) {
	if r.rx == nil || r.tx == nil {
		log.Fatal("Reactor channel is not initialized!")
	}
	r.running = true
	var err error
	for r.running {
		msg := <-r.rx
		err = t.ProcessMessage(msg)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func NewReactorCore() Reactor {
	// NOTE:  If there are about 10,000 Regoters,
	// 			  this Core routine may not able to recv msg quick enough.
	//			 	So Regoter will block on sending data to Core and can not receive data.
	// .      At same time Core are going to send data to this Regoter.
	//        It will be Dead Lock!!!
	// 			  So we set Regoter chan buffer size to 100 and keep Core buffer size at 1.
	//				So Core will not be blocked on Sending. And Regoter need wait Core.
	c := make(chan ReactorEventMessage)
	rc := Reactor{c, c, true}
	return rc
}
