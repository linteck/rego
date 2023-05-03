package game

import (
	"fmt"
	"lintech/rego/iregoter"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/harbdog/raycaster-go"
)

func (g *Core) eventHandleGameEventDraw(e iregoter.GameEventDraw) {

	g.drawScreen(e.Screen)

	// Debug
	// g.removeAllUnregisteredRogeter()
	r := iregoter.CoreEventDrawDone{}
	g.txToGame <- r

}

func (g *Core) drawScreen(screen *ebiten.Image) {
	// Put projectiles together with sprites for raycasting both as sprites
	sl := g.rgs[iregoter.RegoterEnumSprite]
	numSprites := sl.Len()
	raycastSprites := make([]raycaster.Sprite, numSprites)

	index := 0
	sl.ForEach(func(i iregoter.ID, val *regoterInCore) {
		if val.sprite != nil {
			raycastSprites[index] = val.sprite
			index += 1
		}
	})

	// Update camera (calculate raycast)
	g.camera.Update(raycastSprites)

	// Render raycast scene
	g.camera.Draw(g.scene)

	// draw equipped weapon
	// apply lighting setting

	if g.cfg.ShowSpriteBoxes {
		// draw sprite screen indicators to show we know where it was raycasted (must occur after camera.Update)
		sl.ForEach(func(i iregoter.ID, val *regoterInCore) {
			if val.sprite != nil {
				drawSpriteBox(g.scene, val.sprite)
			}
		})
	}

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
	cl := g.rgs[iregoter.RegoterEnumCrosshair]
	cl.ForEach(func(i iregoter.ID, r *regoterInCore) {
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
