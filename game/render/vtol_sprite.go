package render

import (
	"image/color"
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
			vtol, scale, img, color.RGBA{},
		)
	} else {
		p = NewAnimatedSprite(vtol, scale, img, color.RGBA{}, sheet.Columns, sheet.Rows, 4)
		if len(sheet.AngleFacing) > 0 {
			facingMap := make(map[float64]int, len(sheet.AngleFacing))
			for degrees, index := range sheet.AngleFacing {
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
	eClone := &model.VTOL{}

	copier.Copy(vClone, v)
	copier.Copy(sClone, v.Sprite)
	copier.Copy(eClone, v.Entity)

	vClone.Sprite = sClone
	vClone.Sprite.Entity = eClone

	return vClone
}
