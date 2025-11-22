package game

import (
	"fmt"

	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/sprites"

	log "github.com/sirupsen/logrus"
)

func spawnUnit[T model.AnyUnitModel](g *Game, unit string) *T {
	missionMap := g.mission.Map()
	rng := model.NewRNG()

	var spawnPos [2]float64
	switch len(missionMap.SpawnPoints) {
	case 0:
		// generate random spawn point within some min/max distance of player
		w, h := missionMap.Size()
		x, y := int(g.player.Pos().X), int(g.player.Pos().Y)
		rX, rY := rng.RandRelativeLocation(x, y, 10, 20, w, h) // TODO: pick better min/max
		for missionMap.IsWallAt(0, rX, rY) {
			// location is in a wall, re-roll
			rX, rY = rng.RandRelativeLocation(x, y, 10, 20, w, h) // TODO: pick better min/max
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

	var u model.Unit
	var t T
	switch interfaceType := any(t).(type) {
	case model.Mech:
		m, err := createMissionUnitModel[model.Mech](g, missionUnit)
		if err != nil {
			log.Errorf("error spawning mission unit: %v", err)
			return nil
		}
		spriteMech := g.createUnitSprite(m).(*sprites.MechSprite)
		g.sprites.AddMechSprite(spriteMech)
		u = m
	default:
		panic(fmt.Errorf("spawn unit type not implemented: %v", interfaceType))
	}

	// attach AI to unit
	g.ai.NewUnitAI(u)

	return any(u).(*T)
}

// TODO: add constraints to randomness, such as tech base and weight class
func spawnRandomUnit[T model.AnyUnitModel](g *Game) *T {
	var unit string
	var t T
	switch interfaceType := any(t).(type) {
	case model.Mech:
		unit = model.RandomMapKey(g.resources.Mechs)
	default:
		panic(fmt.Errorf("spawn random unit type not implemented: %v", interfaceType))
	}

	return spawnUnit[T](g, unit)
}
