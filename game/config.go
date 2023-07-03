package game

import (
	"image/color"
	"math"
	"os"
	"path/filepath"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/pixelmek-3d/game/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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
	viper.SetDefault("debug", false)
	viper.SetDefault("showFPS", false)

	viper.SetDefault("screen.fovDegrees", 70)
	viper.SetDefault("screen.fullscreen", false)
	viper.SetDefault("screen.vsync", true)
	viper.SetDefault("screen.renderFloor", true)
	viper.SetDefault("screen.renderDistance", 2000)
	viper.SetDefault("screen.clutterDistance", 500)

	if g.osType == osTypeBrowser {
		viper.SetDefault("screen.width", 800)
		viper.SetDefault("screen.height", 600)
		viper.SetDefault("screen.renderScale", 0.5)
	} else {
		viper.SetDefault("screen.width", 1024)
		viper.SetDefault("screen.height", 768)
		viper.SetDefault("screen.renderScale", 1.0)
	}

	if runtime.GOOS == "windows" {
		// default windows to opengl for better performance
		viper.SetDefault("screen.opengl", true)
	}

	viper.SetDefault("hud.enabled", true)
	viper.SetDefault("hud.scale", 1.0)
	viper.SetDefault("hud.font", "pixeloid.otf")
	viper.SetDefault("hud.color.useCustom", false)
	viper.SetDefault("hud.color.red", 100)
	viper.SetDefault("hud.color.green", 255)
	viper.SetDefault("hud.color.blue", 230)
	viper.SetDefault("hud.color.alpha", 255)

	viper.SetDefault("controls.throttleDecay", false)

	// get config values
	g.debug = viper.GetBool("debug")
	g.fpsEnabled = viper.GetBool("showFPS")

	if g.debug {
		log.SetLevel(log.DebugLevel)
	}

	g.screenWidth = viper.GetInt("screen.width")
	g.screenHeight = viper.GetInt("screen.height")
	g.fovDegrees = viper.GetFloat64("screen.fovDegrees")
	g.renderScale = viper.GetFloat64("screen.renderScale")
	g.fullscreen = viper.GetBool("screen.fullscreen")
	g.vsync = viper.GetBool("screen.vsync")
	g.opengl = viper.GetBool("screen.opengl")
	g.initRenderFloorTex = viper.GetBool("screen.renderFloor")

	renderDistanceMeters := viper.GetFloat64("screen.renderDistance")
	g.renderDistance = renderDistanceMeters / model.METERS_PER_UNIT

	clutterDistanceMeters := viper.GetFloat64("screen.clutterDistance")
	g.clutterDistance = clutterDistanceMeters / model.METERS_PER_UNIT

	var err error
	g.fonts.HUDFont, err = g.fonts.LoadFont(viper.GetString("hud.font"))
	if err != nil {
		log.Fatal(err)
		exit(1)
	}
	g.hudEnabled = viper.GetBool("hud.enabled")
	g.hudScale = viper.GetFloat64("hud.scale")
	g.hudUseCustomColor = viper.GetBool("hud.color.useCustom")
	g.hudRGBA = &color.NRGBA{
		R: uint8(viper.GetUint("hud.color.red")),
		G: uint8(viper.GetUint("hud.color.green")),
		B: uint8(viper.GetUint("hud.color.blue")),
		A: uint8(viper.GetUint("hud.color.alpha")),
	}

	g.throttleDecay = viper.GetBool("controls.throttleDecay")
}

func (g *Game) saveConfig() error {
	log.Info("Saving config file ", resources.UserConfigFile)

	userConfigPath := filepath.Dir(resources.UserConfigFile)
	if _, err := os.Stat(userConfigPath); os.IsNotExist(err) {
		err = os.MkdirAll(userConfigPath, os.ModePerm)
		if err != nil {
			log.Error(err)
			return err
		}
	}
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
