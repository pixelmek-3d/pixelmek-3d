package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/joelschutz/stagehand"
)

type MenuScene struct {
	BaseScene
	activeMenu Menu
	main       *MainMenu
	settings   *SettingsMenu
}

func NewMenuScene(g *Game) *MenuScene {
	if !g.audio.IsMusicPlaying() {
		g.audio.StartMenuMusic()
	}

	main := createMainMenu(g)
	settings := createSettingsMenu(g)

	scene := &MenuScene{
		BaseScene: BaseScene{
			game: g,
		},
		main:     main,
		settings: settings,
	}
	return scene
}

func (s *MenuScene) SetMenu(m Menu) {
	s.activeMenu = m
	s.game.menu = m
}

func (s *MenuScene) Load(st SceneState, sm stagehand.SceneController[SceneState]) {
	s.BaseScene.Load(st, sm)
	s.SetMenu(s.main)
}

func (s *MenuScene) Update() error {
	g := s.game

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
	s.activeMenu.Update()

	return nil
}

func (s *MenuScene) Draw(screen *ebiten.Image) {
	s.activeMenu.Draw(screen)
}
