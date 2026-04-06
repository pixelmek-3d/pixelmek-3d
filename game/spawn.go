package game

import (
	"fmt"

	"github.com/harbdog/raycaster-go/geom"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/sprites"

	log "github.com/sirupsen/logrus"
)

func randEnemySpawnLocation(g *Game) geom.Vector2 {
	// generate random spawn point within some min/max distance of player
	rng := model.NewRNG()
	missionMap := g.mission.Map()
	w, h := missionMap.Size()

	x, y := int(g.player.Pos().X), int(g.player.Pos().Y)
	rX, rY := rng.RandRelativeLocation(x, y, 20, 40, w, h)
	for missionMap.IsWallAt(0, rX, rY) {
		// location is in a wall, re-roll
		rX, rY = rng.RandRelativeLocation(x, y, 20, 40, w, h)
	}
	return geom.Vector2{X: float64(rX) + 0.5, Y: float64(rY) + 0.5}
}

func spawnUnit[T model.AnyUnitModel](g *Game, u model.Unit) *T {
	unit := u.CloneUnit()
	unit.SetInitialPoweredStatus(model.POWER_OFF_MANUAL)

	missionMap := g.mission.Map()
	rng := model.NewRNG()

	var spawnPos geom.Vector2
	switch len(missionMap.SpawnPoints) {
	case 0:
		// random spawn point within some distance of player
		spawnPos = randEnemySpawnLocation(g)
	case 1:
		spawnPoint := missionMap.SpawnPoints[0]
		spawnPos = geom.Vector2{X: spawnPoint[0], Y: spawnPoint[1]}
	default:
		// TODO: pick random spawn point that was not picked last time
		spawnPoint := missionMap.SpawnPoints[rng.Intn(len(missionMap.SpawnPoints))]
		spawnPos = geom.Vector2{X: spawnPoint[0], Y: spawnPoint[1]}
	}

	unit.SetPos(&spawnPos)

	// attach AI to unit
	g.ai.NewUnitAI(unit)

	// add sprite to game
	sprite := g.CreateUnitSprite(unit)

	var t T
	switch interfaceType := any(t).(type) {
	case model.Mech:
		g.sprites.AddMechSprite(sprite.(*sprites.MechSprite))
	default:
		panic(fmt.Errorf("spawn unit type not implemented: %v", interfaceType))
	}

	return any(unit).(*T)
}

func spawnMissionUnit[T model.AnyUnitModel](g *Game, unit string) *T {

	missionUnit := model.MissionUnit{Unit: unit}

	var u model.Unit
	var t T
	switch interfaceType := any(t).(type) {
	case model.Mech:
		m, err := createMissionUnitModel[model.Mech](g, missionUnit)
		if err != nil {
			log.Errorf("error spawning mission unit: %v", err)
			return nil
		}
		u = m
	default:
		panic(fmt.Errorf("spawn mission unit type not implemented: %v", interfaceType))
	}

	return spawnUnit[T](g, u)
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

	return spawnMissionUnit[T](g, unit)
}
