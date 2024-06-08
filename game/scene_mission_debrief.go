package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/joelschutz/stagehand"
)

type MissionDebriefScene struct {
	BaseScene
	activeMenu     Menu
	missionDebrief *DebriefMenu
}

func NewMissionDebriefScene(g *Game) *MissionDebriefScene {
	scene := &MissionDebriefScene{
		BaseScene: BaseScene{
			game: g,
		},
	}
	return scene
}

func (s *MissionDebriefScene) SetMenu(m Menu) {
	s.activeMenu = m
	s.game.menu = m
}

func (s *MissionDebriefScene) Load(st SceneState, sm stagehand.SceneController[SceneState]) {
	s.BaseScene.Load(st, sm)

	s.missionDebrief = createDebriefMenu(s.game)
	s.SetMenu(s.missionDebrief)
}

func (s *MissionDebriefScene) Update() error {
	g := s.game

	if g.input.ActionIsJustPressed(ActionBack) {
		s.back()
	}

	// update the menu
	s.activeMenu.Update()

	return nil
}

func (s *MissionDebriefScene) Draw(screen *ebiten.Image) {
	// draw menu
	s.activeMenu.Draw(screen)
}

func (s *MissionDebriefScene) back() {
	g := s.game

	// back to main menu
	g.sm.ProcessTrigger(MainMenuTrigger)
}
