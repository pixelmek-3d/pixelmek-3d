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
		w, h := missionMap.Size()
		rX, rY := rng.Intn(w), rng.Intn(h)
		for missionMap.IsWallAt(0, rX, rY) {
			// location is in a wall, re-roll
			rX, rY = rng.Intn(w), rng.Intn(h)
		}
		spawnPos = [2]float64{float64(rX) + 0.5, float64(rY) + 0.5}
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

	// attach AI to unit
	g.ai.NewUnitAI(modelMech)

	return modelMech
}
