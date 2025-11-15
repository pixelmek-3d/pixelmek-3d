package game

import (
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
)

func (g *Game) LoadInstantAction(mapFile string) (*model.Mission, error) {
	mission := &model.Mission{MapPath: mapFile}

	// load mission map
	err := mission.LoadMissionMap()
	if err != nil {
		return nil, err
	}

	missionMap := mission.Map()
	mission.Title = "Instant Action\n" + missionMap.Name
	mission.Briefing = "Destroy never-ending waves of enemies."

	// initialize enemy spawns
	mission.SpawnPoints = make([]*model.SpawnPoint, 0, len(missionMap.SpawnPoints))
	for _, spawnPos := range missionMap.SpawnPoints {
		mission.SpawnPoints = append(mission.SpawnPoints, &model.SpawnPoint{Position: spawnPos})
	}

	// set mission objectives
	mission.Objectives = &model.MissionObjectives{
		Destroy: []*model.MissionDestroyObjectives{
			{
				All:   true,
				Waves: true,
			},
		},
	}

	g.mission = mission
	return mission, nil
}
