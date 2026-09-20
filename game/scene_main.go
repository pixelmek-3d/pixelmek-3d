package game

import (
	renderFx "github.com/pixelmek-3d/pixelmek-3d/game/render/effects"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/transitions"

	"github.com/hajimehoshi/ebiten/v2"
)

type MainMenuScene struct {
	Game          *Game
	splash        *SplashScreen
	main          *MainMenu
	settings      *SettingsMenu
	animation     []*ebiten.Image
	animationRate int
	numFrames     int
	animIndex     int
	animCounter   int
	loopCounter   int
	bufferScreen  *ebiten.Image
}

func NewMainMenuScene(g *Game) Scene {
	if !g.audio.IsMusicPlaying() {
		g.audio.StartMenuMusic()
	}

	main := createMainMenu(g)
	settings := createSettingsMenu(g)

	// PixelMek 3D intro animation behind menu
	images := loadIntroImages()

	splash := NewSplashScreen(g)
	splash.geoM = introGeoM(images[0], g.screenRect())
	splash.shader = renderFx.NewCRT()

	tOpts := &transitions.TransitionOptions{
		InDuration:   1.5,
		HoldDuration: 0,
		OutDuration:  0,
	}
	splash.transition = transitions.NewPixelize(splash.screen, tOpts, ebiten.GeoM{})

	scene := &MainMenuScene{
		Game:          g,
		splash:        splash,
		main:          main,
		settings:      settings,
		animation:     images,
		animationRate: introAnimationRate,
		numFrames:     len(images),
		bufferScreen:  ebiten.NewImage(g.screenWidth, g.screenHeight),
	}
	scene.SetMenu(main)
	return scene
}

func (s *MainMenuScene) handleResolutionChange() {
	g := s.Game
	s.settings.initResources()
	s.settings.initMenu()
	s.main.initResources()
	s.main.initMenu()

	s.splash.handleResolutionChange(g.screenRect(), introGeoM(s.animation[0], g.screenRect()))
	s.bufferScreen = ebiten.NewImage(g.screenWidth, g.screenHeight)
}

func (s *MainMenuScene) SetMenu(m Menu) {
	s.Game.menu = m
}

func (s *MainMenuScene) Update() error {
	g := s.Game

	if g.input.ActionIsJustPressed(ActionMenuBack) {
		// if exit window is open, close it
		closedWindow := g.menu.CloseWindow()
		if closedWindow == nil {
			switch g.menu {
			case s.settings:
				g.closeMenu()
			case s.main:
				fallthrough
			default:
				openExitWindow(s.main)
			}
		}
	}

	// determine when to move to next animation frame
	if s.animationRate > 0 {
		if s.animCounter >= s.animationRate {
			s.animCounter = 0
			s.animIndex++
			if s.animIndex >= s.numFrames {
				s.animIndex = 0
				s.loopCounter++
			}
		} else {
			s.animCounter++
		}
	}

	if s.splash.shader != nil && g.crtShader {
		s.splash.shader.Update()
	}

	if s.splash.transition != nil {
		s.splash.transition.Update()
	}

	// update the menu
	g.menu.Update()

	return nil
}

func (s *MainMenuScene) Draw(screen *ebiten.Image) {
	g := s.Game

	// draw current animation frame to screen
	splash := s.splash
	splash.screen.Clear()
	s.bufferScreen.Clear()
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest, GeoM: splash.geoM}
	splash.screen.DrawImage(s.animation[s.animIndex], op)

	if splash.shader != nil && g.crtShader {
		// draw shader effect with splash screen to buffer
		splash.shader.Draw(s.bufferScreen, splash.screen)
	} else {
		// draw splash screen to buffer
		s.bufferScreen.DrawImage(splash.screen, nil)
	}

	if splash.transition != nil {
		// draw transition from buffer
		splash.transition.SetImage(s.bufferScreen)
		splash.transition.Draw(screen)
	} else {
		// draw buffer directly to screen
		screen.DrawImage(s.bufferScreen, nil)
	}

	// draw menu
	g.menu.Draw(screen)
}
