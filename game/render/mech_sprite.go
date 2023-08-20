package render

import (
	_ "image/png"

	"github.com/harbdog/pixelmek-3d/game/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"
)

type MechSprite struct {
	*Sprite
	mechAnimate  *MechSpriteAnimate
	animateIndex MechAnimationIndex
	strideStomp  bool

	// TODO: move to separate AI handler
	PatrolPathIndex int
	PatrolPath      [][2]float64
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

func NewMechSprite(
	mech *model.Mech, scale float64, img *ebiten.Image,
) *MechSprite {
	// all mech sprite sheets have 6 columns of images in the sheet:
	// [full, torso, left arm, right arm, left leg, right leg]
	mechAnimate := NewMechAnimationSheetFromImage(img)
	p := NewAnimatedSprite(
		mech, scale, mechAnimate.sheet,
		mechAnimate.maxCols, mechAnimate.maxRows, 0,
	)
	s := &MechSprite{
		Sprite:       p,
		mechAnimate:  mechAnimate,
		animateIndex: ANIMATE_STATIC,
	}

	return s
}

func (m *MechSprite) Clone(asUnit model.Unit) *MechSprite {
	mClone := &MechSprite{}
	sClone := &Sprite{}

	copier.Copy(mClone, m)
	copier.Copy(sClone, m.Sprite)

	mClone.Sprite = sClone

	if asUnit == nil {
		eClone := m.Entity.Clone()
		mClone.Sprite.Entity = eClone
	} else {
		mClone.Sprite.Entity = asUnit
	}

	return mClone
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

func (s *MechSprite) StrideStomp() bool {
	return s.strideStomp
}

func (s *MechSprite) ResetStrideStomp() {
	s.strideStomp = false
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

		if s.animateIndex == ANIMATE_STRUT {
			// use texture index for when the stomp audio occurs
			if s.texNum == minTexNum || s.texNum == minTexNum+(maxTexNum-minTexNum)/2 {
				s.strideStomp = true
			}
		}
	} else {
		s.animCounter++
	}
}
