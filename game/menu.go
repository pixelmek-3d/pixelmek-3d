package game

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/gabstv/ebiten-imgui/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/inkyblackness/imgui-go/v4"
)

type DemoMenu struct {
	mgr    *renderer.Manager
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

func mainMenu() DemoMenu {
	mgr := renderer.New(nil)
	return DemoMenu{
		mgr:    mgr,
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
	g.menu.newRenderDistance = float32(g.renderDistance)
	g.menu.newClutterDistance = float32(g.clutterDistance)

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

func (m *DemoMenu) layout(w, h int) {
	m.mgr.SetDisplaySize(float32(w), float32(h))
}

func (m *DemoMenu) update(g *Game) {
	if !m.active {
		return
	}

	m.mgr.Update(1.0 / float32(ebiten.TPS()))

	m.mgr.BeginFrame()

	windowFlags := imgui.WindowFlagsNone
	windowFlags |= imgui.WindowFlagsAlwaysAutoResize
	windowFlags |= imgui.WindowFlagsAlwaysVerticalScrollbar
	windowFlags |= imgui.WindowFlagsHorizontalScrollbar
	windowFlags |= imgui.WindowFlagsMenuBar

	if !imgui.BeginV("Settings", nil, windowFlags) {
		// Early out if the window is collapsed, as an optimization.
		imgui.End()
		m.mgr.EndFrame()
		return
	}

	if imgui.BeginMenuBar() {
		if imgui.BeginMenu("Resume") {
			if imgui.MenuItem("Return to duty") {
				g.closeMenu()
			}
			imgui.EndMenu()
		}

		// provide separation between Resume and Exit options
		if imgui.BeginMenuV(strings.Repeat(" ", 8), false) {
			imgui.EndMenu()
		}
		imgui.Separator()
		if imgui.BeginMenuV(strings.Repeat(" ", 7), false) {
			imgui.EndMenu()
		}

		if imgui.BeginMenu("Exit") {
			if imgui.MenuItem("Embrace cowardice") {
				exit(0)
			}
			imgui.EndMenu()
		}

		imgui.EndMenuBar()
	}

	// Set resolution by using int input fields and button to set it
	{
		imgui.Text("Resolution:")

		imgui.Indent()
		imgui.Text(" Width")
		imgui.SameLine()
		imgui.InputInt("##renderWidth", &m.newRenderWidth)

		imgui.Text("Height")
		imgui.SameLine()
		imgui.InputInt("##renderHeight", &m.newRenderHeight)

		if imgui.Button("Apply") {
			g.setResolution(int(m.newRenderWidth), int(m.newRenderHeight))
		}

		imgui.Unindent()
	}

	// Render scaling: +/- buttons to constrict values (0.25 <= s <= 1.0 in 0.25 increments only)
	{
		imgui.Text(fmt.Sprintf("Render Scaling: %0.2f", m.newRenderScale))
		imgui.SameLine()

		if imgui.Button("-") {
			m.newRenderScale -= 0.25
			if m.newRenderScale < 0.25 {
				m.newRenderScale = 0.25
			}
			g.setRenderScale(float64(m.newRenderScale))
		}

		imgui.SameLine()
		if imgui.Button("+") {
			m.newRenderScale += 0.25
			if m.newRenderScale > 1.0 {
				m.newRenderScale = 1.0
			}
			g.setRenderScale(float64(m.newRenderScale))
		}
	}

	if imgui.SliderFloatV("FOV", &m.newFovDegrees, 40, 140, "%.0f", imgui.SliderFlagsNone) {
		g.setFovAngle(float64(m.newFovDegrees))
	}

	if imgui.SliderFloatV("Render Distance", &m.newRenderDistance, -1, 1000, "%.0f", imgui.SliderFlagsNone) {
		g.renderDistance = float64(m.newRenderDistance)
		g.camera.SetRenderDistance(g.renderDistance)
	}

	if imgui.SliderFloatV("Clutter Distance", &m.newClutterDistance, 0, 20, "%.0f", imgui.SliderFlagsNone) {
		g.clutterDistance = float64(m.newClutterDistance)
		g.clutter.Update(g, true)
	}

	if imgui.Checkbox("Fullscreen", &g.fullscreen) {
		g.setFullscreen(g.fullscreen)
	}

	if imgui.Checkbox("Use VSync", &g.vsync) {
		g.setVsyncEnabled(g.vsync)
	}

	imgui.Checkbox("Floor Texturing", &g.tex.renderFloorTex)

	// New section for HUD options
	imgui.Separator()
	imgui.Text("HUD:")

	if imgui.SliderFloatV("Scaling", &m.newHudScale, 0.2, 5.0, "%.1f", imgui.SliderFlagsNone) {
		g.hudScale = float64(m.newHudScale)
	}

	hudColorChanged := false
	if imgui.ColorEdit4V("Color", &m.newHudRGBA, imgui.ColorEditFlagsAlphaBar) {
		hudColorChanged = true
	}
	if hudColorChanged {
		g.hudRGBA = color.RGBA{
			R: byte(m.newHudRGBA[0] * 255),
			G: byte(m.newHudRGBA[1] * 255),
			B: byte(m.newHudRGBA[2] * 255),
			A: byte(m.newHudRGBA[3] * 255),
		}
	}

	// New section for lighting options (TODO: should be DEBUG only)
	imgui.Separator()

	imgui.Text("Lighting:")

	if imgui.SliderFloatV("Light Falloff", &m.newLightFalloff, -500, 500, "%.0f", imgui.SliderFlagsNone) {
		g.lightFalloff = float64(m.newLightFalloff)
		g.camera.SetLightFalloff(g.lightFalloff)
	}

	if imgui.SliderFloatV("Global Illumination", &m.newGlobalIllumination, 0, 1000, "%.0f", imgui.SliderFlagsNone) {
		g.globalIllumination = float64(m.newGlobalIllumination)
		g.camera.SetGlobalIllumination(g.globalIllumination)
	}

	lightColorChanged := false
	if imgui.ColorEdit3("Min Lighting", &m.newMinLightRGB) {
		lightColorChanged = true
	}
	if imgui.ColorEdit3("Max Lighting", &m.newMaxLightRGB) {
		lightColorChanged = true
	}

	if lightColorChanged {
		// need to handle menu derived value as a fraction of 1/255
		g.minLightRGB = color.NRGBA{
			R: byte(m.newMinLightRGB[0] * 255), G: byte(m.newMinLightRGB[1] * 255), B: byte(m.newMinLightRGB[2] * 255),
		}
		g.maxLightRGB = color.NRGBA{
			R: byte(m.newMaxLightRGB[0] * 255), G: byte(m.newMaxLightRGB[1] * 255), B: byte(m.newMaxLightRGB[2] * 255),
		}
		g.camera.SetLightRGB(g.minLightRGB, g.maxLightRGB)
	}

	imgui.End()
	m.mgr.EndFrame()
}

func (m *DemoMenu) draw(screen *ebiten.Image) {
	if !m.active {
		return
	}

	m.mgr.Draw(screen)
}
