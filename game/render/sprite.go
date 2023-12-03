package render

import (
	"image"
	_ "image/png"
	"math"
	"sort"

	"github.com/harbdog/pixelmek-3d/game/model"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/jinzhu/copier"
)

type Sprite struct {
	model.Entity
	w, h              int
	scale             float64
	illumination      float64
	dimmingPerTick    float64
	animationRate     int
	focusable         bool
	animReversed      bool
	animCounter       int
	loopCounter       int
	maxLoops          int
	destroyCounter    int
	columns, rows     int
	texNum, lenTex    int
	staticTexNum      int
	camFacingOverride *facingAngleOverride
	texFacingMap      map[float64]int
	texFacingKeys     []float64
	texRects          []image.Rectangle
	textures          []*ebiten.Image
	screenRect        *image.Rectangle
}

type facingAngleOverride struct {
	angle float64
}

func (s *Sprite) Pos() *geom.Vector2 {
	return s.Entity.Pos()
}

func (s *Sprite) PosZ() float64 {
	return s.Entity.PosZ()
}

func (s *Sprite) Scale() float64 {
	return s.scale
}

func (s *Sprite) VerticalAnchor() raycaster.SpriteAnchor {
	return s.Entity.Anchor()
}

func (s *Sprite) StaticTexture() *ebiten.Image {
	return s.textures[s.staticTexNum]
}

func (s *Sprite) Texture() *ebiten.Image {
	return s.textures[s.texNum]
}

func (s *Sprite) TextureRect() image.Rectangle {
	return s.texRects[s.texNum]
}

func (s *Sprite) Illumination() float64 {
	return s.illumination
}

func (s *Sprite) SetScreenRect(rect *image.Rectangle) {
	s.screenRect = rect
}

func (s *Sprite) IsFocusable() bool {
	return s.focusable
}

func NewSprite(
	modelEntity model.Entity, scale float64, img *ebiten.Image,
) *Sprite {
	s := &Sprite{
		Entity:    modelEntity,
		focusable: true,
		scale:     scale,
	}

	s.w, s.h = img.Bounds().Dx(), img.Bounds().Dy()
	s.textures, s.texRects = GetSpriteSheetSlices(img, 1, 1)
	s.lenTex = 1

	return s
}

func NewSpriteFromSheet(
	modelEntity model.Entity, scale float64, img *ebiten.Image,
	columns, rows, spriteIndex int,
) *Sprite {
	s := &Sprite{
		Entity:    modelEntity,
		focusable: true,
		scale:     scale,
	}

	s.texNum = spriteIndex
	s.columns, s.rows = columns, rows

	// crop sheet by given number of columns and rows into a single dimension array
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	wFloat, hFloat := float64(w)/float64(columns), float64(h)/float64(rows)
	s.w, s.h = int(wFloat), int(hFloat)

	s.textures, s.texRects = GetSpriteSheetSlices(img, columns, rows)
	s.lenTex = len(s.textures)

	return s
}

func NewAnimatedSprite(
	modelEntity model.Entity, scale float64, img *ebiten.Image,
	columns, rows, animationRate int,
) *Sprite {
	s := &Sprite{
		Entity:    modelEntity,
		focusable: true,
		scale:     scale,
	}

	s.animationRate = animationRate
	s.animCounter = 0
	s.loopCounter = 0

	s.texNum = 0
	s.columns, s.rows = columns, rows

	// crop sheet by given number of columns and rows into a single dimension array
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
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

	w, h := img.Bounds().Dx(), img.Bounds().Dy()

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

func (s *Sprite) SetIlluminationPeriod(illumination float64, periodSeconds float64) {
	s.illumination = illumination

	// determine the amount of illumination to dim by per tick
	s.dimmingPerTick = illumination * model.SECONDS_PER_TICK / periodSeconds
}

func (s *Sprite) updateIllumination() {
	if s.dimmingPerTick > 0 && s.illumination > 0 {
		s.illumination -= s.dimmingPerTick
		if s.illumination <= 0 {
			s.illumination = 0
			s.dimmingPerTick = 0
		}
	}
}

func (s *Sprite) SetTextureFacingMap(texFacingMap map[float64]int) {
	s.texFacingMap = texFacingMap

	// create pre-sorted list of keys used during facing determination
	s.texFacingKeys = make([]float64, 0, len(texFacingMap))
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

func (s *Sprite) AnimationReversed() bool {
	return s.animReversed
}

func (s *Sprite) SetAnimationReversed(isReverse bool) {
	s.animReversed = isReverse
}

func (s *Sprite) SetTextureFrame(texNum int) {
	s.texNum = texNum
}

func (s *Sprite) AnimationFrameCounter() int {
	return s.animCounter
}

func (s *Sprite) NumAnimationFrames() int {
	return s.lenTex
}

func (e *EffectSprite) AnimationDuration() int {
	return (e.animationRate + 1) * e.NumAnimationFrames()
}

func (s *Sprite) ResetAnimation() {
	s.animCounter = 0
	s.loopCounter = 0
	s.texNum = 0
}

func (s *Sprite) LoopCounter() int {
	return s.loopCounter
}

// DestroyCounter used only when destroyed to remain visible until ready to be removed
func (s *Sprite) DestroyCounter() int {
	return s.destroyCounter
}

// SetDestroyCounter to set how many ticks to remain visible after destroyed before removal
func (s *Sprite) SetDestroyCounter(counter int) int {
	s.destroyCounter = counter
	return s.destroyCounter
}

func (s *Sprite) ScreenRect(renderScale float64) *image.Rectangle {
	if s.screenRect == nil {
		return nil
	}
	if renderScale == 1 {
		return s.screenRect
	}

	// convert scene position to screen position
	x0, y0 := float64(s.screenRect.Min.X)*1/renderScale, float64(s.screenRect.Min.Y)*1/renderScale
	x1, y1 := float64(s.screenRect.Max.X)*1/renderScale, float64(s.screenRect.Max.Y)*1/renderScale
	sRect := image.Rect(int(x0), int(y0), int(x1), int(y1))
	return &sRect
}

func (s *Sprite) Update(camPos *geom.Vector2) {
	s.updateIllumination()

	if s.animationRate <= 0 {
		return
	}

	if s.animCounter >= s.animationRate {
		minTexNum := 0
		maxTexNum := s.lenTex - 1

		if len(s.texFacingMap) > 1 && camPos != nil {
			// use facing from camera position to determine min/max texNum in texFacingMap
			// to update facing of sprite relative to camera and sprite angle
			texRow := 0

			// calculate angle from sprite relative to camera position by getting angle of line between them
			var facingAngle float64
			if s.camFacingOverride != nil {
				facingAngle = s.camFacingOverride.angle
			} else {
				lineToCam := geom.Line{X1: s.Pos().X, Y1: s.Pos().Y, X2: camPos.X, Y2: camPos.Y}
				facingAngle = lineToCam.Angle() - s.Heading()
			}

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
