package sprites

import (
	"github.com/harbdog/raycaster-go/geom"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"

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
			infantry, scale, img,
		)
	} else {
		p = NewAnimatedSprite(infantry, scale, img, sheet.Columns, sheet.Rows, sheet.AnimationRate)
		if len(sheet.AngleFacingRow) > 0 {
			facingMap := make(map[float64]int, len(sheet.AngleFacingRow))
			for degrees, index := range sheet.AngleFacingRow {
				rads := geom.Radians(degrees)
				facingMap[rads] = index
			}
			p.SetTextureFacingMap(facingMap)
		}
		p.staticTexNum = infantry.Resource.ImageSheet.StaticIndex
	}

	s := &InfantrySprite{
		Sprite: p,
	}

	return s
}

func (t *InfantrySprite) Infantry() *model.Infantry {
	if t.Entity == nil {
		return nil
	}
	return t.Entity.(*model.Infantry)
}

func (t *InfantrySprite) Clone(asUnit model.Unit) *InfantrySprite {
	tClone := &InfantrySprite{}
	sClone := &Sprite{}

	copier.Copy(tClone, t)
	copier.Copy(sClone, t.Sprite)

	tClone.Sprite = sClone

	if asUnit == nil {
		eClone := t.Entity.Clone()
		tClone.Sprite.Entity = eClone
	} else {
		tClone.Sprite.Entity = asUnit
	}

	return tClone
}
