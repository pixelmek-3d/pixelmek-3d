package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
	input "github.com/quasilyte/ebitengine-input"
	log "github.com/sirupsen/logrus"
	orderedmap "github.com/wk8/go-ordered-map/v2"
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
	actionToString map[input.Action]string
	stringToAction map[string]input.Action
)

func init() {
	actionToString = map[input.Action]string{
		ActionUp:                     "up",
		ActionDown:                   "down",
		ActionLeft:                   "left",
		ActionRight:                  "right",
		ActionMoveAxes:               "move_axes",
		ActionTurretUp:               "turret_up",
		ActionTurretDown:             "turret_down",
		ActionTurretLeft:             "turret_left",
		ActionTurretRight:            "turret_right",
		ActionTurretAxes:             "turret_axes",
		ActionMenu:                   "menu",
		ActionBack:                   "back",
		ActionThrottleReverse:        "throttle_reverse",
		ActionThrottle0:              "throttle_0",
		ActionJumpJet:                "jump_jet",
		ActionDescend:                "descend",
		ActionWeaponFire:             "weapon_fire",
		ActionWeaponCycle:            "weapon_cycle",
		ActionWeaponGroupFireToggle:  "weapon_group_toggle",
		ActionWeaponGroupSetModifier: "weapon_group_set",
		ActionWeaponGroup1:           "weapon_group_1",
		ActionWeaponGroup2:           "weapon_group_2",
		ActionWeaponGroup3:           "weapon_group_3",
		ActionWeaponGroup4:           "weapon_group_4",
		ActionWeaponGroup5:           "weapon_group_5",
		ActionWeaponFireGroup1:       "weapon_fire_group_1",
		ActionWeaponFireGroup2:       "weapon_fire_group_2",
		ActionWeaponFireGroup3:       "weapon_fire_group_3",
		ActionWeaponFireGroup4:       "weapon_fire_group_4",
		ActionWeaponFireGroup5:       "weapon_fire_group_5",
		ActionNavCycle:               "nav_cycle",
		ActionRadarRangeCycle:        "radar_range_cycle",
		ActionTargetCrosshairs:       "target_crosshairs",
		ActionTargetNearest:          "target_nearest",
		ActionTargetNext:             "target_next",
		ActionTargetPrevious:         "target_prev",
		ActionZoomToggle:             "zoom_toggle",
		ActionLightAmpToggle:         "light_amplification",
		ActionPowerToggle:            "power_toggle",
		ActionCameraCycle:            "camera_cycle",
	}

	// Build a reverse index to get an action by its name
	stringToAction = make(map[string]input.Action, len(actionToString))
	for a := ActionUnknown + 1; a < actionCount; a++ {
		stringToAction[actionString(a)] = a
	}
}

func stringAction(aName string) input.Action {
	a, ok := stringToAction[aName]
	if !ok {
		return ActionUnknown
	}
	return a
}

func actionString(a input.Action) string {
	if s, ok := actionToString[a]; ok {
		return s
	}
	panic(fmt.Errorf("currently unable to handle actionString for input.Action: %v", a))
}

func (g *Game) initControls() {
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
