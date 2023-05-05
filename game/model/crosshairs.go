package model

import (
	"fmt"
	"image/color"
	"lintech/rego/game/loader"
	"log"

	"github.com/harbdog/raycaster-go"
)

type Crosshairs struct {
	Reactor
	rgData RegoterData
}

func (r *Crosshairs) Run() {
	r.running = true
	var err error
	for r.running {
		msg := <-r.rx
		err = r.process(msg)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (r *Crosshairs) process(m ReactorEventMessage) error {
	// log.Print(fmt.Sprintf("(%v) recv %T", r.thing.GetData().Entity.RgId, e))
	switch m.event.(type) {
	case EventUpdateTick:
		r.eventHandleUpdateTick(m.sender, m.event.(EventUpdateTick))
	case EventUpdateData:
		r.eventHandleUpdateData(m.sender, m.event.(EventUpdateData))
	default:
		r.eventHandleUnknown(m.sender, m.event)
	}
	return nil
}

func (r *Crosshairs) eventHandleUnknown(sender RcTx, e IReactorEvent) error {
	log.Fatal("Unknown event:", e)
	return nil
}

func NewCrosshairs(coreTx RcTx) RcTx {
	//loadCrosshairsResource()
	entity := Entity{
		RgId:            RgIdGenerator.GenId(),
		RgType:          RegoterEnumCrosshair,
		Position:        Position{X: 5, Y: 5, Z: 0},
		Scale:           2,
		MapColor:        color.RGBA{255, 0, 0, 255},
		Anchor:          raycaster.AnchorCenter,
		CollisionRadius: 0,
		CollisionHeight: 0,
	}
	di := DrawInfo{
		ImgLayer:    ImgLayerSprite,
		Img:         loader.GetSpriteFromFile("crosshairs_sheet.png"),
		Columns:     8,
		Rows:        8,
		SpriteIndex: 55,
		HitIndex:    57,
	}
	t := &Crosshairs{
		rgData: RegoterData{
			Entity:   entity,
			DrawInfo: di,
		},
	}

	go t.Run()
	m := ReactorEventMessage{t.tx, EventRegisterRegoter{t.tx, t.rgData}}
	coreTx <- m
	return t.tx
}

// func (c *Crosshairs) ActivateHitIndicator(hitTime int) {
// 	if c.HitIndicator != nil {
// 		c.hitTimer = hitTime
// 	}
// }

// func (c *Crosshairs) IsHitIndicatorActive() bool {
// 	return c.HitIndicator != nil && c.hitTimer > 0
// }

func (c *Crosshairs) eventHandleUpdateTick(sender RcTx, e EventUpdateTick) {
}

func (c *Crosshairs) eventHandleUpdateData(sender RcTx, e EventUpdateData) {
}

func (c *Crosshairs) SetConfig(cfg GameCfg) {
}
