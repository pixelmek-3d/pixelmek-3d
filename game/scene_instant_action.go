package game

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type InstantActionScene struct {
	Game *Game
	// mapSelect  *MissionMenu // TODO: MapMenu
	playerUnitSelect *UnitMenu
	enemyUnitSelect  *UnitMenu
	// launchBriefing *LaunchMenu // TODO: launch briefing menu
}

func NewInstantActionScene(g *Game) Scene {
	// mapSelect
	unitSelect := createUnitMenu(g)
	enemySelect := createUnitMenu(g)
	// launchBriefing

	scene := &InstantActionScene{
		Game:             g,
		playerUnitSelect: unitSelect,
		enemyUnitSelect:  enemySelect,
	}
	scene.SetMenu(unitSelect)
	return scene
}

func (s *InstantActionScene) SetMenu(m Menu) {
	s.Game.menu = m
}

func (s *InstantActionScene) Update() error {
	g := s.Game

	if g.input.ActionIsJustPressed(ActionBack) {
		s.back()
	}

	// update the menu
	g.menu.Update()

	return nil
}

func (s *InstantActionScene) Draw(screen *ebiten.Image) {
	g := s.Game

	// draw menu
	g.menu.Draw(screen)
}

func (s *InstantActionScene) back() {
	g := s.Game

	switch g.menu {
	// TODO: case s.launchBriefing:
	//           back to enemy unit select

	case s.enemyUnitSelect:
		s.SetMenu(s.playerUnitSelect)

	case s.playerUnitSelect:
		// TODO: back to map select, then s.mapSelect is the fallthrough case

		// TODO: case s.mapSelect:
		fallthrough
	default:
		// back to main menu
		g.scene = NewMainMenuScene(g)
	}
}

func (s *InstantActionScene) next() {
	// FIXME: refactor back/next to work properly when not MissionScene
	g := s.Game

	switch g.menu {
	// TODO: case s.launchBriefing:
	//           launch game scene into map mission

	case s.enemyUnitSelect:
		// to pre-launch briefing after enemy player unit and map
		if s.enemyUnitSelect.selectedUnit == nil {
			// set enemy unit nil to indicate randomized pick
		} else {

		}

		// TODO: s.SetMenu(s.launchBriefing)
		// TODO: move launch game scene step to launchBriefing case

	case s.playerUnitSelect:
		// to enemy unit select after setting player unit
		if s.playerUnitSelect.selectedUnit == nil {
			// set player unit nil to indicate randomized pick for launch briefing
			g.player = nil
		} else {
			g.SetPlayerUnit(s.playerUnitSelect.selectedUnit)
		}

		s.SetMenu(s.enemyUnitSelect)

	// TODO: case s.mapSelect:
	//           to unit select

	default:
		// back to main menu
		g.scene = NewMainMenuScene(g)
	}
}
