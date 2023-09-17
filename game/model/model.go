package model

import (
	"math/rand"
	"time"
)

const (
	METERS_PER_UNIT  float64 = 20
	TICKS_PER_SECOND float64 = 60
	SECONDS_PER_TICK float64 = 1 / TICKS_PER_SECOND

	VELOCITY_TO_KPH float64 = (METERS_PER_UNIT / 1000) * (TICKS_PER_SECOND * 60 * 60)
	KPH_TO_VELOCITY float64 = 1 / VELOCITY_TO_KPH

	GRAVITY_METERS_PSS float64 = 9.80665
	GRAVITY_UNITS_PTT  float64 = GRAVITY_METERS_PSS / METERS_PER_UNIT / (TICKS_PER_SECOND * TICKS_PER_SECOND)

	CEILING_JUMP float64 = 2.0
	CEILING_VTOL float64 = 5.0 // TODO: set flight ceiling in map yaml
)

var (
	Randish *rand.Rand
)

func init() {
	Randish = rand.New(rand.NewSource(time.Now().UnixNano()))
}

type TechBase int

const (
	CLAN TechBase = iota
	IS
)

func TechBaseString(t TechBase) string {
	switch t {
	case CLAN:
		return "clan"
	case IS:
		return "is"
	}
	return "unknown"
}

type ModelTech struct {
	TechBase
}

type HeatSinkType int

const (
	NONE   HeatSinkType = iota // 0
	SINGLE                     // 1
	DOUBLE                     // 2
)

type ModelHeatSinkType struct {
	HeatSinkType
}

type Location int

const (
	HEAD Location = iota
	CENTER_TORSO
	LEFT_TORSO
	RIGHT_TORSO
	LEFT_ARM
	RIGHT_ARM
	LEFT_LEG
	RIGHT_LEG
	FRONT
	RIGHT
	LEFT
	TURRET
)

type ModelLocation struct {
	Location
}

type AmmoType int

const (
	AMMO_BALLISTIC = iota
	AMMO_LRM
	AMMO_SRM
	AMMO_STREAK_SRM
)

type Ammo struct {
	AmmoBins map[AmmoType]*AmmoBin
}

type AmmoBin struct {
	AmmoCount int
	AmmoMax   int
}

type ModelWeaponType struct {
	WeaponType
}
