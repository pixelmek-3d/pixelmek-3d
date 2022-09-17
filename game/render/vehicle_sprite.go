package render

import (
	"image/color"
	_ "image/png"

	"github.com/harbdog/pixelmek-3d/game/model"

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
	p := NewSprite(
		vehicle, scale, img, color.RGBA{},
	)
	s := &VehicleSprite{
		Sprite: p,
	}

	return s
}

func (v *VehicleSprite) Clone() *VehicleSprite {
	vClone := &VehicleSprite{}
	sClone := &Sprite{}
	eClone := &model.Vehicle{}

	copier.Copy(vClone, v)
	copier.Copy(sClone, v.Sprite)
	copier.Copy(eClone, v.Entity)

	vClone.Sprite = sClone
	vClone.Sprite.Entity = eClone

	return vClone
}
