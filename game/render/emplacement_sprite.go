package render

import (
	_ "image/png"

	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/raycaster-go/geom"

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
	}

	s := &EmplacementSprite{
		Sprite: p,
	}

	return s
}

func (t *EmplacementSprite) Clone() *EmplacementSprite {
	tClone := &EmplacementSprite{}
	sClone := &Sprite{}
	eClone := t.Entity.Clone()

	copier.Copy(tClone, t)
	copier.Copy(sClone, t.Sprite)

	tClone.Sprite = sClone
	tClone.Sprite.Entity = eClone

	return tClone
}
