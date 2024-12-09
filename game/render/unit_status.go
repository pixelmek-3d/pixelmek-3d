package render

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2/colorm"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/tinne26/etxt"
)

var (
	_colorStatusOk         = color.NRGBA{R: 0, G: 155, B: 255, A: 255}
	_colorStatusWarn       = _colorDefaultYellow
	_colorStatusCritical   = color.NRGBA{R: 255, G: 30, B: 30, A: 255}
	_colorStatusBackground = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
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
	isSpectating   bool
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

func (u *UnitStatus) SetIsPlayer(isPlayer bool) {
	u.isPlayer = isPlayer
}

func (u *UnitStatus) SetIsSpectating(isSpectating bool) {
	u.isSpectating = isSpectating
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

func (u *UnitStatus) updateFontSize(_, height int) {
	// set font size based on element size
	pxSize := float64(height) / 8
	if pxSize < 1 {
		pxSize = 1
	}

	u.fontRenderer.SetSizePx(int(pxSize))
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
		bColor := hudOpts.HudColor(_colorStatusBackground)

		sAlpha := uint8(int(bColor.A) / 3)
		vector.DrawFilledRect(screen, float32(sX), float32(sY), float32(sW), float32(sH), color.NRGBA{bColor.R, bColor.G, bColor.B, sAlpha}, false)
	}

	op := &colorm.DrawImageOptions{}
	var uColor color.NRGBA
	// Set unit image color based on health status
	switch {
	case internalPercent <= 50:
		uColor = hudOpts.HudColor(_colorStatusCritical)
	case armorPercent <= 25 || internalPercent <= 75:
		uColor = hudOpts.HudColor(_colorStatusWarn)
	default:
		uColor = hudOpts.HudColor(_colorStatusOk)
	}

	// create static outline image of unit
	uTexture := u.unit.StaticTexture()

	// Reset RGB (not Alpha) 0 forcibly
	var cm colorm.ColorM
	cm.Scale(0, 0, 0, 1)
	cm.Translate(1, 1, 1, 0)
	cm.ScaleWithColor(uColor)

	iH := bounds.Dy()
	uW, uH := uTexture.Bounds().Dx(), uTexture.Bounds().Dy()

	uSize := uH
	if uW > uH {
		uSize = uW
	}

	var uScale float64
	if u.isPlayer {
		uScale = (0.9 * float64(iH)) / float64(uSize)
	} else {
		uScale = (0.6 * float64(iH)) / float64(uSize)
	}

	op.GeoM.Scale(uScale, uScale)
	op.GeoM.Translate(sX, sY+sH/2-uScale*float64(uH)/2)
	colorm.DrawImage(screen, uTexture, cm, op)

	// setup text color
	tColor := hudOpts.HudColor(_colorStatusText)
	u.fontRenderer.SetColor(tColor)

	// armor readout
	armorStr := fmt.Sprintf("ARMOR\n %0.0f%%", armorPercent)
	u.fontRenderer.Draw(armorStr, int(sX)+int(3*sW/5), int(sY)+int(sH/3))

	// internal structure readout
	internalStr := fmt.Sprintf("STRUCT\n %0.0f%%", internalPercent)
	u.fontRenderer.Draw(internalStr, int(sX)+int(3*sW/5), int(sY)+int(2*sH/3))

	if u.isSpectating {
		sUnit := model.EntityUnit(u.unit.Entity)
		if sUnit != nil {
			// spectated chassis name
			var eColor color.NRGBA
			if sUnit.Team() < 0 {
				eColor = hudOpts.HudColor(_colorFriendly)
			} else {
				eColor = hudOpts.HudColor(_colorEnemy)
			}
			u.fontRenderer.SetColor(eColor)

			chassisVariant := strings.ToUpper(sUnit.Variant())
			u.fontRenderer.SetAlign(etxt.Top, etxt.XCenter)
			u.fontRenderer.Draw(chassisVariant, bX+bW/2, bY)

			// show spectating text
			sColor := hudOpts.HudColor(_colorStatusWarn)
			u.fontRenderer.SetColor(sColor)
			u.fontRenderer.SetAlign(etxt.Bottom, etxt.XCenter)
			u.fontRenderer.Draw("SPECTATING", bX+bW/2, bY-2)
		}
	} else if !u.isPlayer {
		// target distance
		if u.unitDistance >= 0 {
			u.fontRenderer.SetAlign(etxt.Bottom, etxt.XCenter)
			distanceStr := fmt.Sprintf("%0.0fm", u.unitDistance)
			u.fontRenderer.Draw(distanceStr, bX+bW/2, bY+bH)
		}

		tUnit := model.EntityUnit(u.unit.Entity)
		if tUnit != nil {
			// target chassis name
			var eColor color.NRGBA
			if tUnit.Team() < 0 {
				eColor = hudOpts.HudColor(_colorFriendly)
			} else {
				eColor = hudOpts.HudColor(_colorEnemy)
			}
			u.fontRenderer.SetColor(eColor)

			chassisVariant := strings.ToUpper(tUnit.Variant())
			switch tUnit.Objective() {
			case model.DestroyUnitObjective:
				chassisVariant = "*" + chassisVariant + "*"
			case model.ProtectUnitObjective:
				chassisVariant = "^" + chassisVariant + "^"
			}

			u.fontRenderer.SetAlign(etxt.Top, etxt.XCenter)
			u.fontRenderer.Draw(chassisVariant, bX+bW/2, bY)

			// if lock-ons equipped, display lock percent on target
			if u.showTargetLock {
				lColor := hudOpts.HudColor(eColor)
				if u.targetLock < 1.0 {
					lColor = hudOpts.HudColor(_colorStatusWarn)
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
		if u.unit.Team() < 0 {
			// TODO: friendly reticle needs to look different in case custom HUD color is used
			u.targetReticle.Friendly = true
		} else {
			u.targetReticle.Friendly = false
		}
		u.targetReticle.Draw(bounds, hudOpts)
	}
}
