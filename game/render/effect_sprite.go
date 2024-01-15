package render

import (
	"path"

	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go"
	"github.com/jinzhu/copier"
)

type EffectSprite struct {
	*Sprite
	LoopCount int
	AudioFile string

	AttachedTo    *Sprite
	AttachedDepth float64
}

func NewAnimatedEffect(
	r *model.ModelEffectResource, img *ebiten.Image, loopCount int,
) *EffectSprite {
	e := &EffectSprite{
		Sprite: NewAnimatedSprite(
			model.BasicVisualEntity(0, 0, 0, raycaster.AnchorCenter),
			r.Scale, img, r.ImageSheet.Columns, r.ImageSheet.Rows, r.ImageSheet.AnimationRate,
		),
		LoopCount: loopCount,
	}

	// effects cannot be focused upon by player reticle
	e.Sprite.focusable = false

	// effects self illuminate so they do not get dimmed in night conditions
	e.illumination = 5000

	// optional audio for sound effects (not for projectile impact effect audio)
	if len(r.Audio) > 0 {
		e.AudioFile = path.Join("audio/sfx/impacts", r.Audio)
	}

	return e
}

func (e *EffectSprite) Clone() *EffectSprite {
	fClone := &EffectSprite{}
	sClone := &Sprite{}
	eClone := &model.BasicEntity{}

	copier.Copy(fClone, e)
	copier.Copy(sClone, e.Sprite)
	copier.Copy(eClone, e.Entity)

	fClone.Sprite = sClone
	fClone.Sprite.Entity = eClone

	return fClone
}

func (e *EffectSprite) Update(camPos *geom.Vector2) {
	if e.AttachedTo != nil && e.AttachedDepth != 0 {
		// keep this effect moving along with sprite it is attached to
		x, y, z := e.AttachedTo.Pos().X, e.AttachedTo.Pos().Y, e.AttachedTo.PosZ()

		// use attached depth to set relative from camera position so it appears in front or behind attached
		camLine := &geom.Line{X1: camPos.X, Y1: camPos.Y, X2: x, Y2: y}
		angle, distance := camLine.Angle(), camLine.Distance()

		attachedLine := geom.LineFromAngle(camPos.X, camPos.Y, angle, distance+e.AttachedDepth)
		e.SetPos(&geom.Vector2{X: attachedLine.X2, Y: attachedLine.Y2})
		e.SetPosZ(z)
	} else {
		if e.Velocity() != 0 {
			ePos := e.Pos()
			trajectory := geom3d.Line3dFromAngle(ePos.X, ePos.Y, e.PosZ(), e.Heading(), e.Pitch(), e.Velocity())
			e.SetPos(&geom.Vector2{X: trajectory.X2, Y: trajectory.Y2})
			e.SetPosZ(trajectory.Z2)
		}

		if e.VelocityZ() != 0 {
			e.SetPosZ(e.PosZ() + e.VelocityZ())
		}
	}

	e.Sprite.Update(camPos)
}
