package render

import (
	"image"
	"image/color"
	_ "image/png"
	"math"
	"sort"

	"github.com/harbdog/pixelmek-3d/game/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"
)

type Sprite struct {
	model.Entity
	w, h           int
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
	MapColor       color.RGBA
}

func (s *Sprite) Pos() *geom.Vector2 {
	return s.Entity.Pos()
}

func (s *Sprite) PosZ() float64 {
	return s.Entity.PosZ()
}

func (s *Sprite) Scale() float64 {
	return s.Entity.Scale()
}

func (s *Sprite) VerticalAnchor() raycaster.SpriteAnchor {
	return s.Entity.Anchor()
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
	x, y, scale float64, img *ebiten.Image, mapColor color.RGBA, anchor raycaster.SpriteAnchor, collisionRadius, collisionHeight float64,
) *Sprite {
	s := &Sprite{
		Entity:   &model.BasicEntity{},
		MapColor: mapColor,
	}

	s.SetPos(&geom.Vector2{X: x, Y: y})
	s.SetPosZ(0)
	s.SetScale(scale)
	s.SetAnchor(anchor)
	s.SetAngle(0)
	s.SetVelocity(0)
	s.SetCollisionRadius(collisionRadius)
	s.SetCollisionHeight(collisionHeight)
	s.SetHitPoints(math.MaxFloat64)

	s.w, s.h = img.Size()
	s.textures, s.texRects = GetSpriteSheetSlices(img, 1, 1)
	s.lenTex = 1

	return s
}

func NewSpriteFromSheet(
	x, y, scale float64, img *ebiten.Image, mapColor color.RGBA,
	columns, rows, spriteIndex int, anchor raycaster.SpriteAnchor, collisionRadius, collisionHeight float64,
) *Sprite {
	s := &Sprite{
		Entity:   &model.BasicEntity{},
		MapColor: mapColor,
	}

	s.SetPos(&geom.Vector2{X: x, Y: y})
	s.SetPosZ(0)
	s.SetScale(scale)
	s.SetAnchor(anchor)
	s.SetAngle(0)
	s.SetVelocity(0)
	s.SetCollisionRadius(collisionRadius)
	s.SetCollisionHeight(collisionHeight)
	s.SetHitPoints(math.MaxFloat64)

	s.texNum = spriteIndex
	s.columns, s.rows = columns, rows

	// crop sheet by given number of columns and rows into a single dimension array
	w, h := img.Size()
	wFloat, hFloat := float64(w)/float64(columns), float64(h)/float64(rows)
	s.w, s.h = int(wFloat), int(hFloat)

	s.textures, s.texRects = GetSpriteSheetSlices(img, columns, rows)
	s.lenTex = len(s.textures)

	return s
}

func NewAnimatedSprite(
	x, y, scale float64, img *ebiten.Image, mapColor color.RGBA,
	columns, rows, animationRate int, anchor raycaster.SpriteAnchor, collisionRadius, collisionHeight float64,
) *Sprite {
	s := &Sprite{
		Entity:   &model.BasicEntity{},
		MapColor: mapColor,
	}

	s.SetPos(&geom.Vector2{X: x, Y: y})
	s.SetPosZ(0)
	s.SetScale(scale)
	s.SetAnchor(anchor)
	s.SetAngle(0)
	s.SetVelocity(0)
	s.SetCollisionRadius(collisionRadius)
	s.SetCollisionHeight(collisionHeight)
	s.SetHitPoints(math.MaxFloat64)

	s.AnimationRate = animationRate
	s.animCounter = 0
	s.loopCounter = 0

	s.texNum = 0
	s.columns, s.rows = columns, rows

	// crop sheet by given number of columns and rows into a single dimension array
	w, h := img.Size()
	wFloat, hFloat := float64(w)/float64(columns), float64(h)/float64(rows)
	s.w, s.h = int(wFloat), int(hFloat)

	s.textures, s.texRects = GetSpriteSheetSlices(img, columns, rows)
	s.lenTex = len(s.textures)

	return s
}

func GetSpriteSheetSlices(img *ebiten.Image, columns, rows int) ([]*ebiten.Image, []image.Rectangle) {
	lenTex := columns * rows
	textures := make([]*ebiten.Image, lenTex)
	texRects := make([]image.Rectangle, lenTex)

	w, h := img.Size()

	// crop sheet by given number of columns and rows into a single dimension array
	wFloat, hFloat := float64(w)/float64(columns), float64(h)/float64(rows)
	spriteW, spriteH := int(wFloat), int(hFloat)

	for r := 0; r < rows; r++ {
		y := int(float64(r) * hFloat)
		for c := 0; c < columns; c++ {
			x := int(float64(c) * wFloat)
			cellRect := image.Rect(x, y, x+spriteW-1, y+spriteH-1)
			cellImg := img.SubImage(cellRect).(*ebiten.Image)

			index := c + r*columns
			textures[index] = cellImg
			texRects[index] = cellRect
		}
	}
	return textures, texRects
}

func (s *Sprite) Clone() *Sprite {
	sClone := &Sprite{}
	eClone := &model.BasicEntity{}

	copier.Copy(sClone, s)
	copier.Copy(eClone, s.Entity)

	sClone.Entity = eClone

	return sClone
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

func (s *Sprite) SetTextureFrame(texNum int) {
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
			lineToCam := geom.Line{X1: s.Pos().X, Y1: s.Pos().Y, X2: camPos.X, Y2: camPos.Y}
			facingAngle := lineToCam.Angle() - s.Angle()
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
	sW, sH := float64(s.w), float64(s.h)
	cR := s.CollisionRadius()
	sCr := cR * sW

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
		if cR > 0 {
			ebitenutil.DrawRect(img, x+sW/2-sCr-lW/2-1, y, lW, sH, color.White)
			ebitenutil.DrawRect(img, x+sW/2+sCr-lW/2-1, y, lW, sH, color.White)
		}
	}
}
