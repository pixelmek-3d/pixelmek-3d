package render

import (
	"github.com/harbdog/raycaster-go/geom"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"

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
		p.staticTexNum = vtol.Resource.ImageSheet.StaticIndex
	}

	s := &VTOLSprite{
		Sprite: p,
	}

	return s
}

func (v *VTOLSprite) VTOL() *model.VTOL {
	if v.Entity == nil {
		return nil
	}
	return v.Entity.(*model.VTOL)
}

func (v *VTOLSprite) Clone(asUnit model.Unit) *VTOLSprite {
	vClone := &VTOLSprite{}
	sClone := &Sprite{}

	copier.Copy(vClone, v)
	copier.Copy(sClone, v.Sprite)

	vClone.Sprite = sClone

	if asUnit == nil {
		eClone := v.Entity.Clone()
		vClone.Sprite.Entity = eClone
	} else {
		vClone.Sprite.Entity = asUnit
	}

	return vClone
}
