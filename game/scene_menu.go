package game

import (
	"github.com/pixelmek-3d/pixelmek-3d/game/render/transitions"

	"github.com/hajimehoshi/ebiten/v2"
)

type MenuScene struct {
	Game             *Game
	main             *MainMenu
	settings         *SettingsMenu
	transition       SceneTransition
	transitionScreen *ebiten.Image
}

func NewMenuScene(g *Game) *MenuScene {
	if !g.audio.IsMusicPlaying() {
		g.audio.StartMenuMusic()
	}

	main := createMainMenu(g)
	settings := createSettingsMenu(g)

	transitionScreen := ebiten.NewImage(g.screenWidth, g.screenHeight)
	tOpts := &transitions.TransitionOptions{
		InDuration:   1.5,
		HoldDuration: 0,
		OutDuration:  0,
	}
	transition := transitions.NewFade(transitionScreen, tOpts, ebiten.GeoM{})

	scene := &MenuScene{
		Game:             g,
		main:             main,
		settings:         settings,
		transition:       transition,
		transitionScreen: transitionScreen,
	}
	scene.SetMenu(main)
	return scene
}

func (s *MenuScene) SetMenu(m Menu) {
	s.Game.menu = m
}

func (s *MenuScene) Update() error {
	g := s.Game

	if g.input.ActionIsJustPressed(ActionBack) {
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

	if s.transition != nil {
		s.transition.Update()
		if s.transition.Completed() {
			s.transition = nil
			s.transitionScreen = nil
		}
	}

	// update the menu
	g.menu.Update()

	return nil
}

func (s *MenuScene) Draw(screen *ebiten.Image) {
	g := s.Game

	if s.transition != nil {
		// draw menu with transition
		g.menu.Draw(s.transitionScreen)
		s.transition.SetImage(s.transitionScreen)
		s.transition.Draw(screen)
	} else {
		// draw menu
		g.menu.Draw(screen)
	}
}
