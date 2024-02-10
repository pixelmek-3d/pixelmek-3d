package game

import (
	"image/color"
	"math"
	"os"
	"path/filepath"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	CONFIG_KEY_DEBUG   = "debug"
	CONFIG_KEY_SHOWFPS = "show_fps"

	CONFIG_KEY_SCREEN_WIDTH     = "screen.width"
	CONFIG_KEY_SCREEN_HEIGHT    = "screen.height"
	CONFIG_KEY_RENDER_SCALE     = "screen.render_scale"
	CONFIG_KEY_FOV_DEGREES      = "screen.fov_degrees"
	CONFIG_KEY_FULLSCREEN       = "screen.fullscreen"
	CONFIG_KEY_VSYNC            = "screen.vsync"
	CONFIG_KEY_RENDER_FLOOR     = "screen.render_floor"
	CONFIG_KEY_RENDER_DISTANCE  = "screen.render_distance"
	CONFIG_KEY_CLUTTER_DISTANCE = "screen.clutter_distance"
	CONFIG_KEY_OPENGL           = "screen.opengl"

	CONFIG_KEY_HUD_ENABLED      = "hud.enabled"
	CONFIG_KEY_HUD_SCALE        = "hud.scale"
	CONFIG_KEY_HUD_FONT         = "hud.font"
	CONFIG_KEY_HUD_COLOR_CUSTOM = "hud.color.use_custom"
	CONFIG_KEY_HUD_COLOR_R      = "hud.color.red"
	CONFIG_KEY_HUD_COLOR_G      = "hud.color.green"
	CONFIG_KEY_HUD_COLOR_B      = "hud.color.blue"
	CONFIG_KEY_HUD_COLOR_A      = "hud.color.alpha"

	CONFIG_KEY_AUDIO_BGM_VOL      = "audio.bgm_volume"
	CONFIG_KEY_AUDIO_SFX_VOL      = "audio.sfx_volume"
	CONFIG_KEY_AUDIO_SFX_CHANNELS = "audio.sfx_channels"

	CONFIG_KEY_CONTROL_DECAY = "controls.throttle_decay"
)

func (g *Game) initConfig() {
	// special behavior needed for wasm play
	switch runtime.GOOS {
	case "js":
		g.osType = osTypeBrowser
	default:
		g.osType = osTypeDesktop
	}

	// set default config values
	viper.SetDefault(CONFIG_KEY_DEBUG, false)
	viper.SetDefault(CONFIG_KEY_SHOWFPS, false)

	viper.SetDefault(CONFIG_KEY_FOV_DEGREES, 70)
	viper.SetDefault(CONFIG_KEY_FULLSCREEN, false)
	viper.SetDefault(CONFIG_KEY_VSYNC, true)
	viper.SetDefault(CONFIG_KEY_RENDER_FLOOR, true)
	viper.SetDefault(CONFIG_KEY_RENDER_DISTANCE, 2000)
	viper.SetDefault(CONFIG_KEY_CLUTTER_DISTANCE, 500)

	if g.osType == osTypeBrowser {
		viper.SetDefault(CONFIG_KEY_SCREEN_WIDTH, 800)
		viper.SetDefault(CONFIG_KEY_SCREEN_HEIGHT, 600)
		viper.SetDefault(CONFIG_KEY_RENDER_SCALE, 0.5)
	} else {
		viper.SetDefault(CONFIG_KEY_SCREEN_WIDTH, 1024)
		viper.SetDefault(CONFIG_KEY_SCREEN_HEIGHT, 768)
		viper.SetDefault(CONFIG_KEY_RENDER_SCALE, 1.0)
	}

	if runtime.GOOS == "windows" {
		// default windows to opengl for better performance
		viper.SetDefault(CONFIG_KEY_OPENGL, true)
	}

	// HUD defaults
	viper.SetDefault(CONFIG_KEY_HUD_ENABLED, true)
	viper.SetDefault(CONFIG_KEY_HUD_SCALE, 1.0)
	viper.SetDefault(CONFIG_KEY_HUD_FONT, "pixeloid.otf")
	viper.SetDefault(CONFIG_KEY_HUD_COLOR_CUSTOM, false)
	viper.SetDefault(CONFIG_KEY_HUD_COLOR_R, 100)
	viper.SetDefault(CONFIG_KEY_HUD_COLOR_G, 255)
	viper.SetDefault(CONFIG_KEY_HUD_COLOR_B, 230)
	viper.SetDefault(CONFIG_KEY_HUD_COLOR_A, 255)

	// audio defaults
	viper.SetDefault(CONFIG_KEY_AUDIO_BGM_VOL, 0.65)
	viper.SetDefault(CONFIG_KEY_AUDIO_SFX_VOL, 1.0)
	viper.SetDefault(CONFIG_KEY_AUDIO_SFX_CHANNELS, 16)

	// control defaults
	viper.SetDefault(CONFIG_KEY_CONTROL_DECAY, false)

	// get config values
	g.debug = viper.GetBool(CONFIG_KEY_DEBUG)
	g.fpsEnabled = viper.GetBool(CONFIG_KEY_SHOWFPS)

	if g.debug {
		log.SetLevel(log.DebugLevel)
	}

	g.screenWidth = viper.GetInt(CONFIG_KEY_SCREEN_WIDTH)
	g.screenHeight = viper.GetInt(CONFIG_KEY_SCREEN_HEIGHT)
	g.fovDegrees = viper.GetFloat64(CONFIG_KEY_FOV_DEGREES)
	g.renderScale = viper.GetFloat64(CONFIG_KEY_RENDER_SCALE)
	g.fullscreen = viper.GetBool(CONFIG_KEY_FULLSCREEN)
	g.vsync = viper.GetBool(CONFIG_KEY_VSYNC)
	g.opengl = viper.GetBool(CONFIG_KEY_OPENGL)
	g.initRenderFloorTex = viper.GetBool(CONFIG_KEY_RENDER_FLOOR)

	renderDistanceMeters := viper.GetFloat64(CONFIG_KEY_RENDER_DISTANCE)
	g.renderDistance = renderDistanceMeters / model.METERS_PER_UNIT

	clutterDistanceMeters := viper.GetFloat64(CONFIG_KEY_CLUTTER_DISTANCE)
	g.clutterDistance = clutterDistanceMeters / model.METERS_PER_UNIT

	g.hudEnabled = viper.GetBool(CONFIG_KEY_HUD_ENABLED)
	g.hudFont = viper.GetString(CONFIG_KEY_HUD_FONT)
	g.hudScale = viper.GetFloat64(CONFIG_KEY_HUD_SCALE)
	g.hudUseCustomColor = viper.GetBool(CONFIG_KEY_HUD_COLOR_CUSTOM)
	g.hudRGBA = &color.NRGBA{
		R: uint8(viper.GetUint(CONFIG_KEY_HUD_COLOR_R)),
		G: uint8(viper.GetUint(CONFIG_KEY_HUD_COLOR_G)),
		B: uint8(viper.GetUint(CONFIG_KEY_HUD_COLOR_B)),
		A: uint8(viper.GetUint(CONFIG_KEY_HUD_COLOR_A)),
	}

	bgmVolume = viper.GetFloat64(CONFIG_KEY_AUDIO_BGM_VOL)
	sfxVolume = viper.GetFloat64(CONFIG_KEY_AUDIO_SFX_VOL)
	sfxChannels = viper.GetInt(CONFIG_KEY_AUDIO_SFX_CHANNELS)

	g.throttleDecay = viper.GetBool(CONFIG_KEY_CONTROL_DECAY)
}

func (g *Game) saveConfig() error {
	log.Debug("saving config file ", resources.UserConfigFile)

	userConfigPath := filepath.Dir(resources.UserConfigFile)
	if _, err := os.Stat(userConfigPath); os.IsNotExist(err) {
		err = os.MkdirAll(userConfigPath, os.ModePerm)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	// update stored values in viper in case any may have changed since last written
	viper.Set(CONFIG_KEY_SHOWFPS, g.fpsEnabled)
	viper.Set(CONFIG_KEY_SCREEN_WIDTH, g.screenWidth)
	viper.Set(CONFIG_KEY_SCREEN_HEIGHT, g.screenHeight)
	viper.Set(CONFIG_KEY_RENDER_SCALE, g.renderScale)
	viper.Set(CONFIG_KEY_FOV_DEGREES, g.fovDegrees)
	viper.Set(CONFIG_KEY_FULLSCREEN, g.fullscreen)
	viper.Set(CONFIG_KEY_VSYNC, g.vsync)
	viper.Set(CONFIG_KEY_OPENGL, g.opengl)

	viper.Set(CONFIG_KEY_RENDER_FLOOR, g.initRenderFloorTex)
	viper.Set(CONFIG_KEY_RENDER_DISTANCE, g.renderDistance*model.METERS_PER_UNIT)
	viper.Set(CONFIG_KEY_CLUTTER_DISTANCE, g.clutterDistance*model.METERS_PER_UNIT)

	viper.Set(CONFIG_KEY_HUD_ENABLED, g.hudEnabled)
	viper.Set(CONFIG_KEY_HUD_SCALE, g.hudScale)
	viper.Set(CONFIG_KEY_HUD_COLOR_CUSTOM, g.hudUseCustomColor)
	viper.Set(CONFIG_KEY_HUD_COLOR_R, g.hudRGBA.R)
	viper.Set(CONFIG_KEY_HUD_COLOR_G, g.hudRGBA.G)
	viper.Set(CONFIG_KEY_HUD_COLOR_B, g.hudRGBA.B)
	viper.Set(CONFIG_KEY_HUD_COLOR_A, g.hudRGBA.A)

	viper.Set(CONFIG_KEY_CONTROL_DECAY, g.throttleDecay)

	viper.Set(CONFIG_KEY_AUDIO_BGM_VOL, bgmVolume)
	viper.Set(CONFIG_KEY_AUDIO_SFX_VOL, sfxVolume)
	viper.Set(CONFIG_KEY_AUDIO_SFX_CHANNELS, sfxChannels)

	err := viper.WriteConfigAs(resources.UserConfigFile)
	if err != nil {
		log.Error(err)
	}

	return err
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

	g.windowScene = ebiten.NewImage(g.screenWidth, g.screenHeight)
	g.rayScene = ebiten.NewImage(g.width, g.height)
	if g.camera != nil {
		g.camera.SetViewSize(g.width, g.height)
	}
}

func (g *Game) setRenderDistance(renderDistance float64) {
	g.renderDistance = renderDistance
	if g.camera != nil {
		g.camera.SetRenderDistance(g.renderDistance)
	}
}

func (g *Game) setLightFalloff(lightFalloff float64) {
	g.lightFalloff = lightFalloff
	if g.camera != nil {
		g.camera.SetLightFalloff(g.lightFalloff)
	}
}

func (g *Game) setGlobalIllumination(globalIllumination float64) {
	g.globalIllumination = globalIllumination
	if g.camera != nil {
		g.camera.SetGlobalIllumination(g.globalIllumination)
	}
}

func (g *Game) setLightRGB(minLightRGB, maxLightRGB *color.NRGBA) {
	g.minLightRGB = minLightRGB
	g.maxLightRGB = maxLightRGB
	if g.camera != nil {
		g.camera.SetLightRGB(*g.minLightRGB, *g.maxLightRGB)
	}
}

func (g *Game) setVsyncEnabled(enableVsync bool) {
	g.vsync = enableVsync
	ebiten.SetVsyncEnabled(enableVsync)
}

func (g *Game) setFovAngle(fovDegrees float64) {
	g.fovDegrees = fovDegrees
	if g.camera != nil {
		g.camera.SetFovAngle(fovDegrees, 1.0)
	}
}
