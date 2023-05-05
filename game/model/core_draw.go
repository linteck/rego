package model

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/harbdog/raycaster-go"
)

func (g *Core) eventHandleGameEventDraw(sender RcTx, e EventDraw) {

	g.drawScreen(e.Screen)

	// Debug
	g.removeAllUnregisteredRogeter()
	m := ReactorEventMessage{g.tx, EventDrawDone{}}
	sender <- m

}

func (g *Core) drawScreen(screen *ebiten.Image) {
	// Put projectiles together with sprites for raycasting both as sprites
	typesNeedRaycast := []RegoterEnum{
		RegoterEnumSprite,
		RegoterEnumProjectile,
		RegoterEnumEffect,
	}
	raycastSpritesLen := 0
	for _, t := range typesNeedRaycast {
		sl := g.rgs[t]
		raycastSpritesLen += sl.Len()
	}

	raycastSprites := make([]raycaster.Sprite, 0, raycastSpritesLen)
	for _, t := range typesNeedRaycast {
		sl := g.rgs[t]
		sl.ForEach(func(i ID, val *regoterInCore) {
			if val.sprite != nil {
				raycastSprites = append(raycastSprites, val.sprite)
			}
		})
	}

	// Update camera (calculate raycast)
	g.camera.Update(raycastSprites)

	// Render raycast scene
	g.camera.Draw(g.scene)

	// draw equipped weapon
	g.drawWeapon(g.scene)
	// apply lighting setting

	g.drawSpriteBoxes(g.scene)
	// Todo
	// // draw sprite screen indicator only for sprite at point of convergence
	// convergenceSprite := g.camera.GetConvergenceSprite()
	// if convergenceSprite != nil {
	// 	for sprite := range g.sprites {
	// 		if convergenceSprite == sprite {
	// 			drawSpriteIndicator(g.scene, sprite)
	// 			break
	// 		}
	// 	}
	// }

	// draw raycasted scene
	op := &ebiten.DrawImageOptions{}
	if g.cfg.RenderScale < 1 {
		op.Filter = ebiten.FilterNearest
		op.GeoM.Scale(1/g.cfg.RenderScale, 1/g.cfg.RenderScale)
	}
	screen.DrawImage(g.scene, op)

	// draw minimap
	g.drawMiniMap(screen)

	// draw crosshairs
	g.drawCrosshairs(screen)

	// draw DebugInfo
	g.drawDebugInfo(screen)

}

func (g *Core) drawSpriteBoxes(scene *ebiten.Image) {
	if g.cfg.ShowSpriteBoxes {
		typesNeedDrawbox := []RegoterEnum{
			RegoterEnumSprite,
			RegoterEnumProjectile,
			RegoterEnumEffect,
		}
		// draw sprite screen indicators to show we know where it was raycasted (must occur after camera.Update)
		for _, t := range typesNeedDrawbox {
			sl := g.rgs[t]
			sl.ForEach(func(i ID, val *regoterInCore) {
				if val.sprite != nil {
					drawSpriteBox(g.scene, val.sprite)
				}
			})
		}
	}

}
func (g *Core) drawWeapon(scene *ebiten.Image) {
	sl := g.rgs[RegoterEnumWeapon]
	sl.ForEach(func(i ID, val *regoterInCore) {
		if val.sprite != nil {
			op := &ebiten.DrawImageOptions{}
			op.Filter = ebiten.FilterNearest
			weaponScale := val.sprite.Scale() * g.cfg.RenderScale
			op.GeoM.Scale(weaponScale, weaponScale)
			op.GeoM.Translate(
				float64(g.cfg.Width)/2-float64(val.sprite.W)*weaponScale/2,
				float64(g.cfg.Height)-float64(val.sprite.H)*weaponScale+1,
			)
			op.ColorScale.Scale(float32(g.cfg.MaxLightRGB.R)/255, float32(g.cfg.MaxLightRGB.G)/255,
				float32(g.cfg.MaxLightRGB.B)/255, 1)
			scene.DrawImage(val.sprite.Texture(), op)
		}
	})
}

func (g *Core) drawDebugInfo(screen *ebiten.Image) {
	// draw FPS/TPS counter debug display
	dbgMsg := fmt.Sprintf("FPS: %.1f\nTPS: %.1f/%v\n", ebiten.ActualFPS(), ebiten.ActualTPS(), ebiten.TPS())
	cp := g.camera.GetPosition()
	dbgMsg += fmt.Sprintf("Camera: {X:%.1f, Y:%.1f, Z: %1f", cp.X, cp.Y, g.camera.GetPositionZ())
	//ebitenutil.DebugPrint(screen, fps)
	g.debugMessages.ForEach(func(val string) {
		dbgMsg += (val + "\n")
	})
	g.debugMessages.Clear()
	ebitenutil.DebugPrint(screen, dbgMsg)
}

func (g *Core) drawCrosshairs(screen *ebiten.Image) {
	cl := g.rgs[RegoterEnumCrosshair]
	cl.ForEach(func(i ID, r *regoterInCore) {
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest

		crosshairScale := r.sprite.Scale()
		op.GeoM.Scale(crosshairScale, crosshairScale)
		op.GeoM.Translate(
			float64(g.cfg.ScreenWidth)/2-float64(r.sprite.W)*crosshairScale/2,
			float64(g.cfg.ScreenHeight)/2-float64(r.sprite.H)*crosshairScale/2,
		)
		screen.DrawImage(r.sprite.Texture(), op)
		// if g.crosshairs.IsHitIndicatorActive() {
		// 	screen.DrawImage(g.crosshairs.HitIndicator.Texture(), op)
		// 	g.crosshairs.Update()
		// }
	})
}

func (g *Core) drawMiniMap(screen *ebiten.Image) {
	// draw minimap
	mm := g.miniMap()
	mmImg := ebiten.NewImageFromImage(mm)
	if mmImg != nil {
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest

		op.GeoM.Scale(5.0, 5.0)
		op.GeoM.Translate(0, 50)
		screen.DrawImage(mmImg, op)
	}
}

func drawSpriteBox(screen *ebiten.Image, sprite *Sprite) {
	r := sprite.ScreenRect()
	if r == nil {
		return
	}

	minX, minY := float32(r.Min.X), float32(r.Min.Y)
	maxX, maxY := float32(r.Max.X), float32(r.Max.Y)

	vector.StrokeRect(screen, minX, minY, maxX-minX, maxY-minY, 1, color.RGBA{255, 0, 0, 255}, false)
}

func drawSpriteIndicator(screen *ebiten.Image, sprite *Sprite) {
	r := sprite.ScreenRect()
	if r == nil {
		return
	}

	dX, dY := float32(r.Dx())/8, float32(r.Dy())/8
	midX, minY := float32(r.Max.X)-float32(r.Dx())/2, float32(r.Min.Y)-dY

	vector.StrokeLine(screen, midX, minY+dY, midX-dX, minY, 1, color.RGBA{0, 255, 0, 255}, false)
	vector.StrokeLine(screen, midX, minY+dY, midX+dX, minY, 1, color.RGBA{0, 255, 0, 255}, false)
	vector.StrokeLine(screen, midX-dX, minY, midX+dX, minY, 1, color.RGBA{0, 255, 0, 255}, false)
}
