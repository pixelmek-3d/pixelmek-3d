package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/joelschutz/stagehand"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
)

type InstantActionScene struct {
	BaseScene
	activeMenu     Menu
	missionSelect  *MissionMenu
	unitSelect     *UnitMenu
	launchBriefing *LaunchMenu
}

func NewInstantActionScene(g *Game) *InstantActionScene {
	missionSelect := createMissionMenu(g)
	unitSelect := createUnitMenu(g)
	launchBriefing := createLaunchMenu(g)

	scene := &InstantActionScene{
		BaseScene: BaseScene{
			game: g,
		},
		missionSelect:  missionSelect,
		unitSelect:     unitSelect,
		launchBriefing: launchBriefing,
	}
	return scene
}

func (s *InstantActionScene) SetMenu(m Menu) {
	s.activeMenu = m
	s.game.menu = m
}

func (s *InstantActionScene) Load(st SceneState, sm stagehand.SceneController[SceneState]) {
	s.BaseScene.Load(st, sm)
	s.SetMenu(s.missionSelect)
}

func (s *InstantActionScene) Update() error {
	g := s.game
	if g.input.ActionIsJustPressed(ActionBack) {
		s.back()
	}

	// update the menu
	s.activeMenu.Update()

	return nil
}

func (s *InstantActionScene) Draw(screen *ebiten.Image) {
	// draw menu
	s.activeMenu.Draw(screen)
}

func (s *InstantActionScene) back() {
	g := s.game

	switch g.menu {
	case s.launchBriefing:
		// back to unit select
		s.SetMenu(s.unitSelect)

	case s.unitSelect:
		// back to mission select
		s.SetMenu(s.missionSelect)

	case s.missionSelect:
		fallthrough
	default:
		// back to main menu
		g.sm.ProcessTrigger(MainMenuTrigger)
	}
}

func (s *InstantActionScene) next() {
	g := s.game

	switch g.menu {
	case s.launchBriefing:
		// launch mission scene
		if g.player == nil {
			// pick player unit at random
			g.SetPlayerUnit(g.randomUnit(model.MechResourceType))
		}

		g.sm.ProcessTrigger(LaunchGameTrigger)

	case s.unitSelect:
		// to pre-launch briefing after setting player unit and mission
		g.mission = s.missionSelect.selectedMission

		if s.unitSelect.selectedUnit == nil {
			// set player unit nil to indicate randomized pick for launch briefing
			g.player = nil
		} else {
			g.SetPlayerUnit(s.unitSelect.selectedUnit)
		}

		s.launchBriefing.loadBriefing()
		s.SetMenu(s.launchBriefing)

	case s.missionSelect:
		// to unit select
		s.SetMenu(s.unitSelect)

	default:
		// back to main menu
		g.sm.ProcessTrigger(MainMenuTrigger)
	}
}
