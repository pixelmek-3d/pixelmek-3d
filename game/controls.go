package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
	input "github.com/quasilyte/ebitengine-input"
	log "github.com/sirupsen/logrus"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type KeymapType int

const (
	KeymapTypeKeyboardMouse KeymapType = iota
	KeymapTypeGamepad
)

const (
	ActionUnknown input.Action = iota
	ActionMoveAxes
	ActionTurnAxes
	ActionUp
	ActionDown
	ActionLeft
	ActionRight
	ActionTurretAxes
	ActionTurretUp
	ActionTurretDown
	ActionTurretLeft
	ActionTurretRight
	ActionMenuBack
	ActionThrottleAxes
	ActionThrottleReverse
	ActionThrottle0
	ActionThrottle10
	ActionThrottle20
	ActionThrottle30
	ActionThrottle40
	ActionThrottle50
	ActionThrottle60
	ActionThrottle70
	ActionThrottle80
	ActionThrottle90
	ActionThrottle100
	ActionJumpJet
	ActionDescend
	ActionWeaponFire
	ActionWeaponCycle
	ActionWeaponCyclePrevious
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
	ActionZoomIn
	ActionZoomOut
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
		ActionTurnAxes:               "turn_axes",
		ActionThrottleAxes:           "throttle_axes",
		ActionTurretUp:               "turret_up",
		ActionTurretDown:             "turret_down",
		ActionTurretLeft:             "turret_left",
		ActionTurretRight:            "turret_right",
		ActionTurretAxes:             "turret_axes",
		ActionMenuBack:               "menu_back",
		ActionThrottleReverse:        "throttle_reverse",
		ActionThrottle0:              "throttle_0",
		ActionThrottle10:             "throttle_10",
		ActionThrottle20:             "throttle_20",
		ActionThrottle30:             "throttle_30",
		ActionThrottle40:             "throttle_40",
		ActionThrottle50:             "throttle_50",
		ActionThrottle60:             "throttle_60",
		ActionThrottle70:             "throttle_70",
		ActionThrottle80:             "throttle_80",
		ActionThrottle90:             "throttle_90",
		ActionThrottle100:            "throttle_100",
		ActionJumpJet:                "jump_jet",
		ActionDescend:                "descend",
		ActionWeaponFire:             "weapon_fire",
		ActionWeaponCycle:            "weapon_cycle",
		ActionWeaponCyclePrevious:    "weapon_cycle_prev",
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
		ActionZoomIn:                 "zoom_in",
		ActionZoomOut:                "zoom_out",
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

// allActionsKeymap returns a keymap initialilzed of all available actions with empty key lists
func allActionsKeymap() input.Keymap {
	keymap := input.Keymap{}
	for a := ActionUnknown + 1; a < actionCount; a++ {
		keymap[a] = []input.Key{}
	}
	return keymap
}

func defaultKeyboardMouseControls() input.Keymap {
	keymap := input.Keymap{
		ActionUp:         {input.KeyW, input.KeyUp},
		ActionDown:       {input.KeyS, input.KeyDown},
		ActionLeft:       {input.KeyA, input.KeyLeft},
		ActionRight:      {input.KeyD, input.KeyRight},
		ActionTurretAxes: {input.KeyMouseMotion},

		ActionMenuBack: {input.KeyEscape, input.KeyF1},

		ActionThrottleReverse: {input.KeyBackspace},
		ActionThrottle0:       {input.KeyX},
		ActionJumpJet:         {input.KeySpace},
		ActionDescend:         {input.KeyControl},

		ActionWeaponFire:             {input.KeyMouseLeft},
		ActionWeaponCycle:            {input.KeyMouseRight},
		ActionWeaponGroupFireToggle:  {input.KeyBackslash},
		ActionWeaponGroupSetModifier: {input.KeyShift},
		ActionWeaponGroup1:           {input.Key1},
		ActionWeaponGroup2:           {input.Key2},
		ActionWeaponGroup3:           {input.Key3},
		ActionWeaponGroup4:           {input.Key4},
		ActionWeaponGroup5:           {input.Key5},
		ActionWeaponFireGroup1:       {input.KeyMouseBack},
		ActionWeaponFireGroup2:       {input.KeyMouseForward},

		ActionNavCycle:         {input.KeyN},
		ActionRadarRangeCycle:  {input.KeySlash},
		ActionTargetCrosshairs: {input.KeyQ},
		ActionTargetNearest:    {input.KeyE},
		ActionTargetNext:       {input.KeyT},
		ActionTargetPrevious:   {input.KeyR},

		ActionZoomToggle:     {input.KeyZ},
		ActionLightAmpToggle: {input.KeyL},
		ActionPowerToggle:    {input.KeyP},
		ActionCameraCycle:    {input.KeyF3},
	}
	return input.MergeKeymaps(keymap, allActionsKeymap())
}

func defaultGamepadControls() input.Keymap {
	keymap := input.Keymap{
		ActionMoveAxes:   {input.KeyGamepadLStickMotion},
		ActionTurretAxes: {input.KeyGamepadRStickMotion},

		ActionMenuBack: {input.KeyGamepadStart, input.KeyGamepadBack},

		ActionJumpJet: {input.KeyGamepadLStick},

		ActionWeaponFire:            {input.KeyGamepadR2},
		ActionWeaponCycle:           {input.KeyGamepadR1},
		ActionWeaponGroupFireToggle: {input.KeyGamepadY},

		ActionNavCycle:         {input.KeyGamepadDown},
		ActionTargetCrosshairs: {input.KeyGamepadL2},
		ActionTargetNearest:    {input.KeyGamepadUp},
		ActionTargetNext:       {input.KeyGamepadRight},
		ActionTargetPrevious:   {input.KeyGamepadLeft},

		ActionZoomToggle:     {input.KeyGamepadRStick},
		ActionLightAmpToggle: {input.KeyGamepadDown},
	}
	return input.MergeKeymaps(keymap, allActionsKeymap())
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

func AddKeyBind(keymap input.Keymap, action input.Action, key input.Key) error {
	keyList, exists := keymap[action]
	if !exists {
		keyList = make([]input.Key, 0, 1)
	}
	if slices.Contains(keyList, key) {
		// the key is already bound to this action
		return nil
	}

	if action == ActionMenuBack {
		// do not allow rebinding a key as the menu/back button that is already assigned to any other action
		for a := ActionUnknown + 1; a < actionCount; a++ {
			if a == action {
				continue
			}
			if aKeyList, ok := keymap[a]; ok && slices.Contains(aKeyList, key) {
				return fmt.Errorf("key '%s' already bound to '%s'", key.String(), actionString(a))
			}
		}
	} else {
		// do not allow rebinding anything else to a key already assigned to Menu/Back action
		if menuBackKeyList, ok := keymap[ActionMenuBack]; ok && slices.Contains(menuBackKeyList, key) {
			return fmt.Errorf("key '%s' restricted to '%s'", key.String(), actionString(ActionMenuBack))
		}
	}

	keymap[action] = append(keyList, key)
	return nil
}

func ClearAction(keymap input.Keymap, action input.Action) {
	delete(keymap, action)
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

func (h *InputHandler) PressedActionInfo(action input.Action) (input.EventInfo, bool) {
	return h.handler.PressedActionInfo(action)
}

func (h *InputHandler) JustReleasedActionInfo(action input.Action) (input.EventInfo, bool) {
	return h.handler.JustReleasedActionInfo(action)
}

func (h *InputHandler) SetControls(keyboardMouseMap input.Keymap, gamepadMap input.Keymap) {
	h.keyboardMouseMap = keyboardMouseMap
	h.gamepadMap = gamepadMap
	mergedKeyMap := input.MergeKeymaps(keyboardMouseMap, gamepadMap)
	h.handler.Remap(mergedKeyMap)
}

func (h *InputHandler) KeyboardMouseControls() input.Keymap {
	return h.keyboardMouseMap
}

func (h *InputHandler) GamepadControls() input.Keymap {
	return h.gamepadMap
}

func (g *Game) initControls() {
	g.input = NewInputHandler()

	// import from keymap files if exists
	var keyboardMouseMap input.Keymap
	var gamepadMap input.Keymap

	if _, err := os.Stat(resources.UserKeyboardMouseKeymapFile); err == nil {
		keyboardMouseMap, gamepadMap, err = restoreControls()
		if err != nil {
			panic(fmt.Errorf("error loading keymaps: %v", err))
		}
	}

	// for first time, intitialize defaults and later save to file
	var appliedDefaults bool
	if len(keyboardMouseMap) == 0 {
		log.Debug("initializing default keyboard/mouse keymaps")
		appliedDefaults = true
		keyboardMouseMap = defaultKeyboardMouseControls()
	}
	if len(gamepadMap) == 0 {
		log.Debug("initializing default gamepad keymaps")
		appliedDefaults = true
		gamepadMap = defaultGamepadControls()
	}
	if appliedDefaults {
		defer g.input.saveControls()
	}

	// TODO: initialize default for new actions even if not first time?

	g.input.SetControls(keyboardMouseMap, gamepadMap)
}

func restoreControls() (input.Keymap, input.Keymap, error) {
	log.Debug("restoring keymap file ", resources.UserKeyboardMouseKeymapFile)
	var combinedErrs error
	keyboardMouseMap, err := _loadKeymapFile(resources.UserKeyboardMouseKeymapFile)
	if err != nil {
		combinedErrs = errors.Join(combinedErrs, err)
	}
	gamepadMap, err := _loadKeymapFile(resources.UserGamepadKeymapFile)
	if err != nil {
		combinedErrs = errors.Join(combinedErrs, err)
	}
	return keyboardMouseMap, gamepadMap, combinedErrs
}

func _loadKeymapFile(keymapFilePath string) (input.Keymap, error) {
	keymap := input.Keymap{}

	keymapFile, err := os.Open(keymapFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// caller expected to handle non-existent keymap without error
			return keymap, nil
		}
		return keymap, err
	}
	defer keymapFile.Close()

	fileBytes, err := io.ReadAll(keymapFile)
	if err != nil {
		return keymap, err
	}

	if len(fileBytes) == 0 {
		// caller expected to handle empty keymap without error
		return keymap, nil
	}

	var keymapConfig map[string][]string
	err = json.Unmarshal(fileBytes, &keymapConfig)
	if err != nil {
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
		return keymap, err
	}
	return keymap, nil
}

func (h *InputHandler) saveControls() error {
	log.Debugf("saving keymap files [%s, %s]", resources.UserKeyboardMouseKeymapFile, resources.UserGamepadKeymapFile)

	var combinedErrs error
	err := _saveKeymapFile(h.inputSystem, h.keyboardMouseMap, resources.UserKeyboardMouseKeymapFile)
	if err != nil {
		combinedErrs = errors.Join(combinedErrs, err)
	}
	err = _saveKeymapFile(h.inputSystem, h.gamepadMap, resources.UserGamepadKeymapFile)
	if err != nil {
		combinedErrs = errors.Join(combinedErrs, err)
	}

	if combinedErrs != nil {
		log.Error(combinedErrs)
		return combinedErrs
	}
	return nil
}

func _saveKeymapFile(inputSystem input.System, keymap input.Keymap, keymapFilePath string) error {
	log.Debug("saving keymap file ", keymapFilePath)

	// create local handler to fetch action key names for the given keymap only
	h := inputSystem.NewHandler(0, keymap)

	userConfigPath := filepath.Dir(keymapFilePath)
	if _, err := os.Stat(userConfigPath); os.IsNotExist(err) {
		err = os.MkdirAll(userConfigPath, os.ModePerm)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	keymapFile, err := os.Create(keymapFilePath)
	if err != nil {
		log.Error(err)
		return err
	}
	defer keymapFile.Close()

	keymapConfig := orderedmap.New[string, []string]()
	for a := ActionUnknown + 1; a < actionCount; a++ {
		actionKey := actionString(a)
		keymapConfig.Set(actionKey, h.ActionKeyNames(a, input.AnyDevice))
	}
	keymapJson, _ := json.MarshalIndent(keymapConfig, "", "    ")
	_, err = keymapFile.Write(keymapJson)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

type keyScanHandler struct {
	keymapType       KeymapType
	inputSystem      input.System
	handler          *input.Handler
	keyScanner       *input.KeyScanner
	key              input.Key
	axes             input.Key
	scanningKey      bool
	scanningAxes     bool
	scanCompleteFunc func()
}

func NewKeyScanHandler(keymapType KeymapType) *keyScanHandler {
	keyScanner := &keyScanHandler{keymapType: keymapType}

	var devices input.DeviceKind
	switch keymapType {
	case KeymapTypeKeyboardMouse:
		devices = input.KeyboardDevice | input.MouseDevice
	case KeymapTypeGamepad:
		devices = input.GamepadDevice
	default:
		log.Fatalf("unhandled KeymapType: %v", keymapType)
	}

	keyScanner.inputSystem.Init(input.SystemConfig{DevicesEnabled: devices})
	keyScanner.handler = keyScanner.inputSystem.NewHandler(0, input.Keymap{})
	keyScanner.keyScanner = input.NewKeyScanner(keyScanner.handler)
	return keyScanner

}

func (s *keyScanHandler) update() {
	if !s.scanningKey && !s.scanningAxes {
		return
	}
	s.inputSystem.Update()
	if s.scanningKey {
		s.handleRemapKey()
	} else if s.scanningAxes {
		s.handleRemapAxes()
	}
}

func (s *keyScanHandler) handleRemapKey() {
	key, status := s.keyScanner.Scan()
	if status == input.KeyScanCompleted {
		s.key = key
		s.scanningKey = false
		s.scanCompleteFunc()
	}
}

func (s *keyScanHandler) handleRemapAxes() {
	axes, status := s.keyScanner.ScanAxes()
	if status == input.KeyScanCompleted {
		s.axes = axes
		s.scanningAxes = false
		s.scanCompleteFunc()
	}
}

func (s *keyScanHandler) startKeyScan(scanType keyScanType, scanCompleteFunc func()) {
	s.scanCompleteFunc = scanCompleteFunc
	switch scanType {
	case keyScanKeys:
		s.scanningKey = true
	case keyScanAxes:
		s.scanningAxes = true
	}
}
