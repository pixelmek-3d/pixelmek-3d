package model

import (
	"image"
	"image/color"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"
)

type MechSprite struct {
	*Sprite
	texture *ebiten.Image
	texRect image.Rectangle
	static  *ebiten.Image
	ct      *ebiten.Image
	la      *ebiten.Image
	ra      *ebiten.Image
	ll      *ebiten.Image
	rl      *ebiten.Image
}

type MechPart int

const (
	STATIC MechPart = 0
	CT     MechPart = 1
	LA     MechPart = 2
	RA     MechPart = 3
	LL     MechPart = 4
	RL     MechPart = 5
)

func (s *MechSprite) Scale() float64 {
	return s.scale
}

func (s *MechSprite) VerticalOffset() float64 {
	return s.vOffset
}

func (s *MechSprite) Texture() *ebiten.Image {
	return s.texture
}

func (s *MechSprite) TextureRect() image.Rectangle {
	return s.texRect
}

func NewMechSprite(
	x, y float64, img *ebiten.Image, collisionRadius float64,
) *MechSprite {
	// all mech sprite sheets have 6 columns of images in the sheet:
	// [full, torso, left arm, right arm, left leg, right leg]
	p := NewSpriteFromSheet(x, y, 1.0, img, color.RGBA{}, 6, 1, 0, 0, collisionRadius)

	// mech texture for raycasting must be square
	tSize := p.W
	if p.H > p.W {
		tSize = p.H
	}

	s := &MechSprite{
		Sprite:  p,
		texture: ebiten.NewImage(tSize, tSize),
		texRect: image.Rect(0, 0, tSize-1, tSize-1),
		static:  p.textures[STATIC],
		ct:      p.textures[CT],
		la:      p.textures[LA],
		ra:      p.textures[RA],
		ll:      p.textures[LL],
		rl:      p.textures[RL],
	}

	// TESTING: start by just drawing as static image
	s.texture.Clear()
	op := &ebiten.DrawImageOptions{}
	centerX, bottomY := float64(tSize/2-s.W/2), float64(tSize-s.H)
	op.GeoM.Translate(centerX, bottomY)
	s.texture.DrawImage(s.ct, op)
	s.texture.DrawImage(s.la, op)
	s.texture.DrawImage(s.ra, op)
	s.texture.DrawImage(s.ll, op)
	s.texture.DrawImage(s.rl, op)

	return s
}

func NewMechSpriteFromMech(x, y float64, origMech *MechSprite) *MechSprite {
	s := &MechSprite{}
	copier.Copy(s, origMech)

	s.Sprite = &Sprite{}
	copier.Copy(s.Sprite, origMech.Sprite)

	s.Position = &geom.Vector2{X: x, Y: y}

	return s
}
