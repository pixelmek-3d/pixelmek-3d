package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
	input "github.com/quasilyte/ebitengine-input"
	log "github.com/sirupsen/logrus"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type MouseMode int

const (
	MouseModeTurret MouseMode = iota
	MouseModeBody
	MouseModeCursor
)

const (
	ActionUnknown input.Action = iota
	ActionUp
	ActionDown
	ActionLeft
	ActionRight
	ActionMoveAxes
	ActionTurretUp
	ActionTurretDown
	ActionTurretLeft
	ActionTurretRight
	ActionTurretAxes
	ActionMenu
	ActionBack
	ActionThrottleReverse
	ActionThrottle0
	ActionJumpJet
	ActionDescend
	ActionWeaponFire
	ActionWeaponCycle
	ActionWeaponGroupFireToggle
	ActionWeaponGroupSetModifier
	ActionWeaponGroup1
	ActionWeaponGroup2
	ActionWeaponGroup3
	ActionWeaponGroup4
	ActionWeaponGroup5
	ActionWeaponFireGroup1
	ActionWeaponFireGroup2
	ActionWeaponFireGroup3
	ActionWeaponFireGroup4
	ActionWeaponFireGroup5
	ActionNavCycle
	ActionRadarRangeCycle
	ActionTargetCrosshairs
	ActionTargetNearest
	ActionTargetNext
	ActionTargetPrevious
	ActionZoomToggle
	ActionLightAmpToggle
	ActionPowerToggle
	ActionCameraCycle
	actionCount
)

var (
	stringToAction map[string]input.Action

	debugProfFile *os.File
)

func stringAction(aName string) input.Action {
	a, ok := stringToAction[aName]
	if !ok {
		return ActionUnknown
	}
	return a
}

func actionString(a input.Action) string {
	switch a {
	case ActionUp:
		return "up"
	case ActionDown:
		return "down"
	case ActionLeft:
		return "left"
	case ActionRight:
		return "right"
	case ActionMoveAxes:
		return "move_axes"
	case ActionTurretUp:
		return "turret_up"
	case ActionTurretDown:
		return "turret_down"
	case ActionTurretLeft:
		return "turret_left"
	case ActionTurretRight:
		return "turret_right"
	case ActionTurretAxes:
		return "turret_axes"
	case ActionMenu:
		return "menu"
	case ActionBack:
		return "back"
	case ActionThrottleReverse:
		return "throttle_reverse"
	case ActionThrottle0:
		return "throttle_0"
	case ActionJumpJet:
		return "jump_jet"
	case ActionDescend:
		return "descend"
	case ActionWeaponFire:
		return "weapon_fire"
	case ActionWeaponCycle:
		return "weapon_cycle"
	case ActionWeaponGroupFireToggle:
		return "weapon_group_toggle"
	case ActionWeaponGroupSetModifier:
		return "weapon_group_set"
	case ActionWeaponGroup1:
		return "weapon_group_1"
	case ActionWeaponGroup2:
		return "weapon_group_2"
	case ActionWeaponGroup3:
		return "weapon_group_3"
	case ActionWeaponGroup4:
		return "weapon_group_4"
	case ActionWeaponGroup5:
		return "weapon_group_5"
	case ActionWeaponFireGroup1:
		return "weapon_fire_group_1"
	case ActionWeaponFireGroup2:
		return "weapon_fire_group_2"
	case ActionWeaponFireGroup3:
		return "weapon_fire_group_3"
	case ActionWeaponFireGroup4:
		return "weapon_fire_group_4"
	case ActionWeaponFireGroup5:
		return "weapon_fire_group_5"
	case ActionNavCycle:
		return "nav_cycle"
	case ActionRadarRangeCycle:
		return "radar_range_cycle"
	case ActionTargetCrosshairs:
		return "target_crosshairs"
	case ActionTargetNearest:
		return "target_nearest"
	case ActionTargetNext:
		return "target_next"
	case ActionTargetPrevious:
		return "target_prev"
	case ActionZoomToggle:
		return "zoom_toggle"
	case ActionLightAmpToggle:
		return "light_amplification"
	case ActionPowerToggle:
		return "power_toggle"
	case ActionCameraCycle:
		return "camera_cycle"
	default:
		panic(fmt.Errorf("currently unable to handle actionString for input.Action: %v", a))
	}
}

func (g *Game) initControls() {
	// Build a reverse index to get an action by its name
	stringToAction = map[string]input.Action{}
	for a := ActionUnknown + 1; a < actionCount; a++ {
		stringToAction[actionString(a)] = a
	}

	// import from keymap file if exists
	var keymap input.Keymap
	if _, err := os.Stat(resources.UserKeymapFile); err == nil {
		keymap, err = g.restoreControls()
		if err != nil {
			panic(fmt.Errorf("error loading keymap file %s: %v", resources.UserKeymapFile, err))
		}
	}

	// TODO: initialize default for new controls even if not first time?

	if len(keymap) == 0 {
		// first time intitialize defaults into file
		g.setDefaultControls()
		g.saveControls()
	}
}

func (g *Game) setDefaultControls() {
	keymap := input.Keymap{
		ActionUp:       {input.KeyW, input.KeyUp},
		ActionDown:     {input.KeyS, input.KeyDown},
		ActionLeft:     {input.KeyA, input.KeyLeft},
		ActionRight:    {input.KeyD, input.KeyRight},
		ActionMoveAxes: {input.KeyGamepadLStickMotion},

		ActionTurretUp:    {},
		ActionTurretDown:  {},
		ActionTurretLeft:  {},
		ActionTurretRight: {},
		ActionTurretAxes:  {input.KeyGamepadRStickMotion},

		ActionMenu: {input.KeyEscape, input.KeyF1, input.KeyGamepadStart},
		ActionBack: {input.KeyEscape, input.KeyGamepadBack},

		ActionThrottleReverse: {input.KeyBackspace},
		ActionThrottle0:       {input.KeyX},
		ActionJumpJet:         {input.KeySpace, input.KeyGamepadLStick},
		ActionDescend:         {input.KeyControl},

		ActionWeaponFire:             {input.KeyMouseLeft, input.KeyGamepadR2},
		ActionWeaponCycle:            {input.KeyMouseRight, input.KeyGamepadR1},
		ActionWeaponGroupFireToggle:  {input.KeyBackslash, input.KeyGamepadY},
		ActionWeaponGroupSetModifier: {input.KeyShift},
		ActionWeaponGroup1:           {input.Key1},
		ActionWeaponGroup2:           {input.Key2},
		ActionWeaponGroup3:           {input.Key3},
		ActionWeaponGroup4:           {input.Key4},
		ActionWeaponGroup5:           {input.Key5},
		ActionWeaponFireGroup1:       {input.KeyMouseBack},
		ActionWeaponFireGroup2:       {input.KeyMouseForward},

		ActionNavCycle:         {input.KeyN, input.KeyGamepadDown},
		ActionRadarRangeCycle:  {input.KeySlash},
		ActionTargetCrosshairs: {input.KeyQ, input.KeyGamepadL2},
		ActionTargetNearest:    {input.KeyE, input.KeyGamepadUp},
		ActionTargetNext:       {input.KeyT, input.KeyGamepadRight},
		ActionTargetPrevious:   {input.KeyR, input.KeyGamepadLeft},

		ActionZoomToggle:     {input.KeyZ, input.KeyGamepadRStick},
		ActionLightAmpToggle: {input.KeyL, input.KeyGamepadDown},
		ActionPowerToggle:    {input.KeyP},
		ActionCameraCycle:    {input.KeyF3},
	}

	g.inputSystem.Init(input.SystemConfig{
		DevicesEnabled: input.AnyDevice,
	})
	g.input = g.inputSystem.NewHandler(0, keymap)
}

func (g *Game) restoreControls() (input.Keymap, error) {
	log.Debug("restoring keymap file ", resources.UserKeymapFile)
	keymap := input.Keymap{}

	keymapFile, err := os.Open(resources.UserKeymapFile)
	if err != nil {
		log.Error(err)
		return keymap, err
	}
	defer keymapFile.Close()

	fileBytes, err := io.ReadAll(keymapFile)
	if err != nil {
		log.Error(err)
		return keymap, err
	}

	if len(fileBytes) == 0 {
		// caller expected to handle empty keymap without error
		return keymap, nil
	}

	var keymapConfig map[string][]string
	err = json.Unmarshal(fileBytes, &keymapConfig)
	if err != nil {
		log.Error(err)
		return keymap, err
	}

	// Parse our config file into a keymap object.
	var actionErrorString string
	var actionWarningString string

	for actionName, keyNames := range keymapConfig {
		a := stringAction(actionName)
		if a == ActionUnknown {
			actionWarningString += fmt.Sprintf("unexpected action name: %s\n", actionName)
		}
		keys := make([]input.Key, len(keyNames))
		for i, keyString := range keyNames {
			k, err := input.ParseKey(keyString)
			if err != nil {
				actionErrorString += err.Error() + "\n"
			}
			keys[i] = k
		}
		keymap[a] = keys
	}

	if len(actionWarningString) > 0 {
		log.Warning(actionWarningString)
	}

	if len(actionErrorString) > 0 {
		err = errors.New(actionErrorString)
		log.Error(err)
		return keymap, err
	}

	g.inputSystem.Init(input.SystemConfig{
		DevicesEnabled: input.AnyDevice,
	})
	g.input = g.inputSystem.NewHandler(0, keymap)

	return keymap, nil
}

func (g *Game) saveControls() error {
	log.Debug("saving keymap file ", resources.UserKeymapFile)

	userConfigPath := filepath.Dir(resources.UserKeymapFile)
	if _, err := os.Stat(userConfigPath); os.IsNotExist(err) {
		err = os.MkdirAll(userConfigPath, os.ModePerm)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	keymapFile, err := os.Create(resources.UserKeymapFile)
	if err != nil {
		log.Error(err)
		return err
	}
	defer keymapFile.Close()

	// first time intitialize defaults into file
	keymapConfig := orderedmap.New[string, []string]()
	for a := ActionUnknown + 1; a < actionCount; a++ {
		actionKey := actionString(a)
		keymapConfig.Set(actionKey, g.input.ActionKeyNames(a, input.AnyDevice))
	}
	keymapJson, _ := json.MarshalIndent(keymapConfig, "", "    ")
	_, err = keymapFile.Write(keymapJson)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (g *Game) handleInput() {
	menuKeyPressed := g.input.ActionIsJustPressed(ActionMenu)
	if menuKeyPressed {
		if g.menu.Active() {
			if g.osType == osTypeBrowser && inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
				// do not allow Esc key close menu in browser, since Esc key releases browser mouse capture
			} else {
				g.closeMenu()
			}
		} else if !g.InProgress() {
			// instantly leave game when it is over
			g.paused = true
			g.LeaveGame()
		} else {
			g.openMenu()
		}
	}

	if g.paused {
		return
	}

	g.handleDebugInput()

	_, isInfantry := g.player.Unit.(*model.Infantry)
	//_, isMech := g.player.Unit.(*model.Mech)
	_, isVTOL := g.player.Unit.(*model.VTOL)

	if g.input.ActionIsJustPressed(ActionPowerToggle) {
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
	cursorX, cursorY := ebiten.CursorPosition()

	if moveAxes, ok := g.input.PressedActionInfo(ActionMoveAxes); ok {
		// TODO: configurable deadzone and sensitivity (for mouse and gamepad)
		if math.Abs(moveAxes.Pos.X) >= 0.2 {
			moveDx = 10 * -moveAxes.Pos.X
		}
		if math.Abs(moveAxes.Pos.Y) >= 0.2 {
			moveDy = 5 * -moveAxes.Pos.Y
		}
	} // else {
	// TODO: handle mouse mode body
	//}

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

	if turretAxes, ok := g.input.PressedActionInfo(ActionTurretAxes); ok {
		// TODO: configurable deadzone and sensitivity (for mouse and gamepad)
		if math.Abs(turretAxes.Pos.X) >= 0.2 {
			turretDx = 10 * -turretAxes.Pos.X
		}
		if math.Abs(turretAxes.Pos.Y) >= 0.2 {
			turretDy = 5 * -turretAxes.Pos.Y
		}
	} else {
		// handle mouse mode turret
		switch {
		case g.mouseX == math.MinInt32 && g.mouseY == math.MinInt32:
			// initialize first position to establish delta
			if cursorX != 0 && cursorY != 0 {
				g.mouseX, g.mouseY = cursorX, cursorY
			}

		default:
			turretDx, turretDy = float64(g.mouseX-cursorX), float64(g.mouseY-cursorY)
			g.mouseX, g.mouseY = cursorX, cursorY
		}
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
		if g.input.ActionIsJustPressed(ActionWeaponFire) {
			justFired = true
		} else {
			for _, actionGroup := range weaponFireGroups {
				if g.input.ActionIsJustPressed(actionGroup) {
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

	for weaponGroup, actionGroup := range weaponFireGroups {
		if g.input.ActionIsPressed(actionGroup) {
			g.firePlayerWeapon(weaponGroup)
		}
	}

	if g.input.ActionIsPressed(ActionWeaponFire) {
		g.firePlayerWeapon(-1)
	}

	isFireButtonJustReleased := g.input.ActionIsJustReleased(ActionWeaponFire)
	if isFireButtonJustReleased {
		if g.player.fireMode == model.CHAIN_FIRE {
			// cycle to next weapon only in same group (g.player.selectedGroup)
			prevWeapon := g.player.Armament()[g.player.selectedWeapon]
			groupWeapons := g.player.weaponGroups[g.player.selectedGroup]

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

	if g.input.ActionIsJustPressed(ActionWeaponCycle) {
		playerPrevGroup := g.player.selectedGroup
		playerPrevWeapon := g.player.selectedWeapon

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

		if playerPrevGroup != g.player.selectedGroup || playerPrevWeapon != g.player.selectedWeapon {
			// play interface sound on weapon/group cycle
			go g.audio.PlayButtonAudio(AUDIO_BUTTON_AFF)
		}
	}

	if g.input.ActionIsPressed(ActionWeaponGroupSetModifier) {
		if g.player.fireMode == model.CHAIN_FIRE {
			// set group for selected weapon
			setGroupIndex := -1
			switch {
			case g.input.ActionIsJustPressed(ActionWeaponGroup1):
				setGroupIndex = 0
			case g.input.ActionIsJustPressed(ActionWeaponGroup2):
				setGroupIndex = 1
			case g.input.ActionIsJustPressed(ActionWeaponGroup3):
				setGroupIndex = 2
			case g.input.ActionIsJustPressed(ActionWeaponGroup4):
				setGroupIndex = 3
			case g.input.ActionIsJustPressed(ActionWeaponGroup5):
				setGroupIndex = 4
			}

			if setGroupIndex >= 0 {
				addToGroup := true
				weapon := g.player.Armament()[g.player.selectedWeapon]
				groups := model.GetGroupsForWeapon(weapon, g.player.weaponGroups)
				for _, gIndex := range groups {
					if int(gIndex) == setGroupIndex {
						// already in group
						addToGroup = false
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

				if addToGroup {
					// add to selected group
					g.player.weaponGroups[setGroupIndex] = append(g.player.weaponGroups[setGroupIndex], weapon)
				}
				g.player.selectedGroup = uint(setGroupIndex)

				go g.audio.PlayButtonAudio(AUDIO_BUTTON_OVER)
			}
		}
	} else {
		// set currently selected weapon/group if weapon group number key pressed
		selectGroupIndex := -1
		switch {
		case g.input.ActionIsJustPressed(ActionWeaponGroup1):
			selectGroupIndex = 0
		case g.input.ActionIsJustPressed(ActionWeaponGroup2):
			selectGroupIndex = 1
		case g.input.ActionIsJustPressed(ActionWeaponGroup3):
			selectGroupIndex = 2
		case g.input.ActionIsJustPressed(ActionWeaponGroup4):
			selectGroupIndex = 3
		case g.input.ActionIsJustPressed(ActionWeaponGroup5):
			selectGroupIndex = 4
		}

		if selectGroupIndex >= 0 {
			g.player.selectedGroup = uint(selectGroupIndex)
			weapons := g.player.weaponGroups[selectGroupIndex]
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

			go g.audio.PlayButtonAudio(AUDIO_BUTTON_AFF)
		}
	}

	if g.input.ActionIsJustPressed(ActionWeaponGroupFireToggle) {
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

	if g.input.ActionIsJustPressed(ActionNavCycle) {
		// cycle nav points
		g.navPointCycle(true)
	}

	if g.input.ActionIsJustPressed(ActionRadarRangeCycle) {
		// cycle radar HUD range
		g.cycleRadarRange()
	}

	if g.input.ActionIsJustPressed(ActionTargetCrosshairs) {
		// target on crosshairs
		targetEntity := g.targetCrosshairs()
		if targetEntity != nil {
			go g.audio.PlayButtonAudio(AUDIO_SELECT_TARGET)
		}
	}

	if g.input.ActionIsJustPressed(ActionTargetNearest) {
		// target nearest to player
		targetEntity := g.targetCycle(TARGET_NEAREST)
		if targetEntity != nil {
			go g.audio.PlayButtonAudio(AUDIO_SELECT_TARGET)
		}
	}

	if g.input.ActionIsJustPressed(ActionTargetNext) {
		// cycle player targets
		targetEntity := g.targetCycle(TARGET_NEXT)
		if targetEntity != nil {
			go g.audio.PlayButtonAudio(AUDIO_SELECT_TARGET)
		}
	}

	if g.input.ActionIsJustPressed(ActionTargetPrevious) {
		// cycle player targets in reverse order
		targetEntity := g.targetCycle(TARGET_PREVIOUS)
		if targetEntity != nil {
			go g.audio.PlayButtonAudio(AUDIO_SELECT_TARGET)
		}
	}

	if g.input.ActionIsJustPressed(ActionZoomToggle) {
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

	if g.input.ActionIsJustPressed(ActionLightAmpToggle) {
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

	if g.input.ActionIsJustPressed(ActionThrottleReverse) {
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

	if g.input.ActionIsPressed(ActionJumpJet) {
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
		// TODO: infantry jump (or jump jet infantry)

	} else if g.player.JumpJetsActive() {
		// reset jump jet active status
		g.player.SetJumpJetsActive(false)

	} else if g.input.ActionIsPressed(ActionDescend) {
		if isVTOL {
			// TODO: use unit tonnage and gravity to determine descent speed
			g.player.SetTargetVelocityZ(-g.player.MaxVelocity() / 2)
		}
	}

	var stop, forward, backward bool
	var rotLeft, rotRight bool
	var lookUp, lookDown, lookLeft, lookRight bool

	if g.input.ActionIsPressed(ActionTurretLeft) {
		lookLeft = true
	} else if g.input.ActionIsPressed(ActionTurretRight) {
		lookRight = true
	}

	if g.input.ActionIsPressed(ActionTurretUp) {
		lookUp = true
	} else if g.input.ActionIsPressed(ActionTurretDown) {
		lookDown = true
	}

	if g.input.ActionIsPressed(ActionLeft) {
		rotLeft = true
	}
	if g.input.ActionIsPressed(ActionRight) {
		rotRight = true
	}

	if g.input.ActionIsPressed(ActionUp) || moveDy >= 0.2 {
		forward = true
	}
	if g.input.ActionIsPressed(ActionDown) || moveDy <= -0.2 {
		backward = true
	}

	if g.input.ActionIsPressed(ActionThrottle0) {
		stop = true
	}

	switch {
	case g.player.JumpJetsActive() && (forward || backward):
		if forward {
			// set forward directional jump jet heading
			g.player.SetJumpJetsDirectional(true)
			g.player.SetJumpJetHeading(g.player.cameraAngle)
		} else if backward {
			// set reverse directional jump jet heading
			g.player.SetJumpJetsDirectional(true)
			g.player.SetJumpJetHeading(model.ClampAngle2Pi(g.player.cameraAngle - geom.Pi))

		}

	case g.throttleDecay:
		if forward {
			g.player.SetTargetVelocity(g.player.MaxVelocity())
		} else if backward {
			g.player.SetTargetVelocity(-g.player.MaxVelocity() / 2)
		} else {
			g.player.SetTargetVelocity(0)
		}

	case !g.throttleDecay:
		deltaV := 0.0004 // FIXME: testing
		if math.Abs(moveDy) >= 0.2 {
			deltaV *= math.Abs(moveDy)
		}
		if stop {
			g.player.SetTargetVelocity(0)
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

func (g *Game) handleDebugInput() {
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

	if ctrl_test && alt_test && g.input.ActionIsJustPressed(ActionCameraCycle) {
		// debug only: start/stop CPU profiler
		if debugProfCPU {
			pprof.StopCPUProfile()
			debugProfCPU = false
		} else {
			debugProfFile, _ = os.Create("cpu_" + strconv.Itoa(os.Getpid()) + ".prof")
			pprof.StartCPUProfile(debugProfFile)
			debugProfCPU = true
		}
	} else if g.input.ActionIsJustPressed(ActionCameraCycle) {
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
