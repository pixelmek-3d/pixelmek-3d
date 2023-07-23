package game

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type MenuScene struct {
	Game     *Game
	main     *MainMenu
	settings *SettingsMenu
}

func NewMenuScene(g *Game) *MenuScene {
	if !g.audio.IsMusicPlaying() {
		g.audio.StartMenuMusic()
	}

	main := createMainMenu(g)
	settings := createSettingsMenu(g)

	scene := &MenuScene{
		Game:     g,
		main:     main,
		settings: settings,
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

	// update the menu
	g.menu.Update()

	return nil
}

func (s *MenuScene) Draw(screen *ebiten.Image) {
	g := s.Game

	// draw menu
	g.menu.Draw(screen)
}
