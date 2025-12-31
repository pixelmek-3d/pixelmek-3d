package game

import (
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
)

type InstantActionMissionOpts struct {
	enemies []model.Unit
}

func (g *Game) LoadInstantActionFromMapPath(mapPath string, opts *InstantActionMissionOpts) (*model.Mission, error) {
	mission, err := model.NewMissionFromMapPath(mapPath)
	if err != nil {
		return nil, err
	}
	initInstantActionMission(mission, opts)
	g.mission = mission
	return mission, nil
}

func (g *Game) LoadInstantActionFromMap(missionMap *model.Map, opts *InstantActionMissionOpts) (*model.Mission, error) {
	mission, err := model.NewMissionFromMap(missionMap)
	if err != nil {
		return nil, err
	}
	initInstantActionMission(mission, opts)
	g.mission = mission
	return mission, nil
}

func initInstantActionMission(mission *model.Mission, opts *InstantActionMissionOpts) {
	missionMap := mission.Map()
	mission.Title = "Instant Action\n" + missionMap.Name
	mission.Briefing = "Destroy never-ending waves of enemies."

	var enemies []model.Unit
	if opts != nil {
		enemies = opts.enemies
	}

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
				Waves: &model.UnitWaves{Units: enemies},
			},
		},
	}
}
