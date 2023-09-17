package model

import (
	"math"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
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
	AMMO_NOT_APPLICABLE = iota
	AMMO_BALLISTIC
	AMMO_LRM
	AMMO_SRM
	AMMO_STREAK_SRM
)

type ModelAmmoType struct {
	AmmoType
}

type ModelWeaponType struct {
	WeaponType
}

type Ammo struct {
	ammoBins []*AmmoBin
}

type AmmoBin struct {
	ammoType  AmmoType
	forWeapon Weapon
	ammoCount int
	ammoMax   int
}

func AmmoTypeForWeapon(forWeapon Weapon) AmmoType {
	switch w := forWeapon.(type) {
	case *EnergyWeapon:
		// energy weapons consume no ammo
		return AMMO_NOT_APPLICABLE
	case *MissileWeapon:
		// missile weapons use ammo pools based on missile class
		switch w.Classification() {
		case MISSILE_LRM:
			return AMMO_LRM
		case MISSILE_SRM:
			// determine if Streak SRM vs. SRM
			if w.IsLockOnLockRequired() {
				return AMMO_STREAK_SRM
			} else {
				return AMMO_SRM
			}
		default:
			log.Errorf("unhandled ammo type for missile class weapon '%s'", forWeapon.File())
		}
	case *BallisticWeapon:
		// ballistic weapons use same ammo type (just weapon specific ammo bins)
		return AMMO_BALLISTIC
	default:
		log.Errorf("unhandled ammo type for weapon '%s'", forWeapon.File())
	}
	return AMMO_NOT_APPLICABLE
}

func NewAmmoStock() *Ammo {
	return &Ammo{ammoBins: make([]*AmmoBin, 0)}
}

// AddAmmoBin creates ammo bin or updates existing one for the same ammo type/weapon
func (a *Ammo) AddAmmoBin(ammoType AmmoType, ammoTons float64, forWeapon Weapon) {
	if forWeapon == nil {
		log.Errorf("forWeapon parameter is required to add ammo bin for ammo type '%v'", ammoType)
		return
	}

	// find existing ammo bin if present and update it, otherwise create new one
	ammoBin := a.GetAmmoBin(ammoType, forWeapon)
	if ammoBin == nil {
		ammoBin = &AmmoBin{
			ammoType:  ammoType,
			forWeapon: forWeapon,
		}
		a.ammoBins = append(a.ammoBins, ammoBin)
	}

	// use weapon to determine ammo count based on ammo per ton
	ammoPerTon := float64(forWeapon.AmmoPerTon())
	addAmmoCount := int(math.Ceil(ammoPerTon * ammoTons))
	ammoBin.ammoCount += addAmmoCount
	ammoBin.ammoMax += addAmmoCount
}

// GetAmmoBin finds existing ammo bin, if present, for given weapon
func (a *Ammo) GetAmmoBin(ammoType AmmoType, forWeapon Weapon) *AmmoBin {
	if ammoType == AMMO_NOT_APPLICABLE {
		return nil
	}

	forWeaponFile := forWeapon.File()
	for _, ammoBin := range a.ammoBins {
		switch ammoType {
		case AMMO_BALLISTIC:
			// ballistic ammo weapons only share ammo bins with same caliber weapon
			if ammoType == ammoBin.ammoType && forWeaponFile == ammoBin.forWeapon.File() {
				return ammoBin
			}
		default:
			if ammoType == ammoBin.ammoType {
				return ammoBin
			}
		}
	}
	return nil
}
