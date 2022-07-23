package model

import (
	"image"
	"image/color"
	_ "image/png"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Sprite struct {
	*Entity
	scale          float64
	anchor         raycaster.SpriteAnchor
	W, H           int
	AnimationRate  int
	animReversed   bool
	animCounter    int
	loopCounter    int
	columns, rows  int
	texNum, lenTex int
	texRects       []image.Rectangle
	textures       []*ebiten.Image
}

func (s *Sprite) Scale() float64 {
	return s.scale
}

func (s *Sprite) VerticalAnchor() raycaster.SpriteAnchor {
	return s.anchor
}

func (s *Sprite) Texture() *ebiten.Image {
	return s.textures[s.texNum]
}

func (s *Sprite) TextureRect() image.Rectangle {
	return s.texRects[s.texNum]
}

func NewSprite(
	x, y, scale float64, img *ebiten.Image, mapColor color.RGBA, anchor raycaster.SpriteAnchor, collisionRadius float64,
) *Sprite {
	s := &Sprite{
		Entity: &Entity{
			Position:        &geom.Vector2{X: x, Y: y},
			PositionZ:       0.5,
			Angle:           0,
			Velocity:        0,
			CollisionRadius: collisionRadius,
			MapColor:        mapColor,
		},
	}
	s.scale = scale
	s.anchor = anchor

	s.texNum = 0
	s.lenTex = 1
	s.textures = make([]*ebiten.Image, s.lenTex)

	s.W, s.H = img.Size()
	s.texRects = []image.Rectangle{image.Rect(0, 0, s.W, s.H)}

	s.textures[0] = img

	return s
}

func NewSpriteFromSheet(
	x, y, scale float64, img *ebiten.Image, mapColor color.RGBA,
	columns, rows, spriteIndex int, anchor raycaster.SpriteAnchor, collisionRadius float64,
) *Sprite {
	s := &Sprite{
		Entity: &Entity{
			Position:        &geom.Vector2{X: x, Y: y},
			PositionZ:       0.5,
			Angle:           0,
			Velocity:        0,
			CollisionRadius: collisionRadius,
			MapColor:        mapColor,
		},
	}
	s.scale = scale
	s.anchor = anchor

	s.texNum = spriteIndex
	s.columns, s.rows = columns, rows
	s.lenTex = columns * rows
	s.textures = make([]*ebiten.Image, s.lenTex)
	s.texRects = make([]image.Rectangle, s.lenTex)

	w, h := img.Size()

	// crop sheet by given number of columns and rows into a single dimension array
	s.W = w / columns
	s.H = h / rows

	for r := 0; r < rows; r++ {
		y := r * s.H
		for c := 0; c < columns; c++ {
			x := c * s.W
			cellRect := image.Rect(x, y, x+s.W-1, y+s.H-1)
			cellImg := img.SubImage(cellRect).(*ebiten.Image)

			index := c + r*columns
			s.textures[index] = cellImg
			s.texRects[index] = cellRect
		}
	}

	return s
}

func NewAnimatedSprite(
	x, y, scale float64, animationRate int, img *ebiten.Image, mapColor color.RGBA,
	columns, rows int, anchor raycaster.SpriteAnchor, collisionRadius float64,
) *Sprite {
	s := &Sprite{
		Entity: &Entity{
			Position:        &geom.Vector2{X: x, Y: y},
			PositionZ:       0.5,
			Angle:           0,
			Velocity:        0,
			CollisionRadius: collisionRadius,
			MapColor:        mapColor,
		},
	}
	s.scale = scale
	s.anchor = anchor

	s.AnimationRate = animationRate
	s.animCounter = 0
	s.loopCounter = 0

	s.texNum = 0
	s.columns, s.rows = columns, rows
	s.lenTex = columns * rows
	s.textures = make([]*ebiten.Image, s.lenTex)
	s.texRects = make([]image.Rectangle, s.lenTex)

	w, h := img.Size()

	// crop sheet by given number of columns and rows into a single dimension array
	s.W = w / columns
	s.H = h / rows

	for r := 0; r < rows; r++ {
		y := r * s.H
		for c := 0; c < columns; c++ {
			x := c * s.W
			cellRect := image.Rect(x, y, x+s.W-1, y+s.H-1)
			cellImg := img.SubImage(cellRect).(*ebiten.Image)

			index := c + r*columns
			s.textures[index] = cellImg
			s.texRects[index] = cellRect
		}
	}

	return s
}

func (s *Sprite) SetAnimationReversed(isReverse bool) {
	s.animReversed = isReverse
}

func (s *Sprite) SetAnimationFrame(texNum int) {
	s.texNum = texNum
}

func (s *Sprite) ResetAnimation() {
	s.animCounter = 0
	s.loopCounter = 0
	s.texNum = 0
}

func (s *Sprite) GetLoopCounter() int {
	return s.loopCounter
}

func (s *Sprite) Update(camPos *geom.Vector2) {
	if s.AnimationRate <= 0 {
		return
	}

	if s.animCounter >= s.AnimationRate {
		minTexNum := 0
		maxTexNum := s.lenTex - 1

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

func (s *Sprite) AddDebugLines(lineWidth int, clr color.Color) {
	lW := float64(lineWidth)
	sW := float64(s.W)
	sH := float64(s.H)
	sCr := s.CollisionRadius * sW

	for i, img := range s.textures {
		imgRect := s.texRects[i]
		x, y := float64(imgRect.Min.X), float64(imgRect.Min.Y)

		// bounding box
		ebitenutil.DrawRect(img, x, y, lW, sH, clr)
		ebitenutil.DrawRect(img, x, y, sW, lW, clr)
		ebitenutil.DrawRect(img, x+sW-lW-1, y+sH-lW-1, lW, -sH, clr)
		ebitenutil.DrawRect(img, x+sW-lW-1, y+sH-lW-1, -sW, lW, clr)

		// center lines
		ebitenutil.DrawRect(img, x+sW/2-lW/2-1, y, lW, sH, clr)
		ebitenutil.DrawRect(img, x, y+sH/2-lW/2-1, sW, lW, clr)

		// collision markers
		if s.CollisionRadius > 0 {
			ebitenutil.DrawRect(img, x+sW/2-sCr-lW/2-1, y, lW, sH, color.White)
			ebitenutil.DrawRect(img, x+sW/2+sCr-lW/2-1, y, lW, sH, color.White)
		}
	}
}
