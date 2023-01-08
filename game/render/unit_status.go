package render

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/tinne26/etxt"
	"github.com/tinne26/etxt/efixed"
)

type UnitStatus struct {
	HUDSprite
	fontRenderer *etxt.Renderer
	unit         *Sprite
	unitDistance float64
	isPlayer     bool
}

//NewUnitStatus creates a unit status element image to be rendered on demand
func NewUnitStatus(isPlayer bool, font *Font) *UnitStatus {
	// create and configure font renderer
	renderer := etxt.NewStdRenderer()
	renderer.SetCacheHandler(font.FontCache.NewHandler())
	renderer.SetFont(font.Font)
	renderer.SetColor(color.RGBA{255, 255, 255, 255})

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

func (u *UnitStatus) updateFontSize(width, height int) {
	// set font size based on element size
	pxSize := float64(height) / 8
	if pxSize < 1 {
		pxSize = 1
	}

	fractSize, _ := efixed.FromFloat64(pxSize)
	u.fontRenderer.SetSizePxFract(fractSize)
}

func (u *UnitStatus) Draw(screen *ebiten.Image, bounds image.Rectangle, clr *color.RGBA) {
	u.fontRenderer.SetTarget(screen)
	u.fontRenderer.SetAlign(etxt.YCenter, etxt.Left)
	u.fontRenderer.SetColor(clr)

	bX, bY, bW, bH := bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy()
	u.updateFontSize(bW, bH)

	if u.unit == nil {
		// TESTING!
		sW, sH := float64(bW), float64(bH)
		sX, sY := float64(bX), float64(bY)
		sAlpha := uint8(int(clr.A) / 10)
		ebitenutil.DrawRect(screen, sX, sY, sW, sH, color.RGBA{clr.R, clr.G, clr.B, sAlpha})
		return
	}

	// background box
	sW, sH := float64(bW), float64(bH)
	sX, sY := float64(bX), float64(bY)
	sAlpha := uint8(int(clr.A) / 5)
	ebitenutil.DrawRect(screen, sX, sY, sW, sH, color.RGBA{clr.R, clr.G, clr.B, sAlpha})

	// unit image
	// create static outline image of unit and store it
	uTexture := u.unit.StaticTexture()

	op := &ebiten.DrawImageOptions{}
	// Reset RGB (not Alpha) 0 forcibly
	op.ColorM.Scale(0, 0, 0, 1)

	// Set color
	r, g, b := float64(clr.R)/255, float64(clr.G)/255, float64(clr.B)/255
	op.ColorM.Translate(r, g, b, 0)

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
	screen.DrawImage(uTexture, op)

	// armor readout
	armorPercent := 100 * u.unit.ArmorPoints() / u.unit.MaxArmorPoints()
	armorStr := fmt.Sprintf("ARMOR\n %0.0f%%", armorPercent)
	u.fontRenderer.Draw(armorStr, int(sX)+int(3*sW/5), int(sY)+int(sH/3))

	// internal structure readout
	internalPercent := 100 * u.unit.StructurePoints() / u.unit.MaxStructurePoints()
	internalStr := fmt.Sprintf("STRUCT\n %0.0f%%", internalPercent)
	u.fontRenderer.Draw(internalStr, int(sX)+int(3*sW/5), int(sY)+int(2*sH/3))

	if !u.isPlayer {
		// target chassis name
		tUnit := model.EntityUnit(u.unit.Entity)
		if tUnit != nil {
			u.fontRenderer.SetAlign(etxt.Top, etxt.XCenter)
			chassisVariant := strings.ToUpper(tUnit.Variant())
			u.fontRenderer.Draw(chassisVariant, bX+bW/2, bY)
		}

		// target distance
		if u.unitDistance >= 0 {
			u.fontRenderer.SetAlign(etxt.Bottom, etxt.XCenter)
			distanceStr := fmt.Sprintf("%0.0fm", u.unitDistance)
			u.fontRenderer.Draw(distanceStr, bX+bW/2, bY+bH)
		}
	}
}
