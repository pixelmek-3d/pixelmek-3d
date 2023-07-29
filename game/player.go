package game

import (
	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/pixelmek-3d/game/render"

	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"

	log "github.com/sirupsen/logrus"
)

type StrideDirection uint

const (
	StrideUp StrideDirection = iota
	StrideDown
)

type Player struct {
	model.Unit
	sprite              *render.Sprite
	cameraZ             float64
	strideDir           StrideDirection
	strideZ             float64
	strideStomp         bool
	moved               bool
	convergenceDistance float64
	convergencePoint    *geom3d.Vector3
	convergenceSprite   *render.Sprite
	weaponGroups        [][]model.Weapon
	selectedWeapon      uint
	selectedGroup       uint
	fireMode            model.WeaponFireMode
	navPoint            *render.NavSprite
}

func NewPlayer(unit model.Unit, sprite *render.Sprite, x, y, z, angle, pitch float64) *Player {
	p := &Player{
		Unit:   unit,
		sprite: sprite,
		moved:  false,
	}

	p.SetAsPlayer(true)

	p.SetPos(&geom.Vector2{X: x, Y: y})
	p.SetPosZ(z)
	p.SetHeading(angle)
	p.SetPitch(pitch)
	p.SetVelocity(0)

	p.selectedWeapon = 0
	p.weaponGroups = make([][]model.Weapon, 3)
	for i := 0; i < cap(p.weaponGroups); i++ {
		p.weaponGroups[i] = make([]model.Weapon, 0, len(unit.Armament()))
	}
	// initialize all weapons as only in first weapon group
	p.weaponGroups[0] = append(p.weaponGroups[0], unit.Armament()...)

	// TODO: save/restore weapon groups for weapons per unit

	return p
}

func (p *Player) SetPosZ(z float64) {
	p.cameraZ = z + p.strideZ + p.Unit.CockpitOffset().Y // TODO: support cockpit offset in sprite X direction
	p.Unit.SetPosZ(z)
}

func (g *Game) SetPlayerUnit(unit model.Unit) {
	var unitSprite *render.Sprite

	var pX, pY, pZ, pH float64
	if g.player != nil {
		// handle in-mission player unit changes
		pX, pY = g.player.Pos().X, g.player.Pos().Y
		pZ = 0.0
		pH = g.player.Heading()
	}

	switch unitType := unit.(type) {
	case *model.Mech:
		unitSprite = g.createUnitSprite(unit).(*render.MechSprite).Sprite

	case *model.Vehicle:
		unitSprite = g.createUnitSprite(unit).(*render.VehicleSprite).Sprite

	case *model.VTOL:
		unitSprite = g.createUnitSprite(unit).(*render.VTOLSprite).Sprite
		if pZ < unit.CollisionHeight() {
			// for VTOL, adjust Z position to not be stuck in the ground
			pZ = unit.CollisionHeight()
		}

	case *model.Infantry:
		unitSprite = g.createUnitSprite(unit).(*render.InfantrySprite).Sprite

	default:
		log.Fatalf("unable to set player unit, resource type %s not handled", unitType)
		return
	}

	g.player = NewPlayer(unit, unitSprite, pX, pY, pZ, pH, 0)
	g.player.SetCollisionRadius(unit.CollisionRadius())
	g.player.SetCollisionHeight(unit.CollisionHeight())

	if unit.HasTurret() {
		g.mouseMode = MouseModeTurret
	} else {
		g.mouseMode = MouseModeBody
	}
}

func (p *Player) Update() bool {
	// handle player specific updates such as camera bobbing from movement
	switch p.Unit.(type) {
	case *model.Mech:
		resource := p.Unit.(*model.Mech).Resource
		// TODO: cap stride height for really tall mechs (or generally slower mechs?)
		maxStrideHeight := 0.1 * resource.Height / model.METERS_PER_UNIT // TODO: calculate this once on init
		velocity := p.Velocity()
		velocityMult := velocity / p.MaxVelocity()

		// TODO: handle stride effects from gravity != 1.0

		if p.PosZ() > 0 {
			if p.JumpJetsActive() {
				// jump jets on, settle view down to 0
				p.strideDir = StrideDown
			} else {
				// jump jets off, raise view due so when it hits ground gets effect going back to 0
				p.strideDir = StrideUp
			}
		} else {

			if velocity == 0 {
				p.strideDir = StrideDown
			} else {
				// cap stride height based on current velocity
				maxStrideHeight = (maxStrideHeight / 2) + velocityMult*(maxStrideHeight/2)
			}
		}

		// set stride delta based on current velocity and max stride height
		strideSeconds := 0.5 / velocityMult
		strideDelta := (2 * maxStrideHeight) / (strideSeconds * model.TICKS_PER_SECOND)

		// update player stride camera offset
		switch p.strideDir {
		case StrideUp:
			p.strideZ += strideDelta
		case StrideDown:
			p.strideZ -= strideDelta
		}

		// cap stride height effect on camera
		if p.strideZ > maxStrideHeight {
			p.strideZ = maxStrideHeight
			if p.PosZ() == 0 {
				p.strideDir = StrideDown
			}
		}
		if p.strideZ < 0 {
			p.strideZ = 0
			if p.PosZ() == 0 && velocity > 0 {
				p.strideDir = StrideUp
			}

			// foot hit the ground, make stompy sound
			p.strideStomp = true
		}
	}

	return p.Unit.Update()
}
