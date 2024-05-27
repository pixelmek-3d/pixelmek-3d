package render

import (
	"github.com/harbdog/raycaster-go/geom"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jinzhu/copier"
)

type EmplacementSprite struct {
	*Sprite
}

func NewEmplacementSprite(
	emplacement *model.Emplacement, scale float64, img *ebiten.Image,
) *EmplacementSprite {
	var p *Sprite
	sheet := emplacement.Resource.ImageSheet

	if sheet == nil {
		p = NewSprite(
			emplacement, scale, img,
		)
	} else {
		p = NewAnimatedSprite(emplacement, scale, img, sheet.Columns, sheet.Rows, sheet.AnimationRate)
		if len(sheet.AngleFacingRow) > 0 {
			facingMap := make(map[float64]int, len(sheet.AngleFacingRow))
			for degrees, index := range sheet.AngleFacingRow {
				rads := geom.Radians(degrees)
				facingMap[rads] = index
			}
			p.SetTextureFacingMap(facingMap)
		}
		p.staticTexNum = emplacement.Resource.ImageSheet.StaticIndex
	}

	s := &EmplacementSprite{
		Sprite: p,
	}

	return s
}

func (t *EmplacementSprite) Emplacement() *model.Emplacement {
	if t.Entity == nil {
		return nil
	}
	return t.Entity.(*model.Emplacement)
}

func (t *EmplacementSprite) Clone(asUnit model.Unit) *EmplacementSprite {
	tClone := &EmplacementSprite{}
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
