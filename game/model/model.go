package model

import (
	"math"
	"reflect"

	"github.com/harbdog/raycaster-go/geom"
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

	CEILING_JUMP float64 = 5.0
	CEILING_VTOL float64 = 5.0 // TODO: set flight ceiling in map yaml
)

type TechBase int

const (
	COMMON TechBase = iota
	CLAN
	IS
)

func TechBaseString(t TechBase) string {
	switch t {
	case COMMON:
		return "common"
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

type Equipment struct {
	destroyed bool
}

func (e *Equipment) Destroyed() bool {
	return e.destroyed
}

func (e *Equipment) SetDestroyed(isDestroyed bool) {
	e.destroyed = isDestroyed
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

func (t AmmoType) Name() string {
	switch t {
	case AMMO_LRM:
		return "Long Range Missile"
	case AMMO_SRM:
		return "Short Range Missile"
	case AMMO_STREAK_SRM:
		return "Streak Short Range Missile"
	case AMMO_BALLISTIC:
		return "Ballistic Weapon"
	default:
		return ""
	}
}

func (t AmmoType) ShortName() string {
	switch t {
	case AMMO_LRM:
		return "LRM"
	case AMMO_SRM:
		return "SRM"
	case AMMO_STREAK_SRM:
		return "SSRM"
	case AMMO_BALLISTIC:
		return "BALLISTIC"
	default:
		return ""

	}
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

// AmmoBinList returns the list of ammo bins
func (a *Ammo) AmmoBinList() []*AmmoBin {
	return a.ammoBins
}

// AddAmmoBin creates ammo bin or updates existing one for the same ammo type/weapon
func (a *Ammo) AddAmmoBin(ammoType AmmoType, ammoTons float64, forWeapon Weapon) *AmmoBin {
	if forWeapon == nil {
		log.Errorf("forWeapon parameter is required to add ammo bin for ammo type '%v'", ammoType)
		return nil
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
	return ammoBin
}

// GetAmmoBin finds existing ammo bin, if present, for given weapon
func (a *Ammo) GetAmmoBin(ammoType AmmoType, forWeapon Weapon) *AmmoBin {
	if forWeapon == nil || ammoType == AMMO_NOT_APPLICABLE {
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

// CheckAmmo checks the available ammount count of a weapon
func (a *Ammo) CheckAmmo(forWeapon Weapon) int {
	ammoType := AmmoTypeForWeapon(forWeapon)
	if ammoType == AMMO_NOT_APPLICABLE {
		return math.MaxInt
	}

	ammoBin := a.GetAmmoBin(ammoType, forWeapon)
	if ammoBin != nil {
		return ammoBin.ammoCount
	}
	return 0
}

// ConsumeAmmo consumes the ammo count of weapon fired N times
func (a *Ammo) ConsumeAmmo(forWeapon Weapon, consumeN int) *AmmoBin {
	ammoType := AmmoTypeForWeapon(forWeapon)
	if forWeapon == nil || ammoType == AMMO_NOT_APPLICABLE {
		return nil
	}

	ammoBin := a.GetAmmoBin(ammoType, forWeapon)
	if ammoBin != nil && ammoBin.ammoCount > 0 {
		ammoBin.ConsumeAmmo(forWeapon, consumeN)
	}
	return ammoBin
}

func (a *AmmoBin) AmmoCount() int {
	return a.ammoCount
}

func (a *AmmoBin) AmmoMax() int {
	return a.ammoMax
}

func (a *AmmoBin) AmmoType() AmmoType {
	return a.ammoType
}

func (a *AmmoBin) ForWeapon() Weapon {
	return a.forWeapon
}

// ConsumeAmmo consumes the ammo count of weapon fired N times
func (a *AmmoBin) ConsumeAmmo(forWeapon Weapon, consumeN int) {
	if forWeapon == nil {
		return
	}

	if a.ammoCount > 0 {
		// consume ammo amount based on weapon and projectile count
		ammoConsumed := consumeN
		switch a.ammoType {
		case AMMO_BALLISTIC:
			// ballistic are burst fire but still only consume one ammo count
			ammoConsumed = consumeN
		default:
			ammoConsumed = consumeN * forWeapon.ProjectileCount()
		}

		a.ammoCount -= ammoConsumed
		if a.ammoCount < 0 {
			a.ammoCount = 0
		}
	}
}

func PointsToVector2(points [][2]float64) []geom.Vector2 {
	vectors := make([]geom.Vector2, len(points))
	for i, p := range points {
		vectors[i] = geom.Vector2{X: p[0], Y: p[1]}
	}
	return vectors
}

func InArray(array, search any) bool {
	a := reflect.ValueOf(array)
	a = a.Convert(a.Type())
	t := reflect.TypeOf(array).Kind()

	switch t {
	case reflect.Slice, reflect.Array:
		for i := 0; i < a.Len(); i++ {
			if reflect.DeepEqual(search, a.Index(i).Interface()) {
				return true
			}
		}
	}

	return false
}
