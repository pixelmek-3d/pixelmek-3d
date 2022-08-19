package model

import (
	"image"
	"image/color"
	_ "image/png"
	"math"
	"sort"

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
	texFacingMap   map[float64]int
	texFacingKeys  []float64
	texRects       []image.Rectangle
	textures       []*ebiten.Image
	screenRect     *image.Rectangle
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

func (s *Sprite) SetScreenRect(rect *image.Rectangle) {
	s.screenRect = rect
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
			HitPoints:       math.MaxFloat64,
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
			HitPoints:       math.MaxFloat64,
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
	wFloat, hFloat := float64(w)/float64(columns), float64(h)/float64(rows)
	s.W = int(wFloat)
	s.H = int(hFloat)

	for r := 0; r < rows; r++ {
		y := int(float64(r) * hFloat)
		for c := 0; c < columns; c++ {
			x := int(float64(c) * wFloat)
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
	x, y, scale float64, img *ebiten.Image, mapColor color.RGBA,
	columns, rows, animationRate int, anchor raycaster.SpriteAnchor, collisionRadius float64,
) *Sprite {
	s := &Sprite{
		Entity: &Entity{
			Position:        &geom.Vector2{X: x, Y: y},
			PositionZ:       0.5,
			Angle:           0,
			Velocity:        0,
			CollisionRadius: collisionRadius,
			HitPoints:       math.MaxFloat64,
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
			cellRect := image.Rect(x, y, x+s.W, y+s.H)
			cellImg := img.SubImage(cellRect).(*ebiten.Image)

			index := c + r*columns
			s.textures[index] = cellImg
			s.texRects[index] = cellRect
		}
	}

	return s
}

func (s *Sprite) SetTextureFacingMap(texFacingMap map[float64]int) {
	s.texFacingMap = texFacingMap

	// create pre-sorted list of keys used during facing determination
	s.texFacingKeys = make([]float64, len(texFacingMap))
	for k := range texFacingMap {
		s.texFacingKeys = append(s.texFacingKeys, k)
	}
	sort.Float64s(s.texFacingKeys)
}

func (s *Sprite) getTextureFacingKeyForAngle(facingAngle float64) float64 {
	var closestKeyAngle float64 = -1
	if s.texFacingMap == nil || len(s.texFacingMap) == 0 || s.texFacingKeys == nil || len(s.texFacingKeys) == 0 {
		return closestKeyAngle
	}

	closestKeyDiff := math.MaxFloat64
	for _, keyAngle := range s.texFacingKeys {
		keyDiff := math.Min(geom.Pi2-math.Abs(float64(keyAngle)-facingAngle), math.Abs(float64(keyAngle)-facingAngle))
		if keyDiff < closestKeyDiff {
			closestKeyDiff = keyDiff
			closestKeyAngle = keyAngle
		}
	}

	return closestKeyAngle
}

func (s *Sprite) SetAnimationReversed(isReverse bool) {
	s.animReversed = isReverse
}

func (s *Sprite) SetAnimationFrame(texNum int) {
	s.texNum = texNum
}

func (s *Sprite) NumAnimationFrames() int {
	return s.lenTex
}

func (s *Sprite) ResetAnimation() {
	s.animCounter = 0
	s.loopCounter = 0
	s.texNum = 0
}

func (s *Sprite) LoopCounter() int {
	return s.loopCounter
}

func (s *Sprite) ScreenRect() *image.Rectangle {
	return s.screenRect
}

func (s *Sprite) Update(camPos *geom.Vector2) {
	if s.AnimationRate <= 0 {
		return
	}

	if s.animCounter >= s.AnimationRate {
		minTexNum := 0
		maxTexNum := s.lenTex - 1

		if len(s.texFacingMap) > 1 && camPos != nil {
			// use facing from camera position to determine min/max texNum in texFacingMap
			// to update facing of sprite relative to camera and sprite angle
			texRow := 0

			// calculate angle from sprite relative to camera position by getting angle of line between them
			lineToCam := geom.Line{X1: s.Position.X, Y1: s.Position.Y, X2: camPos.X, Y2: camPos.Y}
			facingAngle := lineToCam.Angle() - s.Angle
			if facingAngle < 0 {
				// convert to positive angle needed to determine facing index to use
				facingAngle += geom.Pi2
			}
			facingKeyAngle := s.getTextureFacingKeyForAngle(facingAngle)
			if texFacingValue, ok := s.texFacingMap[facingKeyAngle]; ok {
				texRow = texFacingValue
			}

			minTexNum = texRow * s.columns
			maxTexNum = texRow*s.columns + s.columns - 1
		}

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
