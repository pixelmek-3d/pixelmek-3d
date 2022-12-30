package render

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/tinne26/etxt"
)

type UnitStatus struct {
	HUDSprite
	fontRenderer *etxt.Renderer
	unit         *Sprite
}

//NewUnitStatus creates a unit status element image to be rendered on demand
func NewUnitStatus(font *Font) *UnitStatus {
	// create and configure font renderer
	renderer := etxt.NewStdRenderer()
	renderer.SetCacheHandler(font.FontCache.NewHandler())
	renderer.SetFont(font.Font)
	renderer.SetAlign(etxt.YCenter, etxt.Left)
	renderer.SetColor(color.RGBA{255, 255, 255, 255})

	u := &UnitStatus{
		HUDSprite:    NewHUDSprite(nil, 1.0),
		fontRenderer: renderer,
	}

	return u
}

func (u *UnitStatus) SetUnit(unit *Sprite) {
	u.unit = unit
}

func (u *UnitStatus) Draw(screen *ebiten.Image, bounds image.Rectangle, clr *color.RGBA) {
	u.fontRenderer.SetTarget(screen)
	u.fontRenderer.SetColor(clr)
	u.fontRenderer.SetSizePx(int(16.0 * u.Scale()))

	bX, bY, bW, bH := bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy()

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
	uTexture := u.unit.Texture() // FIXME: unit texture needs to always be the static, front facing image

	op := &ebiten.DrawImageOptions{}
	// Reset RGB (not Alpha) 0 forcibly
	op.ColorM.Scale(0, 0, 0, 1)

	// Set color
	r, g, b := float64(clr.R)/255, float64(clr.G)/255, float64(clr.B)/255
	op.ColorM.Translate(r, g, b, 0)

	iH := bounds.Dy()
	uH := uTexture.Bounds().Dy()
	uScale := (0.9 * float64(iH)) / float64(uH)
	op.GeoM.Scale(uScale, uScale)
	op.GeoM.Translate(sX, sY)
	screen.DrawImage(uTexture, op)

	// armor readout
	armorPercent := 100 * u.unit.ArmorPoints() / u.unit.MaxArmorPoints()
	armorStr := fmt.Sprintf("ARMOR\n%0.0f%%", armorPercent)
	u.fontRenderer.Draw(armorStr, int(sX)+int(2*sW/3), int(sY)+int(sH/3))

	// internal structure readout
	internalPercent := 100 * u.unit.StructurePoints() / u.unit.MaxStructurePoints()
	internalStr := fmt.Sprintf("STRUCT\n%0.0f%%", internalPercent)
	u.fontRenderer.Draw(internalStr, int(sX)+int(2*sW/3), int(sY)+int(2*sH/3))
}
