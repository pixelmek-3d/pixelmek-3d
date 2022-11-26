package game

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type MouseMode int

const (
	MouseModeTurret MouseMode = iota
	MouseModeBody
	MouseModeCursor
)

func (g *Game) handleInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if g.menu.active {
			g.closeMenu()
		} else {
			g.openMenu()
		}
	}

	if g.paused {
		// currently only paused when menu is active, one could consider other pauses not the subject of this demo
		return
	}

	var stop, forward, backward bool
	var rotLeft, rotRight bool

	moveModifier := 1.0
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		moveModifier = 2.0
	}

	switch {
	case ebiten.IsKeyPressed(ebiten.KeyControl):
		if g.mouseMode != MouseModeCursor {
			ebiten.SetCursorMode(ebiten.CursorModeVisible)
			g.mouseMode = MouseModeCursor
		}

	case ebiten.IsKeyPressed(ebiten.KeyAlt):
		if g.mouseMode != MouseModeBody {
			ebiten.SetCursorMode(ebiten.CursorModeCaptured)
			g.mouseMode = MouseModeBody
			g.mouseX, g.mouseY = math.MinInt32, math.MinInt32
		}

	case !g.menu.active && g.mouseMode != MouseModeTurret:
		ebiten.SetCursorMode(ebiten.CursorModeCaptured)
		g.mouseMode = MouseModeTurret
		g.mouseX, g.mouseY = math.MinInt32, math.MinInt32
	}

	switch g.mouseMode {
	case MouseModeCursor:
		g.mouseX, g.mouseY = ebiten.CursorPosition()
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			fmt.Printf("mouse left clicked: (%v, %v)\n", g.mouseX, g.mouseY)
		}

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
			fmt.Printf("mouse right clicked: (%v, %v)\n", g.mouseX, g.mouseY)
		}

	case MouseModeBody:
		x, y := ebiten.CursorPosition()

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			g.fireWeapon()
		}

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
			// hold right click to zoom view in this mode
			if g.camera.FovDepth() != g.zoomFovDepth {
				zoomFovDegrees := g.fovDegrees / g.zoomFovDepth
				g.camera.SetFovAngle(zoomFovDegrees, g.zoomFovDepth)
				g.camera.SetPitchAngle(g.player.Pitch())
			}
		} else if g.camera.FovDepth() == g.zoomFovDepth {
			// unzoom
			g.camera.SetFovAngle(g.fovDegrees, 1.0)
			g.camera.SetPitchAngle(g.player.Pitch())
		}

		switch {
		case g.mouseX == math.MinInt32 && g.mouseY == math.MinInt32:
			// initialize first position to establish delta
			if x != 0 && y != 0 {
				g.mouseX, g.mouseY = x, y
			}

		default:
			dx, dy := g.mouseX-x, g.mouseY-y
			g.mouseX, g.mouseY = x, y

			if dx != 0 {
				g.Rotate(0.005 * float64(dx) * moveModifier)
			}

			if dy != 0 {
				g.Pitch(0.005 * float64(dy))
			}
		}
	case MouseModeTurret:
		x, y := ebiten.CursorPosition()

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			g.fireWeapon()
		}

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
			// TESTING purposes only
			g.fireTestWeaponAtPlayer()
		}

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
			// hold right click to zoom view in this mode
			if g.camera.FovDepth() != g.zoomFovDepth {
				zoomFovDegrees := g.fovDegrees / g.zoomFovDepth
				g.camera.SetFovAngle(zoomFovDegrees, g.zoomFovDepth)
				g.camera.SetPitchAngle(g.player.Pitch())
			}
		} else if g.camera.FovDepth() == g.zoomFovDepth {
			// unzoom
			g.camera.SetFovAngle(g.fovDegrees, 1.0)
			g.camera.SetPitchAngle(g.player.Pitch())
		}

		switch {
		case g.mouseX == math.MinInt32 && g.mouseY == math.MinInt32:
			// initialize first position to establish delta
			if x != 0 && y != 0 {
				g.mouseX, g.mouseY = x, y
			}

		default:
			dx, dy := g.mouseX-x, g.mouseY-y
			g.mouseX, g.mouseY = x, y

			if dx != 0 {
				if g.player.HasTurret() {
					g.RotateTurret(0.005 * float64(dx) * moveModifier)
				} else {
					g.Rotate(0.005 * float64(dx) * moveModifier)
				}
			}

			if dy != 0 {
				g.Pitch(0.005 * float64(dy))
			}
		}
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

	if ebiten.IsKeyPressed(ebiten.KeyX) {
		stop = true
	}

	switch g.throttleDecay {
	case true:
		if forward {
			g.player.SetTargetVelocity(g.player.MaxVelocity())
		} else if backward {
			g.player.SetTargetVelocity(-g.player.MaxVelocity() / 2)
		} else {
			g.player.SetTargetVelocity(0)
		}
	case false:
		deltaV := 0.0004 // FIXME: testing
		if forward {
			g.player.SetTargetVelocity(g.player.TargetVelocity() + deltaV)
		} else if backward {
			g.player.SetTargetVelocity(g.player.TargetVelocity() - deltaV)
		} else if stop {
			g.player.SetTargetVelocity(0)
		}
	}

	if g.mouseMode == MouseModeBody {
		// TODO: only infantry/battle armor and VTOL can strafe
		// strafe instead of rotate
		if rotLeft {
			g.Strafe(-0.05 * moveModifier)
		} else if rotRight {
			g.Strafe(0.05 * moveModifier)
		}
	} else {
		if rotLeft {
			//g.Rotate(0.03 * moveModifier)
			turnAmount := g.player.TurnRate()
			g.player.SetTargetRelativeHeading(turnAmount)
		} else if rotRight {
			//g.Rotate(-0.03 * moveModifier)
			turnAmount := g.player.TurnRate()
			g.player.SetTargetRelativeHeading(-turnAmount)
		}
	}
}
