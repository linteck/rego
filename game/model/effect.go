package model

import (
	"image/color"
	"lintech/rego/game/loader"
	"log"

	"github.com/harbdog/raycaster-go"
)

type Effect struct {
	Reactor
	EffectTemplate
	cfg          GameCfg
	unregistered bool
}

type EffectTemplate struct {
	rgData    RegoterData
	LoopCount int
}

func (r *Effect) ProcessMessage(m ReactorEventMessage) error {
	// log.Print(fmt.Sprintf("(%v) recv %T", r.thing.GetData().Entity.RgId, e))
	switch m.event.(type) {
	case EventUnregisterConfirmed:
		r.eventHandleUnregisterConfirmed(m.sender, m.event.(EventUnregisterConfirmed))
	default:
		if !r.unregistered {
			switch m.event.(type) {
			case EventUpdateTick:
				r.eventHandleUpdateTick(m.sender, m.event.(EventUpdateTick))
			case EventUnregisterConfirmed:
				r.eventHandleUnregisterConfirmed(m.sender, m.event.(EventUnregisterConfirmed))
			case EventCfgChanged:
				r.eventHandleCfgChanged(m.sender, m.event.(EventCfgChanged))
			default:
				r.eventHandleUnknown(m.sender, m.event)
			}
		}
	}
	return nil
}

func (r *Effect) eventHandleUnknown(sender RcTx, e IReactorEvent) error {
	log.Fatalf("Unknown event: %T", e)
	return nil
}

func NewEffectTemplate(di DrawInfo, scale float64, loopCount int) *EffectTemplate {
	//loadCrosshairsResource()
	entity := Entity{
		RgId:            <-IdGen,
		RgType:          RegoterEnumSprite,
		RgName:          "Effect",
		Scale:           scale,
		Velocity:        0,
		MapColor:        color.RGBA{0, 0, 0, 0},
		Anchor:          raycaster.AnchorCenter,
		CollisionRadius: 0,
		CollisionHeight: 0,
	}
	t := &EffectTemplate{
		rgData: RegoterData{
			Entity:   entity,
			DrawInfo: di,
		},
		LoopCount: loopCount,
	}

	return t
}

func NewRedExplosionEffect() *EffectTemplate {
	di := DrawInfo{
		Img:           loader.GetSpriteFromFile("red_explosion_sheet.png"),
		AnimationRate: 1,
		Columns:       8,
		Rows:          3,
	}
	redExplosionEffect := NewEffectTemplate(di, 0.20, 1)
	return redExplosionEffect
}

func NewBlueExplosionEffect() *EffectTemplate {
	di := DrawInfo{
		Img:           loader.GetSpriteFromFile("blue_explosion_sheet.png"),
		AnimationRate: 3,
		Columns:       5,
		Rows:          3,
	}
	blueExplosionEffect := NewEffectTemplate(di, 0.75, 1)
	return blueExplosionEffect
}

func (ef *Effect) eventHandleUpdateTick(sender RcTx, e EventUpdateTick) {
	if e.RgState.AnimationLoopCnt >= ef.LoopCount {
		m := ReactorEventMessage{ef.tx, EventUnregisterRegoter{RgId: ef.rgData.Entity.RgId}}
		sender <- m
		ef.unregistered = true
	}
}

func (ef *Effect) eventHandleUnregisterConfirmed(sender RcTx, e EventUnregisterConfirmed) {
	ef.running = false
}

func NewEffect(coreTx RcTx, et *EffectTemplate, position Position) RcTx {
	ef := &Effect{
		Reactor:        NewReactor(),
		EffectTemplate: *et,
	}
	// Don't use ID of Template
	ef.rgData.Entity.RgId = <-IdGen
	ef.rgData.Entity.Position = position
	//
	go ef.Reactor.Run(ef)
	m := ReactorEventMessage{ef.tx, EventRegisterRegoter{ef.tx, ef.rgData}}
	coreTx <- m
	return ef.tx
}

func (ef *EffectTemplate) Spawn(coreTx RcTx, position Position) {
	NewEffect(coreTx, ef, position)
}

func (r *Effect) eventHandleCfgChanged(sender RcTx, e EventCfgChanged) {
	r.cfg = e.Cfg
}
