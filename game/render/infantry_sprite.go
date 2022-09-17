package render

import (
	"image/color"
	_ "image/png"

	"github.com/harbdog/pixelmek-3d/game/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jinzhu/copier"
)

type InfantrySprite struct {
	*Sprite

	// TODO: move to separate AI handler
	PatrolPathIndex int
	PatrolPath      [][2]float64
}

func NewInfantrySprite(
	infantry *model.Infantry, scale float64, img *ebiten.Image,
) *InfantrySprite {
	p := NewSprite(
		infantry, scale, img, color.RGBA{},
	)
	s := &InfantrySprite{
		Sprite: p,
	}

	return s
}

func (t *InfantrySprite) Clone() *InfantrySprite {
	tClone := &InfantrySprite{}
	sClone := &Sprite{}
	eClone := &model.Infantry{}

	copier.Copy(tClone, t)
	copier.Copy(sClone, t.Sprite)
	copier.Copy(eClone, t.Entity)

	tClone.Sprite = sClone
	tClone.Sprite.Entity = eClone

	return tClone
}
