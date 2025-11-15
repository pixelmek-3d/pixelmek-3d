package game

import (
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/sprites"

	log "github.com/sirupsen/logrus"
)

func (g *Game) spawnUnit(unit string) model.Unit {
	missionMap := g.mission.Map()
	rng := model.NewRNG()

	var spawnPos [2]float64
	switch len(missionMap.SpawnPoints) {
	case 0:
		// generate random spawn point
		// TODO: pick spawn point within some min/max distance of player
		// TODO: check randomized location if in a wall, if so re-roll
		w, h := missionMap.Size()
		spawnPos = [2]float64{float64(rng.Intn(w)) + 0.5, float64(rng.Intn(h)) + 0.5}
	case 1:
		spawnPos = missionMap.SpawnPoints[0]
	default:
		// TODO: pick random spawn point that was not picked last time
		spawnPos = missionMap.SpawnPoints[rng.Intn(len(missionMap.SpawnPoints))]
	}

	missionUnit := model.MissionUnit{
		Unit:     unit,
		Position: spawnPos,
	}

	// TODO: support non-mech unit types
	modelMech, err := createMissionUnitModel[model.Mech](g, missionUnit)
	if err != nil {
		log.Errorf("error spawning mission unit: %v", err)
		return nil
	}
	spriteMech := g.createUnitSprite(modelMech).(*sprites.MechSprite)
	g.sprites.AddMechSprite(spriteMech)

	g.ai.NewUnitAI(modelMech)
	g.ai.initiative.roll()

	return modelMech
}
