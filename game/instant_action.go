package game

import (
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
)

func (g *Game) LoadInstantActionFromMapPath(mapPath string) (*model.Mission, error) {
	mission, err := model.NewMissionFromMapPath(mapPath)
	if err != nil {
		return nil, err
	}
	initInstantActionMission(mission)
	g.mission = mission
	return mission, nil
}

func (g *Game) LoadInstantActionFromMap(missionMap *model.Map) (*model.Mission, error) {
	mission, err := model.NewMissionFromMap(missionMap)
	if err != nil {
		return nil, err
	}
	initInstantActionMission(mission)
	g.mission = mission
	return mission, nil
}

func initInstantActionMission(mission *model.Mission) {
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
}
