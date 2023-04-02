package render

import (
	_ "image/png"

	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/raycaster-go/geom"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jinzhu/copier"
)

type VehicleSprite struct {
	*Sprite

	// TODO: move to separate AI handler
	PatrolPathIndex int
	PatrolPath      [][2]float64
}

func NewVehicleSprite(
	vehicle *model.Vehicle, scale float64, img *ebiten.Image,
) *VehicleSprite {
	var p *Sprite
	sheet := vehicle.Resource.ImageSheet

	if sheet == nil {
		p = NewSprite(
			vehicle, scale, img,
		)
	} else {
		p = NewAnimatedSprite(vehicle, scale, img, sheet.Columns, sheet.Rows, sheet.AnimationRate)
		if len(sheet.AngleFacingRow) > 0 {
			facingMap := make(map[float64]int, len(sheet.AngleFacingRow))
			for degrees, index := range sheet.AngleFacingRow {
				rads := geom.Radians(degrees)
				facingMap[rads] = index
			}
			p.SetTextureFacingMap(facingMap)
		}
		p.staticTexNum = vehicle.Resource.ImageSheet.StaticIndex
	}

	s := &VehicleSprite{
		Sprite: p,
	}

	return s
}

func (v *VehicleSprite) Clone() *VehicleSprite {
	vClone := &VehicleSprite{}
	sClone := &Sprite{}
	eClone := v.Entity.Clone()

	copier.Copy(vClone, v)
	copier.Copy(sClone, v.Sprite)

	vClone.Sprite = sClone
	vClone.Sprite.Entity = eClone

	return vClone
}
