package model

import (
	"image/color"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"
)

type MechSprite struct {
	*Sprite
	mechAnimate  *MechSpriteAnimate
	animateIndex MechAnimationIndex
	// static *ebiten.Image
	// ct     *ebiten.Image
	// la     *ebiten.Image
	// ra     *ebiten.Image
	// ll     *ebiten.Image
	// rl     *ebiten.Image
}

type MechPart int

const (
	PART_STATIC MechPart = 0
	PART_CT     MechPart = 1
	PART_LA     MechPart = 2
	PART_RA     MechPart = 3
	PART_LL     MechPart = 4
	PART_RL     MechPart = 5
	NUM_PARTS   MechPart = 6
)

// func (s *MechSprite) Scale() float64 {
// 	return s.scale
// }

// func (s *MechSprite) VerticalAnchor() raycaster.SpriteAnchor {
// 	return s.anchor
// }

// func (s *MechSprite) Texture() *ebiten.Image {
// 	return s.texture
// }

// func (s *MechSprite) TextureRect() image.Rectangle {
// 	return s.texRect
// }

func NewMechSprite(
	x, y, scale float64, img *ebiten.Image, collisionRadius float64,
) *MechSprite {
	// all mech sprite sheets have 6 columns of images in the sheet:
	// [full, torso, left arm, right arm, left leg, right leg]
	mechAnimate := NewMechAnimationSheetFromImage(img)
	p := NewAnimatedSprite(
		x, y, scale, 0, mechAnimate.sheet, color.RGBA{},
		mechAnimate.maxCols, mechAnimate.maxRows, raycaster.AnchorBottom, collisionRadius,
	)
	s := &MechSprite{
		Sprite:       p,
		mechAnimate:  mechAnimate,
		animateIndex: ANIMATE_STATIC,
	}

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

func (s *MechSprite) SetMechAnimation(animateIndex MechAnimationIndex) {
	s.animateIndex = animateIndex
	s.ResetAnimation()
}

func (s *MechSprite) ResetAnimation() {
	s.animCounter = 0
	s.loopCounter = 0

	switch {
	case s.animateIndex <= ANIMATE_STATIC:
		s.texNum = 0
	case s.animateIndex > ANIMATE_STATIC:
		animRow := int(s.animateIndex)
		s.texNum = animRow * s.mechAnimate.maxCols
	}
}

func (s *MechSprite) Update(camPos *geom.Vector2) {
	if s.AnimationRate <= 0 {
		return
	}
	if s.animateIndex <= ANIMATE_STATIC {
		s.texNum = 0
		return
	}
	if s.animCounter >= s.AnimationRate {
		animRow := int(s.animateIndex)

		minTexNum := animRow * s.mechAnimate.maxCols
		maxTexNum := minTexNum + s.mechAnimate.numColsAtRow[animRow] - 1

		s.animCounter = 0

		if s.animReversed {
			s.texNum -= 1
			if s.texNum > maxTexNum || s.texNum < minTexNum {
				s.texNum = maxTexNum
				s.loopCounter++
			}
		} else {
			s.texNum += 1
			if s.texNum > maxTexNum || s.texNum < minTexNum {
				s.texNum = minTexNum
				s.loopCounter++
			}
		}
	} else {
		s.animCounter++
	}
}
