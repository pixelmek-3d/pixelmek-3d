package game

import "github.com/pixelmek-3d/pixelmek-3d/game/model"

func (g *Game) LoadInstantAction(mapFile string) (*model.Mission, error) {
	mission := &model.Mission{
		MapPath: mapFile,
		DropZone: &model.MissionDropZone{
			Position: [2]float64{50, 50}, // FIXME: random generated or new spec in map yaml
			Heading:  15,
		},
	}

	// load mission map
	err := mission.LoadMissionMap()
	if err != nil {
		return nil, err
	}

	mission.Title = "Instant Action\n" + mission.Map().Name
	mission.Briefing = "Destroy never ending waves of enemies."

	// TODO: implement enemy waves for instant action map play

	// place enemy units
	mission.Mechs = []model.MissionUnit{
		{
			Unit:     "fire_moth_prime",  // FIXME: random generated or selected by user
			Position: [2]float64{57, 65}, // FIXME: random generated or new spec in map yaml
		},
	}

	g.mission = mission
	return mission, nil
}
