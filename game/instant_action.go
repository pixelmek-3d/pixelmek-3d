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

	// place enemy units
	// mission.Mechs = make([]model.MissionUnit, 0, 1)
	// instantActionSpawnUnit(mission, "fire_moth_prime") // FIXME: use unit selected by user

	g.mission = mission
	return mission, nil
}

func instantActionSpawnUnit(mission *model.Mission, unit string) {
	// missionMap := mission.Map()
	// rng := model.NewRNG()

	// var spawnPos [2]float64
	// switch len(missionMap.SpawnPoints) {
	// case 0:
	// 	// generate random spawn point
	// 	// TODO: pick spawn point within some min/max distance of player
	// 	// TODO: check randomized location if in a wall, if so re-roll
	// 	w, h := missionMap.Size()
	// 	spawnPos = [2]float64{float64(rng.Intn(w)) + 0.5, float64(rng.Intn(h)) + 0.5}
	// case 1:
	// 	spawnPos = missionMap.SpawnPoints[0]
	// default:
	// 	// TODO: pick random spawn point that was not picked last time
	// 	spawnPos = missionMap.SpawnPoints[rng.Intn(len(missionMap.SpawnPoints))]
	// }

	// mission.Mechs = append(
	// 	mission.Mechs,
	// 	model.MissionUnit{
	// 		Unit:     unit, // FIXME: if enemy not selected by user, pick one at random
	// 		Position: spawnPos,
	// 	},
	// )
}
