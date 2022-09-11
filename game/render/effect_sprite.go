package render

import (
	"image/color"

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
	x, y, scale float64, img *ebiten.Image, columns, rows, animationRate int, anchor raycaster.SpriteAnchor, loopCount int,
) *EffectSprite {
	mapColor := color.RGBA{0, 0, 0, 0}
	e := &EffectSprite{
		Sprite:    NewAnimatedSprite(x, y, scale, img, mapColor, columns, rows, animationRate, anchor, 0, 0),
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
