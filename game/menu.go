package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/pixelmek-3d/game/model"
)

type GameMenu struct {
	active bool

	// held vars that should not get updated in real-time
	newRenderWidth     int32
	newRenderHeight    int32
	newRenderScale     float32
	newFovDegrees      float32
	newRenderDistance  float32
	newClutterDistance float32

	newHudScale float32
	newHudRGBA  [4]float32

	// DEBUG only options
	newGlobalIllumination float32
	newLightFalloff       float32
	newMinLightRGB        [3]float32
	newMaxLightRGB        [3]float32
}

func mainMenu() GameMenu {
	return GameMenu{
		active: false,
	}
}

func (g *Game) openMenu() {
	g.paused = true
	g.menu.active = true
	g.mouseMode = MouseModeCursor
	ebiten.SetCursorMode(ebiten.CursorModeVisible)

	// setup initial values for held vars that should not get updated in real-time
	g.menu.newRenderWidth = int32(g.screenWidth)
	g.menu.newRenderHeight = int32(g.screenHeight)
	g.menu.newRenderScale = float32(g.renderScale)
	g.menu.newFovDegrees = float32(g.fovDegrees)
	g.menu.newRenderDistance = float32(g.renderDistance * model.METERS_PER_UNIT)
	g.menu.newClutterDistance = float32(g.clutterDistance * model.METERS_PER_UNIT)

	g.menu.newHudScale = float32(g.hudScale)
	g.menu.newHudRGBA = [4]float32{
		float32(g.hudRGBA.R) * 1 / 255,
		float32(g.hudRGBA.G) * 1 / 255,
		float32(g.hudRGBA.B) * 1 / 255,
		float32(g.hudRGBA.A) * 1 / 255,
	}

	g.menu.newLightFalloff = float32(g.lightFalloff)
	g.menu.newGlobalIllumination = float32(g.globalIllumination)

	// for color menu items [1, 1, 1] represents NRGBA{255, 255, 255}
	g.menu.newMinLightRGB = [3]float32{
		float32(g.minLightRGB.R) * 1 / 255, float32(g.minLightRGB.G) * 1 / 255, float32(g.minLightRGB.B) * 1 / 255,
	}
	g.menu.newMaxLightRGB = [3]float32{
		float32(g.maxLightRGB.R) * 1 / 255, float32(g.maxLightRGB.G) * 1 / 255, float32(g.maxLightRGB.B) * 1 / 255,
	}
}

func (g *Game) closeMenu() {
	g.paused = false
	g.menu.active = false
}

func (m *GameMenu) layout(w, h int) {
	//m.mgr.SetDisplaySize(float32(w), float32(h))
}

func (m *GameMenu) update(g *Game) {
	if !m.active {
		return
	}

	// hudColorChanged := false
	// if hudColorChanged {
	// 	g.hudRGBA = color.RGBA{
	// 		R: byte(m.newHudRGBA[0] * 255),
	// 		G: byte(m.newHudRGBA[1] * 255),
	// 		B: byte(m.newHudRGBA[2] * 255),
	// 		A: byte(m.newHudRGBA[3] * 255),
	// 	}

	// 	// regenerate nav sprites to pick up color change
	// 	g.loadNavSprites()
	// }

	// lightColorChanged := false
	// if lightColorChanged {
	// 	// need to handle menu derived value as a fraction of 1/255
	// 	g.minLightRGB = color.NRGBA{
	// 		R: byte(m.newMinLightRGB[0] * 255), G: byte(m.newMinLightRGB[1] * 255), B: byte(m.newMinLightRGB[2] * 255),
	// 	}
	// 	g.maxLightRGB = color.NRGBA{
	// 		R: byte(m.newMaxLightRGB[0] * 255), G: byte(m.newMaxLightRGB[1] * 255), B: byte(m.newMaxLightRGB[2] * 255),
	// 	}
	// 	g.camera.SetLightRGB(g.minLightRGB, g.maxLightRGB)
	// }
}

func (m *GameMenu) draw(screen *ebiten.Image) {
	if !m.active {
		return
	}

	//m.mgr.Draw(screen)
}
