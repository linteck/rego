package game

import (
	"math"
	"math/rand"
	"os"

	"image/color"
	_ "image/png"

	"lintech/rego/game/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
)

// Game - This is the main type for your game.

func drawSpriteBox(screen *ebiten.Image, sprite *model.Sprite) {
	r := sprite.ScreenRect()
	if r == nil {
		return
	}

	minX, minY := float32(r.Min.X), float32(r.Min.Y)
	maxX, maxY := float32(r.Max.X), float32(r.Max.Y)

	vector.StrokeRect(screen, minX, minY, maxX-minX, maxY-minY, 1, color.RGBA{255, 0, 0, 255}, false)
}

func drawSpriteIndicator(screen *ebiten.Image, sprite *model.Sprite) {
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

func (g *Game) setFullscreen(fullscreen bool) {
	g.fullscreen = fullscreen
	ebiten.SetFullscreen(fullscreen)
}

func (g *Game) setResolution(screenWidth, screenHeight int) {
	g.screenWidth, g.screenHeight = screenWidth, screenHeight
	ebiten.SetWindowSize(screenWidth, screenHeight)
	g.setRenderScale(g.renderScale)
}

func (g *Game) setRenderScale(renderScale float64) {
	g.renderScale = renderScale
	g.width = int(math.Floor(float64(g.screenWidth) * g.renderScale))
	g.height = int(math.Floor(float64(g.screenHeight) * g.renderScale))
	if g.camera != nil {
		g.camera.SetViewSize(g.width, g.height)
	}
	g.scene = ebiten.NewImage(g.width, g.height)
}

func (g *Game) setRenderDistance(renderDistance float64) {
	g.renderDistance = renderDistance
	g.camera.SetRenderDistance(g.renderDistance)
}

func (g *Game) setLightFalloff(lightFalloff float64) {
	g.lightFalloff = lightFalloff
	g.camera.SetLightFalloff(g.lightFalloff)
}

func (g *Game) setGlobalIllumination(globalIllumination float64) {
	g.globalIllumination = globalIllumination
	g.camera.SetGlobalIllumination(g.globalIllumination)
}

func (g *Game) setLightRGB(minLightRGB, maxLightRGB color.NRGBA) {
	g.minLightRGB = minLightRGB
	g.maxLightRGB = maxLightRGB
	g.camera.SetLightRGB(g.minLightRGB, g.maxLightRGB)
}

func (g *Game) setVsyncEnabled(enableVsync bool) {
	g.vsync = enableVsync
	ebiten.SetVsyncEnabled(enableVsync)
}

func (g *Game) setFovAngle(fovDegrees float64) {
	g.fovDegrees = fovDegrees
	g.camera.SetFovAngle(fovDegrees, 1.0)
}

// Move player by move speed in the forward/backward direction
func (g *Game) Move(mSpeed float64) {
	moveLine := geom.LineFromAngle(g.player.Position.X, g.player.Position.Y, g.player.Angle, mSpeed)

	newPos, _, _ := g.getValidMove(g.player.Entity, moveLine.X2, moveLine.Y2, g.player.PositionZ, true)
	if !newPos.Equals(g.player.Pos()) {
		g.player.Position = newPos
		g.player.Moved = true
	}
}

// Move player by strafe speed in the left/right direction
func (g *Game) Strafe(sSpeed float64) {
	strafeAngle := geom.HalfPi
	if sSpeed < 0 {
		strafeAngle = -strafeAngle
	}
	strafeLine := geom.LineFromAngle(g.player.Position.X, g.player.Position.Y, g.player.Angle-strafeAngle, math.Abs(sSpeed))

	newPos, _, _ := g.getValidMove(g.player.Entity, strafeLine.X2, strafeLine.Y2, g.player.PositionZ, true)
	if !newPos.Equals(g.player.Pos()) {
		g.player.Position = newPos
		g.player.Moved = true
	}
}

// Rotate player heading angle by rotation speed
func (g *Game) Rotate(rSpeed float64) {
	g.player.Angle += rSpeed

	pi2 := geom.Pi2
	if g.player.Angle >= pi2 {
		g.player.Angle = pi2 - g.player.Angle
	} else if g.player.Angle <= -pi2 {
		g.player.Angle = g.player.Angle + pi2
	}

	g.player.Moved = true
}

// Update player pitch angle by pitch speed
func (g *Game) Pitch(pSpeed float64) {
	// current raycasting method can only allow up to 22.5 degrees down, 45 degrees up
	g.player.Pitch = geom.Clamp(pSpeed+g.player.Pitch, -math.Pi/8, math.Pi/4)
	g.player.Moved = true
}

func (g *Game) Stand() {
	g.player.CameraZ = 0.5
	g.player.PositionZ = 0
	g.player.Moved = true
}

func (g *Game) IsStanding() bool {
	return g.player.PositionZ == 0 && g.player.CameraZ == 0.5
}

func (g *Game) Jump() {
	g.player.CameraZ = 0.9
	g.player.PositionZ = 0.4
	g.player.Moved = true
}

func (g *Game) Crouch() {
	g.player.CameraZ = 0.3
	g.player.PositionZ = 0
	g.player.Moved = true
}

func (g *Game) Prone() {
	g.player.CameraZ = 0.1
	g.player.PositionZ = 0
	g.player.Moved = true
}

func (g *Game) fireWeapon() {
	w := g.player.Weapon
	if w == nil {
		g.player.NextWeapon(false)
		return
	}
	if w.OnCooldown() {
		return
	}

	// set weapon firing for animation to run
	w.Fire()

	// spawning projectile at player position just slightly below player's center point of view
	pX, pY, pZ := g.player.Position.X, g.player.Position.Y, geom.Clamp(g.player.CameraZ-0.1, 0.05, 0.95)
	// pitch, angle based on raycasted point at crosshairs
	var pAngle, pPitch float64
	convergenceDistance := g.camera.GetConvergenceDistance()
	convergencePoint := g.camera.GetConvergencePoint()
	if convergenceDistance <= 0 || convergencePoint == nil {
		pAngle, pPitch = g.player.Angle, g.player.Pitch
	} else {
		convergenceLine3d := &geom3d.Line3d{
			X1: pX, Y1: pY, Z1: pZ,
			X2: convergencePoint.X, Y2: convergencePoint.Y, Z2: convergencePoint.Z,
		}
		pAngle, pPitch = convergenceLine3d.Heading(), convergenceLine3d.Pitch()
	}

	projectile := w.SpawnProjectile(pX, pY, pZ, pAngle, pPitch, g.player.Entity)
	if projectile != nil {
		g.addProjectile(projectile)
	}
}

// Update camera to match player position and orientation
func (g *Game) updatePlayerCamera(forceUpdate bool) {
	if !g.player.Moved && !forceUpdate {
		// only update camera position if player moved or forceUpdate set
		return
	}

	// reset player moved flag to only update camera when necessary
	g.player.Moved = false

	g.camera.SetPosition(g.player.Position.Copy())
	g.camera.SetPositionZ(g.player.CameraZ)
	g.camera.SetHeadingAngle(g.player.Angle)
	g.camera.SetPitchAngle(g.player.Pitch)
}

func (g *Game) updateProjectiles() {
	// Testing animated projectile movement
	for p := range g.projectiles {
		if p.Velocity != 0 {

			trajectory := geom3d.Line3dFromAngle(p.Position.X, p.Position.Y, p.PositionZ, p.Angle, p.Pitch, p.Velocity)

			xCheck := trajectory.X2
			yCheck := trajectory.Y2
			zCheck := trajectory.Z2

			newPos, isCollision, collisions := g.getValidMove(p.Entity, xCheck, yCheck, zCheck, false)
			if isCollision || p.PositionZ <= 0 {
				// for testing purposes, projectiles instantly get deleted when collision occurs
				g.deleteProjectile(p)

				// make a sprite/wall getting hit by projectile cause some visual effect
				if p.ImpactEffect.Sprite != nil {
					if len(collisions) >= 1 {
						// use the first collision point to place effect at
						newPos = collisions[0].collision
					}

					// TODO: give impact effect optional ability to have some velocity based on the projectile movement upon impact if it didn't hit a wall
					effect := p.SpawnEffect(newPos.X, newPos.Y, p.PositionZ, p.Angle, p.Pitch)

					g.addEffect(effect)
				}

				for _, collisionEntity := range collisions {
					if collisionEntity.entity == g.player.Entity {
						println("ouch!")
					} else {
						// show crosshair hit effect
						// Todo
						// g.crosshairs.ActivateHitIndicator(30)
					}
				}
			} else {
				p.Position = newPos
				p.PositionZ = zCheck
			}
		}
		p.Update(g.player.Position)
	}

	// Testing animated effects (explosions)
	for e := range g.effects {
		e.Update(g.player.Position)
		if e.LoopCounter() >= e.LoopCount {
			g.deleteEffect(e)
		}
	}
}

func (g *Game) updateSprites() {
	// Testing animated sprite movement
	for s := range g.sprites {
		if s.Velocity != 0 {
			vLine := geom.LineFromAngle(s.Position.X, s.Position.Y, s.Angle, s.Velocity)

			xCheck := vLine.X2
			yCheck := vLine.Y2
			zCheck := s.PositionZ

			newPos, isCollision, _ := g.getValidMove(s.Entity, xCheck, yCheck, zCheck, false)
			if isCollision {
				// for testing purposes, letting the sample sprite ping pong off walls in somewhat random direction
				s.Angle = randFloat(-math.Pi, math.Pi)
				s.Velocity = randFloat(0.01, 0.03)
			} else {
				s.Position = newPos
			}
		}
		s.Update(g.player.Position)
	}
}

func randFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func exit(rc int) {
	// TODO: any cleanup?
	os.Exit(rc)
}
