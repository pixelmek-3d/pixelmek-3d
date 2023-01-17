package render

import (
	"github.com/harbdog/pixelmek-3d/game/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go"
	"github.com/jinzhu/copier"
)

type EffectSprite struct {
	*Sprite
	LoopCount int
}

func NewAnimatedEffect(
	scale float64, img *ebiten.Image, columns, rows, animationRate, loopCount int,
) *EffectSprite {
	e := &EffectSprite{
		Sprite: NewAnimatedSprite(
			model.BasicVisualEntity(0, 0, 0, raycaster.AnchorCenter),
			scale, img, columns, rows, animationRate,
		),
		LoopCount: loopCount,
	}

	// effects cannot be focused upon by player reticle
	e.Sprite.Focusable = false

	// effects self illuminate so they do not get dimmed in night conditions
	e.illumination = 5000

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
