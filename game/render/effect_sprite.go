package render

import (
	"image/color"

	"github.com/harbdog/pixelmek-3d/game/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jinzhu/copier"
)

type EffectSprite struct {
	*Sprite
	LoopCount int
}

func NewAnimatedEffect(
	modelEntity model.Entity, scale float64, img *ebiten.Image, columns, rows, animationRate, loopCount int,
) *EffectSprite {
	mapColor := color.RGBA{0, 0, 0, 0}
	e := &EffectSprite{
		Sprite:    NewAnimatedSprite(modelEntity, scale, img, mapColor, columns, rows, animationRate),
		LoopCount: loopCount,
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
