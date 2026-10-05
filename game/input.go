package game

import (
	"fmt"
	"image/color"
	"math"
	"os"
	"runtime/pprof"
	"slices"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
	input "github.com/quasilyte/ebitengine-input"
	log "github.com/sirupsen/logrus"
)

type MouseMode int

const (
	MouseModeTurret MouseMode = iota
	MouseModeBody
	MouseModeCursor
)

var debugProfFile *os.File

type InputEvent struct {
	input.EventInfo
	Dx, Dy float64
}

type InputSensitivities struct {
	mouse   *AxesSensitivity
	gamepad *AxesSensitivity
}

type AxesSensitivity struct {
	X, Y float64
}

// convertToAxesSensitivity handles loading AxesSensitivity settings from viper config interface
func convertToAxesSensitivity(input any) (*AxesSensitivity, error) {
	if sensitivities, ok := input.(AxesSensitivity); ok {
		return &sensitivities, nil
	}
	sensitivities := &AxesSensitivity{}
	if rawMap, ok := input.(map[string]any); ok {
		for k, v := range rawMap {
			var vFloat float64
			switch val := v.(type) {
			case float64:
				vFloat = val
			default:
				return sensitivities, fmt.Errorf("Axes sensitivity key %q has unsupported value type: %T", k, v)
			}

			switch {
			case strings.ToUpper(k) == `X`:
				sensitivities.X = vFloat
			case strings.ToUpper(k) == `Y`:
				sensitivities.Y = vFloat
			default:
				return sensitivities, fmt.Errorf("Aaxes sensitivity unknown key: %q", k)
			}
		}
		return sensitivities, nil
	}
	return sensitivities, fmt.Errorf("Axes sensitivity invalid type: %T", input)
}

type InputHandler struct {
	game             *Game
	handler          *input.Handler
	inputSystem      input.System
	keyboardMouseMap input.Keymap
	gamepadMap       input.Keymap
	sensitivities    *InputSensitivities
}

func NewInputHandler(g *Game) *InputHandler {
	// restore mouse sensitivity from config
	viper := resources.Viper
	mouseSensitivity, err := convertToAxesSensitivity(viper.Get(CONFIG_KEY_CONTROL_MOUSE_SENSITIVITY))
	if err != nil {
		log.Error(err)
	}

	// restore gamepad sensitivity from config
	gamepadSensitivity, err := convertToAxesSensitivity(viper.Get(CONFIG_KEY_CONTROL_GAMEPAD_SENSITIVITY))
	if err != nil {
		log.Error(err)
	}

	h := &InputHandler{
		game: g,
		sensitivities: &InputSensitivities{
			mouse:   mouseSensitivity,
			gamepad: gamepadSensitivity,
		},
	}
	h.inputSystem.Init(input.SystemConfig{
		DevicesEnabled: input.AnyDevice,
	})
	h.handler = h.inputSystem.NewHandler(0, input.Keymap{})
	return h
}

func (h *InputHandler) Update() {
	h.inputSystem.Update()
}

func (h *InputHandler) ActionIsPressed(action input.Action) bool {
	return h.handler.ActionIsPressed(action)
}

func (h *InputHandler) ActionIsJustPressed(action input.Action) bool {
	return h.handler.ActionIsJustPressed(action)
}

func (h *InputHandler) ActionIsJustReleased(action input.Action) bool {
	return h.handler.ActionIsJustReleased(action)
}

func (h *InputHandler) PressedActionInfo(action input.Action) (InputEvent, bool) {
	info, activated := h.handler.PressedActionInfo(action)
	if !activated {
		return InputEvent{EventInfo: info}, activated
	}

	var dX, dY float64
	src := info.Source()
	switch {
	case src.IsMouse():
		dX = -info.DeltaPos.X * h.sensitivities.mouse.X
		dY = -info.DeltaPos.Y * h.sensitivities.mouse.Y
	case src.IsGamepad():
		dX = -info.Pos.X * h.sensitivities.gamepad.X
		dY = -info.Pos.Y * h.sensitivities.gamepad.Y
	}
	return InputEvent{EventInfo: info, Dx: dX, Dy: dY}, activated
}

func (h *InputHandler) JustReleasedActionInfo(action input.Action) (input.EventInfo, bool) {
	return h.handler.JustReleasedActionInfo(action)
}

func (h *InputHandler) handleInput() {
	g := h.game
	menuKeyPressed := h.ActionIsJustPressed(ActionMenuBack)
	if menuKeyPressed {
		if g.menu.Active() {
			if g.osType == osTypeBrowser && inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
				// do not allow Esc key close menu in browser, since Esc key releases browser mouse capture
			} else {
				g.closeMenu()
			}
		} else if !g.InProgress() {
			// instantly leave game when it is over
			g.LeaveGame()
		} else {
			g.openMenu()
		}
	}

	if g.paused {
		return
	}

	h.handleDebugInput()

	_, isInfantry := g.player.Unit.(*model.Infantry)
	//_, isMech := g.player.Unit.(*model.Mech)
	_, isVTOL := g.player.Unit.(*model.VTOL)

	if h.ActionIsJustPressed(ActionPowerToggle) {
		switch g.player.Powered() {
		case model.POWER_ON:
			g.player.SetPowered(model.POWER_OFF_MANUAL)
		case model.POWER_OFF_MANUAL:
			g.player.SetPowered(model.POWER_ON)
		}
	}

	if (g.mouseMode == MouseModeTurret || g.mouseMode == MouseModeBody) && ebiten.CursorMode() != ebiten.CursorModeCaptured {
		ebiten.SetCursorMode(ebiten.CursorModeCaptured)

		// reset initial mouse capture position
		g.mouseX, g.mouseY = math.MinInt32, math.MinInt32
	}

	var moveDx, moveDy float64
	var turretDx, turretDy float64

	if moveAxes, ok := h.PressedActionInfo(ActionMoveAxes); ok {
		moveDx = 10 * -moveAxes.Pos.X
		moveDy = 5 * -moveAxes.Pos.Y
	} // else {
	// TODO: handle mouse mode body
	//}

	if turnAxes, ok := h.PressedActionInfo(ActionTurnAxes); ok {
		moveDx = 10 * -turnAxes.Pos.X
	}
	if throttleAxes, ok := h.PressedActionInfo(ActionThrottleAxes); ok {
		moveDy = 5 * -throttleAxes.Pos.Y
	}

	if moveDx != 0 {
		turnAmount := 0.01 * float64(moveDx) / g.zoomFovDepth
		g.player.SetTargetRelativeHeading(turnAmount)
	} else {
		if !g.player.HasTurret() {
			// reset relative heading target when mouse stops
			g.player.SetTargetRelativeHeading(0)
		}
	}
	// if moveDy != 0 {
	// handled in throttle section below
	// }

	if turretAxes, ok := h.PressedActionInfo(ActionTurretAxes); ok {
		turretDx = turretAxes.Dx
		turretDy = turretAxes.Dy
	}

	if turretDx != 0 {
		if g.player.HasTurret() {
			g.player.RotateCamera(0.005 * turretDx / g.zoomFovDepth)
		} else {
			turnAmount := 0.01 * turretDx / g.zoomFovDepth
			g.player.SetTargetRelativeHeading(turnAmount)
		}
	} else {
		if !g.player.HasTurret() {
			// reset relative heading target when mouse stops
			g.player.SetTargetRelativeHeading(0)
		}
	}
	if turretDy != 0 {
		g.player.PitchCamera(0.005 * turretDy)
	}

	weaponFireGroups := [5]input.Action{
		ActionWeaponFireGroup1,
		ActionWeaponFireGroup2,
		ActionWeaponFireGroup3,
		ActionWeaponFireGroup4,
		ActionWeaponFireGroup5,
	}

	if g.player.Target() == nil {
		// auto-target on crosshairs if just fired weapon without a target selected
		justFired := false
		if h.ActionIsJustPressed(ActionWeaponFire) {
			justFired = true
		} else {
			for _, actionGroup := range weaponFireGroups {
				if h.ActionIsJustPressed(actionGroup) {
					justFired = true
					break
				}
			}
		}

		if justFired {
			targetEntity := g.targetCrosshairs()
			if targetEntity != nil {
				go g.audio.PlayButtonAudio(AUDIO_SELECT_TARGET)
			}
		}
	}

	for i, actionGroup := range weaponFireGroups {
		weaponGroup := i + 1
		if h.ActionIsPressed(actionGroup) {
			g.firePlayerWeapon(weaponGroup)
		}
	}

	if h.ActionIsPressed(ActionWeaponFire) {
		g.firePlayerWeapon(-1)
	}

	isFireButtonJustReleased := h.ActionIsJustReleased(ActionWeaponFire)
	if isFireButtonJustReleased {
		if g.player.fireMode == model.CHAIN_FIRE {
			// cycle to next weapon only in same group (g.player.selectedGroup)
			prevWeapon := g.player.Armament()[g.player.selectedWeapon]
			groupWeapons := g.player.GetWeaponsForGroup(g.player.selectedGroup)

			if len(groupWeapons) == 0 {
				g.player.selectedWeapon = 0
			} else {
				var nextWeapon model.Weapon
				currIndex, nextIndex := 0, 0
				for i, w := range groupWeapons {
					if w == prevWeapon {
						currIndex = i
						nextIndex = i + 1
						if nextIndex >= len(groupWeapons) {
							nextIndex = 0
						}
						break
					}
				}

				// perform weapon check to make sure
				for nextIndex != currIndex {
					nextWeapon = groupWeapons[nextIndex]

					// skip weapon if destroyed, or ammo dependent and has no ammo
					if nextWeapon.Destroyed() || model.WeaponAmmoCount(nextWeapon) == 0 {
						nextIndex += 1
						if nextIndex >= len(groupWeapons) {
							nextIndex = 0
						}
						continue
					}
					// next weapon is ready to cycle
					break
				}

				if nextIndex != currIndex {
					for i, w := range g.player.Armament() {
						if w == nextWeapon {
							g.player.selectedWeapon = uint(i)
							break
						}
					}
				}
			}
		}
	}

	weaponCycleNext, weaponCyclePrev := h.ActionIsJustPressed(ActionWeaponCycle), h.ActionIsJustPressed(ActionWeaponCyclePrevious)
	if weaponCycleNext || weaponCyclePrev {
		playerPrevGroup := g.player.selectedGroup
		playerPrevWeapon := g.player.selectedWeapon

		g.player.CycleWeaponSelection(!weaponCycleNext)

		if playerPrevGroup != g.player.selectedGroup || playerPrevWeapon != g.player.selectedWeapon {
			// play interface sound on weapon/group cycle if changed
			go g.audio.PlayButtonAudio(AUDIO_BUTTON_AFF)
		}
	}

	if h.ActionIsPressed(ActionWeaponGroupSetModifier) {
		// set group for selected weapon
		setGroupIndex := model.WEAPON_GROUP_NONE
		switch {
		case h.ActionIsJustPressed(ActionWeaponGroup1):
			setGroupIndex = model.WEAPON_GROUP_1
		case h.ActionIsJustPressed(ActionWeaponGroup2):
			setGroupIndex = model.WEAPON_GROUP_2
		case h.ActionIsJustPressed(ActionWeaponGroup3):
			setGroupIndex = model.WEAPON_GROUP_3
		case h.ActionIsJustPressed(ActionWeaponGroup4):
			setGroupIndex = model.WEAPON_GROUP_4
		case h.ActionIsJustPressed(ActionWeaponGroup5):
			setGroupIndex = model.WEAPON_GROUP_5
		}

		if setGroupIndex > model.WEAPON_GROUP_NONE {
			addToGroup := true

			weapons := g.player.getSelectedWeapons()

			for _, w := range weapons {
				groups := g.player.GetGroupsForWeapon(w)
				for _, gIndex := range groups {
					if gIndex == setGroupIndex {
						// already in group, remove it
						addToGroup = false
						g.player.weaponGroups = model.RemoveWeaponFromGroup(w, setGroupIndex, g.player.weaponGroups)
						break
					}
				}

				if addToGroup {
					// add to selected group
					g.player.weaponGroups = model.AddWeaponToGroup(w, setGroupIndex, g.player.weaponGroups)
				}
			}
			g.player.selectedGroup = setGroupIndex

			// TODO: use background thread queue to avoid multiple writes at same time
			setUnitWeaponGroups(g.player, g.player.weaponGroups)
			if err := saveUserWeaponGroups(); err != nil {
				log.Error("failed to save user weapon groups: " + err.Error())
			}

			go g.audio.PlayButtonAudio(AUDIO_BUTTON_OVER)
		}
	} else {
		// set currently selected weapon/group if weapon group number key pressed
		selectGroupIndex := model.WEAPON_GROUP_NONE
		switch {
		case h.ActionIsJustPressed(ActionWeaponGroup1):
			selectGroupIndex = model.WEAPON_GROUP_1
		case h.ActionIsJustPressed(ActionWeaponGroup2):
			selectGroupIndex = model.WEAPON_GROUP_2
		case h.ActionIsJustPressed(ActionWeaponGroup3):
			selectGroupIndex = model.WEAPON_GROUP_3
		case h.ActionIsJustPressed(ActionWeaponGroup4):
			selectGroupIndex = model.WEAPON_GROUP_4
		case h.ActionIsJustPressed(ActionWeaponGroup5):
			selectGroupIndex = model.WEAPON_GROUP_5
		}

		if selectGroupIndex > model.WEAPON_GROUP_NONE {
			weapons := g.player.weaponGroups[selectGroupIndex]
			if len(weapons) == 0 {
				go g.audio.PlayButtonAudio(AUDIO_BUTTON_NEG)
			} else {
				for i, w := range g.player.Armament() {
					if w == weapons[0] {
						g.player.selectedGroup = selectGroupIndex
						g.player.selectedWeapon = uint(i)
						go g.audio.PlayButtonAudio(AUDIO_BUTTON_AFF)
						break
					}
				}
			}
		}
	}

	if h.ActionIsJustPressed(ActionWeaponGroupFireToggle) {
		// toggle group fire mode
		if g.player.fireMode == model.CHAIN_FIRE {
			g.player.fireMode = model.GROUP_FIRE
		} else {
			g.player.fireMode = model.CHAIN_FIRE
		}

		switch g.player.fireMode {
		case model.GROUP_FIRE:
			// select the appropriate group from selected weapon when switching to group mode
			prevSelectedWeapon := g.player.Armament()[g.player.selectedWeapon]
			groups := g.player.GetGroupsForWeapon(prevSelectedWeapon)
			if len(groups) == 0 {
				g.player.selectedGroup = model.WEAPON_GROUP_NONE
			} else if !slices.Contains(groups, g.player.selectedGroup) {
				g.player.selectedGroup = groups[0]
			}
		case model.CHAIN_FIRE:
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

	if h.ActionIsJustPressed(ActionNavCycle) {
		// cycle nav points
		g.navPointCycle(true)
	}

	if h.ActionIsJustPressed(ActionRadarRangeCycle) {
		// cycle radar HUD range
		g.cycleRadarRange()
	}

	if h.ActionIsJustPressed(ActionTargetCrosshairs) {
		// target on crosshairs
		targetEntity := g.targetCrosshairs()
		if targetEntity != nil {
			go g.audio.PlayButtonAudio(AUDIO_SELECT_TARGET)
		}
	}

	if h.ActionIsJustPressed(ActionTargetNearest) {
		// target nearest to player
		targetEntity := g.targetCycle(TARGET_NEAREST)
		if targetEntity != nil {
			go g.audio.PlayButtonAudio(AUDIO_SELECT_TARGET)
		}
	}

	if h.ActionIsJustPressed(ActionTargetNext) {
		// cycle player targets
		targetEntity := g.targetCycle(TARGET_NEXT)
		if targetEntity != nil {
			go g.audio.PlayButtonAudio(AUDIO_SELECT_TARGET)
		}
	}

	if h.ActionIsJustPressed(ActionTargetPrevious) {
		// cycle player targets in reverse order
		targetEntity := g.targetCycle(TARGET_PREVIOUS)
		if targetEntity != nil {
			go g.audio.PlayButtonAudio(AUDIO_SELECT_TARGET)
		}
	}

	switch {
	case h.ActionIsJustPressed(ActionZoomToggle):
		g.zoomToggle()
	case h.ActionIsJustPressed(ActionZoomIn):
		g.zoomIn()
	case h.ActionIsJustPressed(ActionZoomOut):
		g.zoomOut()
	}

	if h.ActionIsJustPressed(ActionLightAmpToggle) {
		// toggle light amplification
		if g.lightAmpEngaged {
			// disable light amplification
			g.lightAmpEngaged = false
			g.camera.SetLightFalloff(g.lightFalloff)
			g.camera.SetGlobalIllumination(g.globalIllumination)
			g.camera.SetLightRGB(*g.minLightRGB, *g.maxLightRGB)
		} else {
			// enable light amplification
			g.lightAmpEngaged = true
			g.camera.SetLightFalloff(-128)
			g.camera.SetGlobalIllumination(300)
			g.camera.SetLightRGB(
				color.NRGBA{R: 0, G: 24, B: 0},
				color.NRGBA{R: 16, G: 128, B: 16},
			)
		}

		g.audio.PlayButtonAudio(AUDIO_CLICK_AFF)
	}

	if h.ActionIsJustPressed(ActionThrottleReverse) {
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

	if h.ActionIsPressed(ActionJumpJet) {
		switch {
		case isVTOL:
			// TODO: use unit tonnage and gravity to determine ascent speed
			g.player.SetTargetVelocityZ(g.player.MaxVelocity() / 2)
		default:
			initJumping := !g.player.JumpJetsActive()
			canJumpJet := g.player.JumpJets() > 0 && g.player.JumpJetDuration() < g.player.MaxJumpJetDuration()
			if canJumpJet {
				g.player.SetJumpJetsActive(true)
				g.player.SetJumpJetsDirectional(false)
				if initJumping {
					// initialize jump jet heading if first update with jets active
					g.player.SetJumpJetHeading(g.player.Heading())
				}
			}
		}

	} else if g.player.JumpJetsActive() {
		// reset jump jet active status
		g.player.SetJumpJetsActive(false)

	} else if h.ActionIsPressed(ActionDescend) {
		if isVTOL {
			// TODO: use unit tonnage and gravity to determine descent speed
			g.player.SetTargetVelocityZ(-g.player.MaxVelocity() / 2)
		}
	}

	var forward, backward bool
	var throttlePercent float64 = -math.MaxFloat64
	var rotLeft, rotRight bool
	var lookUp, lookDown, lookLeft, lookRight bool

	if h.ActionIsPressed(ActionTurretLeft) {
		lookLeft = true
	} else if h.ActionIsPressed(ActionTurretRight) {
		lookRight = true
	}

	if h.ActionIsPressed(ActionTurretUp) {
		lookUp = true
	} else if h.ActionIsPressed(ActionTurretDown) {
		lookDown = true
	}

	if h.ActionIsPressed(ActionLeft) {
		rotLeft = true
	}
	if h.ActionIsPressed(ActionRight) {
		rotRight = true
	}

	if h.ActionIsPressed(ActionUp) || moveDy >= 0.2 {
		forward = true
	}
	if h.ActionIsPressed(ActionDown) || moveDy <= -0.2 {
		backward = true
	}

	switch {
	case h.ActionIsPressed(ActionThrottle0):
		throttlePercent = 0
	case h.ActionIsPressed(ActionThrottle10):
		throttlePercent = 0.1
	case h.ActionIsPressed(ActionThrottle20):
		throttlePercent = 0.2
	case h.ActionIsPressed(ActionThrottle30):
		throttlePercent = 0.3
	case h.ActionIsPressed(ActionThrottle40):
		throttlePercent = 0.4
	case h.ActionIsPressed(ActionThrottle50):
		throttlePercent = 0.5
	case h.ActionIsPressed(ActionThrottle60):
		throttlePercent = 0.6
	case h.ActionIsPressed(ActionThrottle70):
		throttlePercent = 0.7
	case h.ActionIsPressed(ActionThrottle80):
		throttlePercent = 0.8
	case h.ActionIsPressed(ActionThrottle90):
		throttlePercent = 0.9
	case h.ActionIsPressed(ActionThrottle100):
		throttlePercent = 1.0
	}

	switch {
	case h.ActionIsPressed(ActionJumpJet) && (forward || backward || rotLeft || rotRight):
		// set jump jets as directional with desired heading
		if g.player.JumpJetsActive() {
			jumpJetHeading := g.player.cameraAngle
			if backward {
				// set reverse directional jump jet heading
				jumpJetHeading -= geom.Pi
			}
			switch {
			case rotLeft:
				// set left directional jump jet heading
				leftHeading := geom.HalfPi
				if backward {
					leftHeading = -leftHeading
				}
				jumpJetHeading += leftHeading
			case rotRight:
				// set right directional jump jet heading
				rightHeading := -geom.HalfPi
				if backward {
					rightHeading = -rightHeading
				}
				jumpJetHeading += rightHeading
			}
			g.player.SetJumpJetsDirectional(true)
			g.player.SetJumpJetHeading(model.ClampAngle2Pi(jumpJetHeading))
		}

		// disable directional input from further processing outside directional jump jet handling
		forward, backward, rotLeft, rotRight = false, false, false, false

	case g.throttleDecay:
		if forward {
			g.player.SetTargetVelocity(g.player.MaxVelocity())
		} else if backward {
			g.player.SetTargetVelocity(-g.player.MaxVelocity() / 2)
		} else {
			g.player.SetTargetVelocity(0)
		}

	case !g.throttleDecay:
		deltaV := 0.0004 // FIXME: something else not randomly hardcoded?
		if math.Abs(moveDy) >= 0.2 {
			deltaV *= math.Abs(moveDy)
		}
		if throttlePercent >= 0 {
			g.player.SetTargetVelocity(throttlePercent * g.player.MaxVelocity())
		} else if forward {
			g.player.SetTargetVelocity(g.player.TargetVelocity() + deltaV)
		} else if backward {
			g.player.SetTargetVelocity(g.player.TargetVelocity() - deltaV)
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

	if lookUp {
		// TODO: better and configurable values for dx/dy
		dy := 2.0
		g.player.PitchCamera(0.005 * dy)
	} else if lookDown {
		dy := -2.0
		g.player.PitchCamera(0.005 * dy)
	}
	if lookLeft {
		dx := 5.0
		g.player.RotateCamera(0.005 * dx / g.zoomFovDepth)
	} else if lookRight {
		dx := -5.0
		g.player.RotateCamera(0.005 * dx / g.zoomFovDepth)
	}

	if isStrafe {
		// TODO: use unit max velocity to determine strafe speed and set target strafe heading
		// if rotLeft {
		// 	g.Strafe(-0.05)
		// } else if rotRight {
		// 	g.Strafe(0.05)
		// }
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

// debug mode only input flags
var debugProfCPU bool

func (h *InputHandler) handleDebugInput() {
	g := h.game
	if !g.debug {
		return
	}

	ctrl_test := ebiten.IsKeyPressed(ebiten.KeyControl)
	alt_test := ebiten.IsKeyPressed(ebiten.KeyAlt)

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
		// TESTING purposes only
		switch {
		case ctrl_test && alt_test:
			destroyEntity(g.player)

		case ctrl_test:
			target := model.EntityUnit(g.player.Target())
			if target != nil {
				destroyEntity(target)
			}

		case alt_test:
			target := model.EntityUnit(g.player.Target())
			if target != nil && target.JumpJets() > 0 {
				target.SetJumpJetsActive(true)
				target.SetTargetVelocityZ(0.05)
			}
		}
	}

	if ctrl_test && alt_test && h.ActionIsJustPressed(ActionCameraCycle) {
		// debug only: start/stop CPU profiler
		if debugProfCPU {
			pprof.StopCPUProfile()
			debugProfCPU = false
		} else {
			debugProfFile, _ = os.Create("cpu_" + strconv.Itoa(os.Getpid()) + ".prof")
			pprof.StartCPUProfile(debugProfFile)
			debugProfCPU = true
		}
	} else if h.ActionIsJustPressed(ActionCameraCycle) {
		// debug only: camera swap with player target or cycle back to player unit
		debugCamTgt := g.player.DebugCameraTarget()
		if debugCamTgt == nil && g.player.Target() != nil {
			g.player.SetDebugCameraTarget(model.EntityUnit(g.player.Target()))
		} else if debugCamTgt != nil {
			g.player.SetDebugCameraTarget(nil)
			g.player.moved = true
		}
	}
}
