package model

import (
	"fmt"
	"lintech/rego/iregoter"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

func (p *Player) handleInput(menuActive bool, si *iregoter.MouseInfo) {
	forward := false
	backward := false
	rotLeft := false
	rotRight := false

	moveModifier := 1.0
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		moveModifier = 2.0
	}

	switch {
	case ebiten.IsKeyPressed(ebiten.KeyControl) && p.cfg.OsType == iregoter.OsTypeBrowser:
		// debug cursor mode not intended for browser purposes
		if si.MouseMode != iregoter.MouseModeCursor {
			ebiten.SetCursorMode(ebiten.CursorModeVisible)
			si.MouseMode = iregoter.MouseModeCursor
		}

	case ebiten.IsKeyPressed(ebiten.KeyAlt):
		if si.MouseMode != iregoter.MouseModeMove {
			ebiten.SetCursorMode(ebiten.CursorModeCaptured)
			si.MouseMode = iregoter.MouseModeMove
			si.MouseX, si.MouseY = math.MinInt32, math.MinInt32
		}

	case !menuActive && si.MouseMode != iregoter.MouseModeLook:
		ebiten.SetCursorMode(ebiten.CursorModeCaptured)
		si.MouseMode = iregoter.MouseModeLook
		si.MouseX, si.MouseY = math.MinInt32, math.MinInt32
	}

	switch si.MouseMode {
	case iregoter.MouseModeCursor:
		si.MouseX, si.MouseY = ebiten.CursorPosition()
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			fmt.Printf("mouse left clicked: (%v, %v)\n", si.MouseX, si.MouseY)
		}

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
			fmt.Printf("mouse right clicked: (%v, %v)\n", si.MouseX, si.MouseY)
		}

	case iregoter.MouseModeMove:
		x, y := ebiten.CursorPosition()

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			p.fireWeapon()
		}

		isStrafeMove := false
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
			// hold right click in this mode to strafe instead of rotate with mouse X axis
			isStrafeMove = true
		}

		switch {
		case si.MouseX == math.MinInt32 && si.MouseY == math.MinInt32:
			// initialize first position to establish delta
			if x != 0 && y != 0 {
				si.MouseX, si.MouseY = x, y
			}

		default:
			dx, dy := si.MouseX-x, si.MouseY-y
			si.MouseX, si.MouseY = x, y

			if dx != 0 {
				if isStrafeMove {
					p.Strafe(iregoter.Distance(-0.01 * float64(dx) * moveModifier))
				} else {
					p.Rotate(iregoter.RotateAngle(0.005 * float64(dx) * moveModifier))
				}
			}

			if dy != 0 {
				p.Move(iregoter.Distance(0.01 * float64(dy) * moveModifier))
			}
		}
	case iregoter.MouseModeLook:
		x, y := ebiten.CursorPosition()

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			p.fireWeapon()
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
		case si.MouseX == math.MinInt32 && si.MouseY == math.MinInt32:
			// initialize first position to establish delta
			if x != 0 && y != 0 {
				si.MouseX, si.MouseY = x, y
			}

		default:
			dx, dy := si.MouseX-x, si.MouseY-y
			si.MouseX, si.MouseY = x, y

			if dx != 0 {
				p.Rotate(iregoter.RotateAngle(0.005 * float64(dx) * moveModifier))
			}

			if dy != 0 {
				p.Pitch(iregoter.PitchAngle(0.005 * float64(dy)))
			}
		}
	}

	_, wheelY := ebiten.Wheel()
	if wheelY != 0 {
		p.NextWeapon(wheelY > 0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyDigit1) {
		p.SelectWeapon(0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyDigit2) {
		p.SelectWeapon(1)
	}
	if ebiten.IsKeyPressed(ebiten.KeyH) {
		// put away/holster weapon
		p.SelectWeapon(-1)
	}

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

	if ebiten.IsKeyPressed(ebiten.KeyC) {
		p.Crouch()
	} else if ebiten.IsKeyPressed(ebiten.KeyZ) {
		p.Prone()
	} else if ebiten.IsKeyPressed(ebiten.KeySpace) {
		p.Jump()
	} else if !p.IsStanding() {
		p.Stand()
	}

	if forward {
		p.Move(iregoter.Distance(0.06 * moveModifier))
	} else if backward {
		p.Move(iregoter.Distance(-0.06 * moveModifier))
	}

	if si.MouseMode == iregoter.MouseModeLook || si.MouseMode == iregoter.MouseModeMove {
		// strafe instead of rotate
		if rotLeft {
			p.Strafe(iregoter.Distance(-0.05 * moveModifier))
		} else if rotRight {
			p.Strafe(iregoter.Distance(0.05 * moveModifier))
		}
	} else {
		if rotLeft {
			p.Rotate(iregoter.RotateAngle(0.03 * moveModifier))
		} else if rotRight {
			p.Rotate(iregoter.RotateAngle(-0.03 * moveModifier))
		}
	}
}
