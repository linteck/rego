package model

import (
	"fmt"
	"lintech/rego/iregoter"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go/geom"
)

func handlePlayerInput(cfg iregoter.GameCfg, lastPosition *iregoter.MousePosition) iregoter.RegoterMove {
	movement := iregoter.RegoterMove{}
	forward := false
	backward := false
	rotLeft := false
	rotRight := false

	moveModifier := 1.0
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		moveModifier = 2.0
	}

	switch {
	case ebiten.IsKeyPressed(ebiten.KeyControl) && cfg.OsType == iregoter.OsTypeBrowser:
		// debug cursor mode not intended for browser purposes
		if cfg.MouseMode != iregoter.MouseModeCursor {
			ebiten.SetCursorMode(ebiten.CursorModeVisible)
			cfg.MouseMode = iregoter.MouseModeCursor
		}

	case ebiten.IsKeyPressed(ebiten.KeyAlt):
		if cfg.MouseMode != iregoter.MouseModeMove {
			ebiten.SetCursorMode(ebiten.CursorModeCaptured)
			cfg.MouseMode = iregoter.MouseModeMove
			lastPosition.X, lastPosition.Y = math.MinInt32, math.MinInt32
		}

	case cfg.MouseMode != iregoter.MouseModeLook:
		ebiten.SetCursorMode(ebiten.CursorModeCaptured)
		cfg.MouseMode = iregoter.MouseModeLook
		lastPosition.X, lastPosition.Y = math.MinInt32, math.MinInt32
	}

	switch cfg.MouseMode {
	case iregoter.MouseModeCursor:
		lastPosition.X, lastPosition.Y = ebiten.CursorPosition()
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			fmt.Printf("mouse left clicked: (%v, %v)\n", lastPosition.X, lastPosition.Y)
		}

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
			fmt.Printf("mouse right clicked: (%v, %v)\n", lastPosition.X, lastPosition.Y)
		}

	case iregoter.MouseModeMove:
		x, y := ebiten.CursorPosition()

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			movement.FireWeapon = true
			// p.fireWeapon()
		}

		isStrafeMove := false
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
			// hold right click in this mode to strafe instead of rotate with mouse X axis
			isStrafeMove = true
		}

		switch {
		case lastPosition.X == math.MinInt32 && lastPosition.Y == math.MinInt32:
			// initialize first position to establish delta
			if x != 0 && y != 0 {
				lastPosition.X, lastPosition.Y = x, y
			}

		default:
			dx, dy := lastPosition.X-x, lastPosition.Y-y
			lastPosition.X, lastPosition.Y = x, y

			if dx != 0 {
				if isStrafeMove {
					movement.MoveRotate = -geom.HalfPi
					movement.Acceleration = 0.01 * float64(dx) * moveModifier
				} else {
					movement.VissionRotate = 0.005 * float64(dx) * moveModifier
				}
			}

			if dy != 0 {
				movement.Acceleration = 0.01 * float64(dy) * moveModifier
			}
		}
	case iregoter.MouseModeLook:
		x, y := ebiten.CursorPosition()

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			movement.FireWeapon = true
		}

		// Todo
		// if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		// 	// hold right click to zoom view in this mode
		// 	if p.camera.FovDepth() != p.zoomFovDepth {
		// 		zoomFovDegrees := p.fovDegrees / p.zoomFovDepth
		// 		p.camera.SetFovAngle(zoomFovDegrees, p.zoomFovDepth)
		// 		p.camera.SetPitchAngle(p.Entity.Pitch)
		// 	}
		// } else if p.camera.FovDepth() == p.zoomFovDepth {
		// 	// unzoom
		// 	p.camera.SetFovAngle(p.fovDegrees, 1.0)
		// 	p.camera.SetPitchAngle(p.Entity.Pitch)
		// }

		switch {
		case lastPosition.X == math.MinInt32 && lastPosition.Y == math.MinInt32:
			// initialize first position to establish delta
			if x != 0 && y != 0 {
				lastPosition.X, lastPosition.Y = x, y
			}

		default:
			dx, dy := lastPosition.X-x, lastPosition.Y-y
			lastPosition.X, lastPosition.Y = x, y

			if dx != 0 {
				movement.VissionRotate = 0.005 * float64(dx) * moveModifier
			}

			if dy != 0 {
				movement.PitchRotate = 0.005 * float64(dy) * moveModifier
			}
		}
	}

	// _, wheelY := ebiten.Wheel()
	// if wheelY != 0 {
	// 	p.NextWeapon(wheelY > 0)
	// }
	// if ebiten.IsKeyPressed(ebiten.KeyDigit1) {
	// 	p.SelectWeapon(0)
	// }
	// if ebiten.IsKeyPressed(ebiten.KeyDigit2) {
	// 	p.SelectWeapon(1)
	// }
	// if ebiten.IsKeyPressed(ebiten.KeyH) {
	// 	// put away/holster weapon
	// 	p.SelectWeapon(-1)
	// }

	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		rotLeft = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		rotRight = true
	}

	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		forward = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		backward = true
	}

	// if ebiten.IsKeyPressed(ebiten.KeyC) {
	// 	p.Crouch()
	// } else if ebiten.IsKeyPressed(ebiten.KeyZ) {
	// 	p.Prone()
	// } else if ebiten.IsKeyPressed(ebiten.KeySpace) {
	// 	p.Jump()
	// } else if !p.IsStanding() {
	// 	p.Stand()
	// }

	if forward {
		movement.Acceleration = 0.06 * moveModifier
	} else if backward {
		movement.Acceleration = 0.06 * moveModifier
		movement.MoveRotate = geom.Pi
	}

	if cfg.MouseMode == iregoter.MouseModeLook || cfg.MouseMode == iregoter.MouseModeMove {
		// strafe instead of rotate
		if rotLeft {
			movement.Acceleration = 0.05 * moveModifier
			movement.MoveRotate = geom.HalfPi
		} else if rotRight {
			movement.Acceleration = 0.05 * moveModifier
			movement.MoveRotate = -geom.HalfPi
		}
	} else {
		if rotLeft {
			movement.VissionRotate = -0.03 * moveModifier
		} else if rotRight {
			movement.VissionRotate = 0.03 * moveModifier
		}
	}

	return movement
}
