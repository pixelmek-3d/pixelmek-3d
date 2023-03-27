package render

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2/colorm"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/tinne26/etxt"
	"github.com/tinne26/etxt/efixed"
)

var (
	_colorStatusOk         = color.RGBA{R: 0, G: 155, B: 255, A: 255}
	_colorStatusWarn       = _colorDefaultYellow
	_colorStatusCritical   = color.RGBA{R: 255, G: 30, B: 30, A: 255}
	_colorStatusBackground = color.RGBA{R: 0, G: 0, B: 0, A: 255}
	_colorStatusText       = _colorDefaultGreen
)

type UnitStatus struct {
	HUDSprite
	fontRenderer   *etxt.Renderer
	unit           *Sprite
	unitDistance   float64
	showTargetLock bool
	targetLock     float64
	isPlayer       bool
	targetReticle  *TargetReticle
}

// NewUnitStatus creates a unit status element image to be rendered on demand
func NewUnitStatus(isPlayer bool, font *Font) *UnitStatus {
	// create and configure font renderer
	renderer := etxt.NewStdRenderer()
	renderer.SetCacheHandler(font.FontCache.NewHandler())
	renderer.SetFont(font.Font)

	u := &UnitStatus{
		HUDSprite:    NewHUDSprite(nil, 1.0),
		fontRenderer: renderer,
		isPlayer:     isPlayer,
		unitDistance: -1,
	}

	return u
}

func (u *UnitStatus) Unit() *Sprite {
	return u.unit
}

func (u *UnitStatus) SetUnit(unit *Sprite) {
	u.unit = unit
}

func (u *UnitStatus) SetUnitDistance(distance float64) {
	u.unitDistance = distance
}

func (u *UnitStatus) ShowTargetLock(show bool) {
	u.showTargetLock = show
}

func (u *UnitStatus) SetTargetLock(lockPercent float64) {
	u.targetLock = lockPercent
}

func (u *UnitStatus) SetTargetReticle(reticle *TargetReticle) {
	u.targetReticle = reticle
}

func (u *UnitStatus) updateFontSize(width, height int) {
	// set font size based on element size
	pxSize := float64(height) / 8
	if pxSize < 1 {
		pxSize = 1
	}

	fractSize, _ := efixed.FromFloat64(pxSize)
	u.fontRenderer.SetSizePxFract(fractSize)
}

func (u *UnitStatus) Draw(bounds image.Rectangle, hudOpts *DrawHudOptions) {
	screen := hudOpts.Screen
	u.fontRenderer.SetTarget(screen)
	u.fontRenderer.SetAlign(etxt.YCenter, etxt.Left)

	bX, bY, bW, bH := bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy()
	u.updateFontSize(bW, bH)

	if u.unit == nil {
		return
	}

	// determine unit status
	armorPercent := 100 * u.unit.ArmorPoints() / u.unit.MaxArmorPoints()
	internalPercent := 100 * u.unit.StructurePoints() / u.unit.MaxStructurePoints()

	sW, sH := float64(bW), float64(bH)
	sX, sY := float64(bX), float64(bY)

	if !u.isPlayer {
		// background box
		bColor := _colorStatusBackground
		if hudOpts.UseCustomColor {
			bColor = hudOpts.Color
		}

		sAlpha := uint8(int(bColor.A) / 3)
		vector.DrawFilledRect(screen, float32(sX), float32(sY), float32(sW), float32(sH), color.RGBA{bColor.R, bColor.G, bColor.B, sAlpha}, false)
	}

	// create static outline image of unit
	uTexture := u.unit.StaticTexture()

	op := &colorm.DrawImageOptions{}
	// Reset RGB (not Alpha) 0 forcibly
	var cm colorm.ColorM
	cm.Scale(0, 0, 0, 1)

	// Set unit image color based on health status
	var uColor color.RGBA
	if hudOpts.UseCustomColor {
		uColor = hudOpts.Color
	} else {
		if armorPercent >= 25 {
			uColor = _colorStatusOk
		} else if internalPercent >= 50 {
			uColor = _colorStatusWarn
		} else {
			uColor = _colorStatusCritical
		}
	}
	r, g, b := float64(uColor.R)/255, float64(uColor.G)/255, float64(uColor.B)/255
	cm.Translate(r, g, b, 0)

	iH := bounds.Dy()
	uH := uTexture.Bounds().Dy()

	var uScale float64
	if u.isPlayer {
		uScale = (0.9 * float64(iH)) / float64(uH)
	} else {
		uScale = (0.6 * float64(iH)) / float64(uH)
	}

	op.GeoM.Scale(uScale, uScale)
	op.GeoM.Translate(sX, sY+sH/2-uScale*float64(uH)/2)
	colorm.DrawImage(screen, uTexture, cm, op)

	// setup text color
	tColor := _colorStatusText
	if hudOpts.UseCustomColor {
		tColor = hudOpts.Color
	}
	u.fontRenderer.SetColor(tColor)

	// armor readout
	armorStr := fmt.Sprintf("ARMOR\n %0.0f%%", armorPercent)
	u.fontRenderer.Draw(armorStr, int(sX)+int(3*sW/5), int(sY)+int(sH/3))

	// internal structure readout
	internalStr := fmt.Sprintf("STRUCT\n %0.0f%%", internalPercent)
	u.fontRenderer.Draw(internalStr, int(sX)+int(3*sW/5), int(sY)+int(2*sH/3))

	if !u.isPlayer {
		// target distance
		if u.unitDistance >= 0 {
			u.fontRenderer.SetAlign(etxt.Bottom, etxt.XCenter)
			distanceStr := fmt.Sprintf("%0.0fm", u.unitDistance)
			u.fontRenderer.Draw(distanceStr, bX+bW/2, bY+bH)
		}

		tUnit := model.EntityUnit(u.unit.Entity)
		if tUnit != nil {
			// target chassis name
			eColor := _colorEnemy
			if hudOpts.UseCustomColor {
				eColor = hudOpts.Color
			}
			u.fontRenderer.SetColor(eColor)

			u.fontRenderer.SetAlign(etxt.Top, etxt.XCenter)
			chassisVariant := strings.ToUpper(tUnit.Variant())
			u.fontRenderer.Draw(chassisVariant, bX+bW/2, bY)

			// if lock-ons equipped, display lock percent on target
			if u.showTargetLock {
				lColor := eColor
				if u.targetLock < 1.0 {
					lColor = _colorStatusWarn
					if hudOpts.UseCustomColor {
						lColor = hudOpts.Color
					}
				}
				u.fontRenderer.SetColor(lColor)
				u.fontRenderer.SetAlign(etxt.Bottom, etxt.Left)

				lockStr := fmt.Sprintf("LOCK: %0.0f%%", u.targetLock*100)
				u.fontRenderer.Draw(lockStr, bX, bY-u.targetReticle.Height())
			}
		}

	}

	if u.targetReticle != nil {
		// render target reticles on outer corners of status display
		u.targetReticle.Draw(bounds, hudOpts)
	}
}
