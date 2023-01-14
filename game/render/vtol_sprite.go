package render

import (
	_ "image/png"

	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/raycaster-go/geom"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jinzhu/copier"
)

type VTOLSprite struct {
	*Sprite

	// TODO: move to separate AI handler
	PatrolPathIndex int
	PatrolPath      [][2]float64
}

func NewVTOLSprite(
	vtol *model.VTOL, scale float64, img *ebiten.Image,
) *VTOLSprite {
	var p *Sprite
	sheet := vtol.Resource.ImageSheet
	if sheet == nil {
		p = NewSprite(
			vtol, scale, img,
		)
	} else {
		p = NewAnimatedSprite(vtol, scale, img, sheet.Columns, sheet.Rows, sheet.AnimationRate)
		if len(sheet.AngleFacingRow) > 0 {
			facingMap := make(map[float64]int, len(sheet.AngleFacingRow))
			for degrees, index := range sheet.AngleFacingRow {
				rads := geom.Radians(degrees)
				facingMap[rads] = index
			}
			p.SetTextureFacingMap(facingMap)
		}
	}

	s := &VTOLSprite{
		Sprite: p,
	}

	return s
}

func (v *VTOLSprite) Clone() *VTOLSprite {
	vClone := &VTOLSprite{}
	sClone := &Sprite{}
	eClone := v.Entity.Clone()

	copier.Copy(vClone, v)
	copier.Copy(sClone, v.Sprite)

	vClone.Sprite = sClone
	vClone.Sprite.Entity = eClone

	return vClone
}
