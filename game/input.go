package game

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/raycaster-go/geom"
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

	_, isInfantry := g.player.Unit.(*model.Infantry)
	_, isVTOL := g.player.Unit.(*model.VTOL)

	var stop, forward, backward bool
	var rotLeft, rotRight bool

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
		// TESTING purposes only
		g.fireTestWeaponAtPlayer()
	}

	switch {
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

		// if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		//     TODO: refactor the new stuff for chain/group fire using shared functions
		// }

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
				turnRate := g.player.TurnRate()
				turnAmount := geom.Clamp(0.1*float64(dx), -turnRate, turnRate)
				g.player.SetTargetRelativeHeading(turnAmount)
			}

			if dy != 0 {
				g.Pitch(0.005 * float64(dy) / g.zoomFovDepth)
			}
		}
	case MouseModeTurret:
		x, y := ebiten.CursorPosition()

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			g.fireWeapon()
		}

		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			if g.player.fireMode == model.CHAIN_FIRE {
				// cycle to next weapon only in same group (g.player.selectedGroup)
				prevWeapon := g.player.Armament()[g.player.selectedWeapon]
				groupWeapons := g.player.weaponGroups[g.player.selectedGroup]

				var nextWeapon model.Weapon
				nextIndex := 0
				for i, w := range groupWeapons {
					if w == prevWeapon {
						nextIndex = i + 1
						break
					}
				}
				if nextIndex >= len(groupWeapons) {
					nextIndex = 0
				}
				nextWeapon = groupWeapons[nextIndex]

				for i, w := range g.player.Armament() {
					if w == nextWeapon {
						g.player.selectedWeapon = uint(i)
						break
					}
				}
			}
		}

		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
			if g.player.fireMode == model.GROUP_FIRE {
				g.player.selectedGroup++
				if int(g.player.selectedGroup) >= len(g.player.weaponGroups) {
					g.player.selectedGroup = 0
				}

				// set next selectedGroup only if >0 weapons in it
				weaponsInGroup := len(g.player.weaponGroups[g.player.selectedGroup])
				for weaponsInGroup == 0 {
					g.player.selectedGroup++
					if int(g.player.selectedGroup) >= len(g.player.weaponGroups) {
						g.player.selectedGroup = 0
					}
					weaponsInGroup = len(g.player.weaponGroups[g.player.selectedGroup])
				}

			} else if g.player.fireMode == model.CHAIN_FIRE {
				g.player.selectedWeapon++
				if int(g.player.selectedWeapon) >= len(g.player.Armament()) {
					g.player.selectedWeapon = 0
				}

				// set selectedGroup if the newly selected weapon is in different group
				newSelectedWeapon := g.player.Armament()[g.player.selectedWeapon]
				groups := model.GetGroupsForWeapon(newSelectedWeapon, g.player.weaponGroups)
				if len(groups) == 0 {
					g.player.selectedGroup = 0
				} else {
					g.player.selectedGroup = groups[0]
				}
			}
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
					g.RotateTurret(0.005 * float64(dx) / g.zoomFovDepth)
				} else {
					turnRate := g.player.TurnRate()
					turnAmount := geom.Clamp(0.005*float64(dx), -turnRate, turnRate)
					g.player.SetTargetRelativeHeading(turnAmount)
				}
			}

			if dy != 0 {
				g.Pitch(0.005 * float64(dy))
			}
		}
	}

	if g.player.fireMode == model.CHAIN_FIRE && ebiten.IsKeyPressed(ebiten.KeyShift) {
		// set group for selected weapon
		setGroupIndex := -1
		switch {
		case inpututil.IsKeyJustPressed(ebiten.Key1):
			setGroupIndex = 0
		case inpututil.IsKeyJustPressed(ebiten.Key2):
			setGroupIndex = 1
		case inpututil.IsKeyJustPressed(ebiten.Key3):
			setGroupIndex = 2
		}

		if setGroupIndex >= 0 {
			weapon := g.player.Armament()[g.player.selectedWeapon]
			groups := model.GetGroupsForWeapon(weapon, g.player.weaponGroups)
			for _, gIndex := range groups {
				if int(gIndex) == setGroupIndex {
					// already in group
					return
				} else {
					// remove from current group
					weaponsInGroup := g.player.weaponGroups[gIndex]
					g.player.weaponGroups[gIndex] = make([]model.Weapon, 0, len(weaponsInGroup)-1)
					for _, chkWeapon := range weaponsInGroup {
						if chkWeapon != weapon {
							g.player.weaponGroups[gIndex] = append(g.player.weaponGroups[gIndex], chkWeapon)
						}
					}
				}
			}

			// add to selected group
			g.player.weaponGroups[setGroupIndex] = append(g.player.weaponGroups[setGroupIndex], weapon)
			g.player.selectedGroup = uint(setGroupIndex)
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyBackslash) {
		// toggle group fire mode
		switch g.player.fireMode {
		case model.CHAIN_FIRE:
			g.player.fireMode = model.GROUP_FIRE
		case model.GROUP_FIRE:
			g.player.fireMode = model.CHAIN_FIRE
		}

		if g.player.fireMode == model.GROUP_FIRE {
			// select the first appropriate group from selected weapon when switching to group mode
			prevSelectedWeapon := g.player.Armament()[g.player.selectedWeapon]
			groups := model.GetGroupsForWeapon(prevSelectedWeapon, g.player.weaponGroups)
			if len(groups) == 0 {
				g.player.selectedGroup = 0
			} else {
				g.player.selectedGroup = groups[0]
			}
		} else if g.player.fireMode == model.CHAIN_FIRE {
			// select the first weapon of the group that was selected when switching to chain mode
			prevSelectedGroup := g.player.selectedGroup
			weapons := g.player.weaponGroups[prevSelectedGroup]
			if len(weapons) == 0 {
				g.player.selectedWeapon = 0
			} else {
				for i, w := range g.player.Armament() {
					if w == weapons[0] {
						g.player.selectedWeapon = uint(i)
						break
					}
				}
			}
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		// cycle nav points
		g.navPointCycle()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		// target on crosshairs
		g.targetCrosshairs()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		// target nearest to player
		g.targetCycle(TARGET_NEAREST)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyT) {
		// cycle player targets
		g.targetCycle(TARGET_NEXT)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		// cycle player targets in reverse order
		g.targetCycle(TARGET_PREVIOUS)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		// toggle zoom
		if g.camera.FovDepth() != g.zoomFovDepth {
			// zoom in
			zoomFovDegrees := g.fovDegrees / g.zoomFovDepth
			g.camera.SetFovAngle(zoomFovDegrees, g.zoomFovDepth)
			g.camera.SetPitchAngle(g.player.Pitch())
		} else {
			// zoom out
			g.camera.SetFovAngle(g.fovDegrees, 1.0)
			g.camera.SetPitchAngle(g.player.Pitch())
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		// toggle reverse throttle
		if g.player.TargetVelocity() > 0 {
			// switch to reverse
			vPercent := g.player.TargetVelocity() / g.player.MaxVelocity()
			g.player.SetTargetVelocity(-vPercent * g.player.MaxVelocity() / 2)
		} else if g.player.TargetVelocity() < 0 {
			// switch to forward
			vPercent := math.Abs(g.player.TargetVelocity()) / (g.player.MaxVelocity() / 2)
			g.player.SetTargetVelocity(vPercent * g.player.MaxVelocity())
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		if isVTOL {
			// TODO: use unit max velocity to determine ascend speed
			g.VerticalMove(0.05)
		}
		// TODO: else jump, if jump jets
	}

	if ebiten.IsKeyPressed(ebiten.KeyControl) {
		if isVTOL {
			// TODO: use unit max velocity to determine descend speed
			g.VerticalMove(-0.05)
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

	isStrafe := false
	if !g.player.HasTurret() && (rotLeft || rotRight) {
		// only infantry/battle armor and VTOL can strafe
		if isInfantry || isVTOL {
			// strafe instead of rotate
			isStrafe = true
		}
	}

	if isStrafe {
		// TODO: use unit max velocity to determine strafe speed
		if rotLeft {
			g.Strafe(-0.05)
		} else if rotRight {
			g.Strafe(0.05)
		}
	} else {
		if rotLeft {
			turnAmount := g.player.TurnRate()
			g.player.SetTargetRelativeHeading(turnAmount)
		} else if rotRight {
			turnAmount := g.player.TurnRate()
			g.player.SetTargetRelativeHeading(-turnAmount)
		}
	}
}
