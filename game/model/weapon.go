package model

import (
	"github.com/harbdog/raycaster-go/geom"
)

type WeaponType int

const (
	ENERGY WeaponType = iota
	BALLISTIC
	MISSILE
)

type Weapon interface {
	Name() string
	ShortName() string
	Type() WeaponType
	Tonnage() float64
	Damage() float64
	Heat() float64
	Distance() float64
	Velocity() float64
	Cooldown() float64
	Offset() *geom.Vector2
	Projectile() Projectile
	Parent() Entity
}

type EnergyWeapon struct {
	name       string
	short      string
	weaponType WeaponType
	tonnage    float64
	damage     float64
	heat       float64
	distance   float64
	velocity   float64
	cooldown   float64
	offset     *geom.Vector2
	projectile Projectile
	parent     Entity
}

func (w *EnergyWeapon) Name() string {
	return w.name
}

func (w *EnergyWeapon) ShortName() string {
	return w.short
}

func (w *EnergyWeapon) Type() WeaponType {
	return w.weaponType
}

func (w *EnergyWeapon) Tonnage() float64 {
	return w.tonnage
}

func (w *EnergyWeapon) Damage() float64 {
	return w.damage
}

func (w *EnergyWeapon) Heat() float64 {
	return w.heat
}

func (w *EnergyWeapon) Distance() float64 {
	return w.distance
}

func (w *EnergyWeapon) Velocity() float64 {
	return w.velocity
}

func (w *EnergyWeapon) Cooldown() float64 {
	return w.cooldown
}

func (w *EnergyWeapon) Offset() *geom.Vector2 {
	return w.offset
}

func (w *EnergyWeapon) Projectile() Projectile {
	return w.projectile
}

func (w *EnergyWeapon) Parent() Entity {
	return w.parent
}
