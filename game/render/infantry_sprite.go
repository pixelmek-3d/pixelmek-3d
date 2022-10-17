package render

import (
	"image/color"
	_ "image/png"

	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/raycaster-go/geom"

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
	var p *Sprite
	sheet := infantry.Resource.ImageSheet

	if sheet == nil {
		p = NewSprite(
			infantry, scale, img, color.RGBA{},
		)
	} else {
		p = NewAnimatedSprite(infantry, scale, img, color.RGBA{}, sheet.Columns, sheet.Rows, sheet.AnimationRate)
		if len(sheet.AngleFacingRow) > 0 {
			facingMap := make(map[float64]int, len(sheet.AngleFacingRow))
			for degrees, index := range sheet.AngleFacingRow {
				rads := geom.Radians(degrees)
				facingMap[rads] = index
			}
			p.SetTextureFacingMap(facingMap)
		}
	}

	s := &InfantrySprite{
		Sprite: p,
	}

	return s
}

func (t *InfantrySprite) Clone() *InfantrySprite {
	tClone := &InfantrySprite{}
	sClone := &Sprite{}
	eClone := t.Entity.Clone()

	copier.Copy(tClone, t)
	copier.Copy(sClone, t.Sprite)

	tClone.Sprite = sClone
	tClone.Sprite.Entity = eClone

	return tClone
}
