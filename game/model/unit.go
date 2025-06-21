package model

import (
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
	"github.com/pixelmek-3d/pixelmek-3d/game/common"
)

const (
	UNIT_POWER_OFF_SECONDS float64 = 1.5
)

type UnitType int

const (
	MapUnitType UnitType = iota
	MechUnitType
	VehicleUnitType
	VTOLUnitType
	InfantryUnitType
	EmplacementUnitType
	TotalUnitTypes
)

type UnitObjective int

const (
	NonUnitObjective UnitObjective = iota
	DestroyUnitObjective
	ProtectUnitObjective
)

type UnitPowerStatus int

const (
	POWER_ON         UnitPowerStatus = 1
	POWER_OFF_MANUAL UnitPowerStatus = 0
	POWER_OFF_HEAT   UnitPowerStatus = -1
)

type Unit interface {
	Entity
	ID() string
	SetID(string)
	Team() int
	SetTeam(int)
	Name() string
	Variant() string
	Tonnage() float64
	UnitType() UnitType

	Heat() float64
	MaxHeat() float64
	HeatDissipation() float64
	OverHeated() bool
	Powered() UnitPowerStatus
	SetPowered(UnitPowerStatus)

	TriggerWeapon(Weapon) bool
	Target() Entity
	SetTarget(Entity)
	TargetLock() float64
	SetTargetLock(float64)

	TurnRate() float64
	TargetHeading() float64
	SetTargetHeading(float64)
	SetTargetPitch(float64)
	MaxVelocity() float64
	TargetVelocity() float64
	SetTargetVelocity(float64)
	TargetVelocityZ() float64
	SetTargetVelocityZ(float64)
	Update() bool

	HasTurret() bool
	TurretAngle() float64
	SetTurretAngle(float64)
	TurretRate() float64
	SetTargetTurretAngle(float64)
	MaxTurretExtentAngle() float64

	CockpitOffset() *geom.Vector2
	Ammunition() *Ammo
	Armament() []Weapon
	AddArmament(Weapon)

	JumpJets() int
	JumpJetsActive() bool
	SetJumpJetsActive(bool)
	JumpJetsDirectional() bool
	SetJumpJetsDirectional(bool)
	JumpJetHeading() float64
	SetJumpJetHeading(float64)
	JumpJetVelocity() float64
	JumpJetDuration() float64
	MaxJumpJetDuration() float64

	GuardArea() *geom.Circle
	SetGuardArea(x, y, radius float64)
	GuardUnit() string
	SetGuardUnit(string)
	PathStack() *common.FIFOStack[geom.Vector2]
	SetPatrolPath([]geom.Vector2)
	WithdrawArea() *Rect
	SetWithdrawArea(*Rect)

	Objective() UnitObjective
	SetObjective(UnitObjective)
	SetAsPlayer(bool)
	IsPlayer() bool

	CloneUnit() Unit
}

type UnitModel struct {
	id                  string
	name                string
	variant             string
	team                int
	unitType            UnitType
	position            *geom.Vector2
	positionZ           float64
	anchor              raycaster.SpriteAnchor
	heading             float64
	targetHeading       float64
	maxTurnRate         float64
	pitch               float64
	targetPitch         float64
	hasTurret           bool
	turretAngle         float64
	targetTurretAngle   float64
	maxTurretRate       float64
	maxTurretExtent     float64
	velocity            float64
	velocityZ           float64
	targetVelocity      float64
	targetVelocityZ     float64
	maxVelocity         float64
	collisionRadius     float64
	collisionHeight     float64
	cockpitOffset       *geom.Vector2
	armor               float64
	structure           float64
	heat                float64
	heatDissipation     float64
	heatSinks           int
	heatSinkType        HeatSinkType
	powered             UnitPowerStatus
	armament            []Weapon
	ammunition          *Ammo
	jumpJets            int
	jumpJetsActive      bool
	jumpJetsDirectional bool
	jumpJetHeading      float64
	jumpJetVelocity     float64
	jumpJetDelay        float64
	jumpJetDuration     float64
	maxJumpJetDuration  float64
	target              Entity
	targetLock          float64
	objective           UnitObjective
	guardArea           *geom.Circle
	guardUnit           string
	pathStack           *common.FIFOStack[geom.Vector2]
	withdrawArea        *Rect
	parent              Entity
	isPlayer            bool
}

func EntityUnit(entity Entity) Unit {
	if unit, ok := entity.(Unit); ok {
		return unit
	}
	return nil
}

func IsEntityUnit(entity Entity) bool {
	_, ok := entity.(Unit)
	return ok
}

func (e1 *UnitModel) DistanceToEntity(e2 Entity) float64 {
	pos1, pos2 := e1.Pos(), e2.Pos()
	x1, y1, z1 := pos1.X, pos1.Y, e1.PosZ()
	x2, y2, z2 := pos2.X, pos2.Y, e2.PosZ()
	line := geom3d.Line3d{
		X1: x1, Y1: y1, Z1: z1,
		X2: x2, Y2: y2, Z2: z2,
	}
	return line.Distance()
}

func (e *UnitModel) ID() string {
	return e.id
}

func (e *UnitModel) SetID(id string) {
	if len(id) == 0 {
		// use random uuid if not given static id reference
		e.id = strings.ReplaceAll(strings.ToLower(e.variant), " ", "-") + "_" + uuid.NewString()
		return
	}
	e.id = id
}

func (e *UnitModel) Name() string {
	return e.name
}

func (e *UnitModel) Variant() string {
	return e.variant
}

func (e *UnitModel) Team() int {
	return e.team
}

func (e *UnitModel) SetTeam(team int) {
	e.team = team
}

func (e *UnitModel) UnitType() UnitType {
	return e.unitType
}

func (e *UnitModel) Pitch() float64 {
	return e.pitch
}

func (e *UnitModel) Heat() float64 {
	return e.heat
}

func (e *UnitModel) MaxHeat() float64 {
	// determine based on unit type and # of heat sinks
	switch e.unitType {
	case MechUnitType:
		return 50 + float64(e.heatSinks)
	case VehicleUnitType:
		return 50 + float64(e.heatSinks)
	case VTOLUnitType:
		return 50 + float64(e.heatSinks)
	case InfantryUnitType:
		return float64(e.heatSinks)
	case EmplacementUnitType:
		return 100 + float64(e.heatSinks)
	}
	return 1
}

func (e *UnitModel) OverHeated() bool {
	if e.powered == POWER_OFF_HEAT {
		// resuming from auto shutdown overheat status requires under 70% of max heat
		return e.heat > 0.7*e.MaxHeat()
	}
	return e.heat > e.MaxHeat()
}

func (e *UnitModel) HeatDissipation() float64 {
	return e.heatDissipation
}

func (e *UnitModel) Powered() UnitPowerStatus {
	return e.powered
}

func (e *UnitModel) SetPowered(powered UnitPowerStatus) {
	e.powered = powered
}

func (e *UnitModel) TriggerWeapon(w Weapon) bool {
	if e.powered != POWER_ON || w.Cooldown() > 0 {
		return false
	}

	missileWeapon, isMissile := w.(*MissileWeapon)
	if isMissile && missileWeapon.IsLockOnLockRequired() {
		// for Streak SRMs that require target lock to fire
		if e.target == nil || e.targetLock < 1.0 {
			return false
		}

		// target must be in weapon range
		targetDistance := e.DistanceToEntity(e.target) - e.CollisionRadius() - e.target.CollisionRadius()
		weaponRange := w.Distance() / METERS_PER_UNIT
		if int(targetDistance) > int(weaponRange) {
			return false
		}
	}

	w.TriggerCooldown()
	e.heat += w.Heat()
	return true
}

func (e *UnitModel) Target() Entity {
	return e.target
}

func (e *UnitModel) SetTarget(t Entity) {
	if e.target != t && e.targetLock != 0 {
		e.SetTargetLock(0)
	}
	e.target = t
}

func (e *UnitModel) TargetLock() float64 {
	return e.targetLock
}

func (e *UnitModel) SetTargetLock(lockPercent float64) {
	e.targetLock = lockPercent
}

func (e *UnitModel) HasTurret() bool {
	return e.hasTurret
}

func (e *UnitModel) TurretAngle() float64 {
	if e.hasTurret {
		return e.turretAngle
	}
	return e.heading
}

func (e *UnitModel) SetTurretAngle(angle float64) {
	if e.hasTurret {
		e.turretAngle = angle
	} else {
		e.SetHeading(angle)
	}
}

func (e *UnitModel) SetTargetTurretAngle(angle float64) {
	if e.hasTurret {
		e.targetTurretAngle = angle
	} else {
		e.SetTargetHeading(angle)
	}
}

func (e *UnitModel) TurretRate() float64 {
	if e.hasTurret {
		return e.maxTurretRate
	}
	return e.maxTurnRate
}

func (e *UnitModel) MaxTurretExtentAngle() float64 {
	if e.hasTurret {
		return e.maxTurretExtent
	}
	return 0
}

func (e *UnitModel) Ammunition() *Ammo {
	return e.ammunition
}

func (e *UnitModel) Armament() []Weapon {
	return e.armament
}

func (e *UnitModel) AddArmament(w Weapon) {
	e.armament = append(e.armament, w)
}

func (e *UnitModel) Pos() *geom.Vector2 {
	return e.position
}

func (e *UnitModel) SetPos(pos *geom.Vector2) {
	e.position = pos
}

func (e *UnitModel) PosZ() float64 {
	return e.positionZ
}

func (e *UnitModel) SetPosZ(posZ float64) {
	e.positionZ = posZ
}

func (e *UnitModel) Anchor() raycaster.SpriteAnchor {
	return e.anchor
}

func (e *UnitModel) SetAnchor(anchor raycaster.SpriteAnchor) {
	e.anchor = anchor
}

func (e *UnitModel) Heading() float64 {
	return e.heading
}

func (e *UnitModel) SetHeading(angle float64) {
	e.heading = angle
}

func (e *UnitModel) TargetHeading() float64 {
	return e.targetHeading
}

func (e *UnitModel) SetTargetHeading(heading float64) {
	e.targetHeading = heading
}

func (e *UnitModel) SetPitch(pitch float64) {
	e.pitch = pitch
}

func (e *UnitModel) SetTargetPitch(pitch float64) {
	e.targetPitch = pitch
}

func (e *UnitModel) TurnRate() float64 {
	if e.velocity == 0 {
		return e.maxTurnRate
	}

	// dynamic turn rate is half of the max turn rate when at max velocity
	vTurnRatio := 0.5 + 0.5*(e.maxVelocity-math.Abs(e.velocity))/e.maxVelocity
	return e.maxTurnRate * vTurnRatio
}

func (e *UnitModel) Velocity() float64 {
	return e.velocity
}

func (e *UnitModel) SetVelocity(velocity float64) {
	e.velocity = velocity
}

func (e *UnitModel) VelocityZ() float64 {
	return e.velocityZ
}

func (e *UnitModel) SetVelocityZ(velocityZ float64) {
	e.velocityZ = velocityZ
}

func (e *UnitModel) MaxVelocity() float64 {
	return e.maxVelocity
}

func (e *UnitModel) TargetVelocity() float64 {
	return e.targetVelocity
}

func (e *UnitModel) SetTargetVelocity(tVelocity float64) {
	maxV := e.MaxVelocity()
	if tVelocity > maxV {
		tVelocity = maxV
	} else if tVelocity < -maxV/2 {
		tVelocity = -maxV / 2
	}
	e.targetVelocity = tVelocity
}

func (e *UnitModel) TargetVelocityZ() float64 {
	return e.targetVelocityZ
}

func (e *UnitModel) SetTargetVelocityZ(tVelocityZ float64) {
	e.targetVelocityZ = tVelocityZ
}

func (e *UnitModel) CollisionRadius() float64 {
	return e.collisionRadius
}

func (e *UnitModel) SetCollisionRadius(collisionRadius float64) {
	e.collisionRadius = collisionRadius
}

func (e *UnitModel) CollisionHeight() float64 {
	return e.collisionHeight
}

func (e *UnitModel) SetCollisionHeight(collisionHeight float64) {
	e.collisionHeight = collisionHeight
}

func (e *UnitModel) CockpitOffset() *geom.Vector2 {
	return e.cockpitOffset
}

func (e *UnitModel) ApplyDamage(damage float64) {
	if e.armor > 0 {
		e.armor -= damage
		if e.armor < 0 {
			// apply remainder of armor damage on structure
			e.structure -= math.Abs(e.armor)
			e.armor = 0
		}
	} else {
		e.structure -= damage
	}
}

func (e *UnitModel) ArmorPoints() float64 {
	return e.armor
}

func (e *UnitModel) SetArmorPoints(armor float64) {
	e.armor = armor
}

func (e *UnitModel) StructurePoints() float64 {
	return e.structure
}

func (e *UnitModel) SetStructurePoints(structure float64) {
	e.structure = structure
}

func (e *UnitModel) IsDestroyed() bool {
	return e.structure <= 0
}

func (e *UnitModel) JumpJets() int {
	return e.jumpJets
}

func (e *UnitModel) JumpJetsActive() bool {
	return e.jumpJetsActive
}

func (e *UnitModel) SetJumpJetsActive(active bool) {
	e.jumpJetsActive = active
	if !active {
		e.SetTargetVelocityZ(0)
		e.SetJumpJetsDirectional(false)
	}
}

func (e *UnitModel) JumpJetsDirectional() bool {
	return e.jumpJetsDirectional
}

func (e *UnitModel) SetJumpJetsDirectional(isDirectional bool) {
	e.jumpJetsDirectional = isDirectional
}

func (e *UnitModel) JumpJetHeading() float64 {
	return e.jumpJetHeading
}

func (e *UnitModel) SetJumpJetHeading(heading float64) {
	e.jumpJetHeading = heading
}

func (e *UnitModel) JumpJetVelocity() float64 {
	return e.jumpJetVelocity
}

func (e *UnitModel) JumpJetDuration() float64 {
	return e.jumpJetDuration
}

func (e *UnitModel) MaxJumpJetDuration() float64 {
	return e.maxJumpJetDuration
}

func (e *UnitModel) GuardArea() *geom.Circle {
	return e.guardArea
}

func (e *UnitModel) SetGuardArea(x, y, radius float64) {
	if radius < 0 {
		e.guardArea = nil
		return
	} else {
		e.guardArea = &geom.Circle{X: x, Y: y, Radius: radius}
	}

	// initialize empty path stack for use in guard area behavior
	e.pathStack = common.NewFIFOStack[geom.Vector2]()
}

func (e *UnitModel) GuardUnit() string {
	return e.guardUnit
}

func (e *UnitModel) SetGuardUnit(unit string) {
	e.guardUnit = unit

	// initialize empty path stack for use in guard unit behavior
	e.pathStack = common.NewFIFOStack[geom.Vector2]()
}

func (e *UnitModel) SetPatrolPath(modelPatrolPath []geom.Vector2) {
	e.pathStack = common.NewFIFOStack[geom.Vector2]()
	for _, point := range modelPatrolPath {
		e.pathStack.Push(point)
	}
}

func (e *UnitModel) PathStack() *common.FIFOStack[geom.Vector2] {
	return e.pathStack
}

func (e *UnitModel) WithdrawArea() *Rect {
	return e.withdrawArea
}

func (e *UnitModel) SetWithdrawArea(withdrawArea *Rect) {
	e.withdrawArea = withdrawArea

	// initialize empty path stack for use in withdraw area behavior
	e.pathStack = common.NewFIFOStack[geom.Vector2]()
}

func (e *UnitModel) Objective() UnitObjective {
	return e.objective
}

func (e *UnitModel) SetObjective(objective UnitObjective) {
	e.objective = objective
}

func (e *UnitModel) Parent() Entity {
	return e.parent
}

func (e *UnitModel) SetParent(parent Entity) {
	e.parent = parent
}

func (e *UnitModel) SetAsPlayer(isPlayer bool) {
	e.isPlayer = isPlayer
}

func (e *UnitModel) IsPlayer() bool {
	return e.isPlayer
}

func (e *UnitModel) needsUpdate() bool {
	if e.jumpJetsActive || e.heat > 0 ||
		e.targetHeading != e.heading || e.targetPitch != e.pitch ||
		e.targetTurretAngle != e.turretAngle ||
		e.targetVelocity != 0 || e.velocity != 0 ||
		e.targetVelocityZ != 0 || e.velocityZ != 0 || e.positionZ != 0 {
		return true
	}
	return false
}

func (e *UnitModel) update() {
	if e.jumpJetsActive {
		// apply heat from active jump jets
		e.heat += 2 * float64(e.jumpJets) / TICKS_PER_SECOND
		// for balance purposes, not allowing heat dissipation while jets enabled
	} else if e.heat > 0 {
		// apply heat dissipation
		e.heat -= e.HeatDissipation()
		if e.heat < 0 {
			e.heat = 0
		}
	}

	// if not powered on, no movement related updates needed
	if e.powered != POWER_ON {
		return
	}

	turnRate := e.TurnRate()
	turretRate := e.TurretRate()

	var deltaH float64
	if e.targetHeading != e.heading {
		// move towards target heading amount allowed by turn rate
		deltaH = geom.Clamp(AngleDistance(e.heading, e.targetHeading), -turnRate, turnRate)
		e.heading = ClampAngle2Pi(e.heading + deltaH)
		if math.Abs(deltaH) < math.Abs(turnRate) && geom.NearlyEqual(e.targetHeading, e.heading, 0.0001) {
			e.heading = e.targetHeading
		}

		if e.jumpJets > 0 && e.jumpJetsActive {
			// set jump jet heading only while jumping
			e.jumpJetHeading = e.heading
		}

		if e.isPlayer {
			// offset player turret angle so it does not have to play catch up
			e.targetTurretAngle -= deltaH
		}
	}

	if e.targetPitch != e.pitch {
		// move towards target pitch amount allowed by turret rate
		distP := AngleDistance(e.pitch, e.targetPitch)

		pitchRate := turretRate
		if e.isPlayer {
			// use logarithmic scale to smooth the approach to the target pitch angle
			pitchRate = math.Log1p(2*math.Abs(distP)) * turretRate
		}

		deltaP := geom.Clamp(distP, -pitchRate, pitchRate)
		e.pitch = ClampAngle(e.pitch + deltaP)
		if math.Abs(deltaP) < math.Abs(pitchRate) && geom.NearlyEqual(e.targetPitch, e.pitch, 0.0001) {
			e.pitch = e.targetPitch
		}
	}

	if e.hasTurret {
		// determine if turret angle needs to be bound by its maximum extent from unit heading
		tDist := AngleDistance(e.heading, e.targetTurretAngle)
		tExtent := e.maxTurretExtent
		switch {
		case tDist < -tExtent:
			switch {
			case tDist < -tExtent:
				e.targetTurretAngle = ClampAngle2Pi(e.heading - tExtent)
			case tDist > tExtent:
				e.targetTurretAngle = ClampAngle2Pi(e.heading + tExtent)
			}
		}

		if e.targetTurretAngle != e.turretAngle {
			// move towards target turret angle amount allowed by turret rate
			distA := AngleDistance(e.turretAngle, e.targetTurretAngle)

			twistRate := turretRate
			if e.isPlayer {
				// use logarithmic scale to smooth the approach to the target turret angle
				twistRate = math.Log1p(2*math.Abs(distA)) * turretRate
			}

			deltaA := geom.Clamp(distA, -twistRate, twistRate)
			e.turretAngle = ClampAngle2Pi(e.turretAngle + deltaA + deltaH)
			if math.Abs(deltaA) < math.Abs(twistRate) && geom.NearlyEqual(e.targetTurretAngle, e.turretAngle, 0.0001) {
				e.turretAngle = e.targetTurretAngle
			}
		}
	}
}
