package render

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/harbdog/raycaster-go/geom"
)

var (
	_colorCrosshair = _colorDefaultGreen
)

type Crosshairs struct {
	HUDSprite
	fovDepth    float64
	angleOffset float64
	pitchOffset float64
}

func NewCrosshairs(
	img *ebiten.Image, scale float64, columns, rows, crosshairIndex int,
) *Crosshairs {
	c := &Crosshairs{
		HUDSprite: NewHUDSpriteFromSheet(img, scale, columns, rows, crosshairIndex),
	}

	return c
}

func (c *Crosshairs) SetFocalDepth(fovDepth float64) {
	c.fovDepth = fovDepth
}

func (c *Crosshairs) SetOffsets(angleOffset, pitchOffset float64) {
	c.angleOffset = angleOffset
	c.pitchOffset = pitchOffset
}

func (c *Crosshairs) Draw(bounds image.Rectangle, hudOpts *DrawHudOptions) {
	screen := hudOpts.Screen
	sW, sH := screen.Bounds().Dx(), screen.Bounds().Dy()
	bX, bY, bW := bounds.Min.X, bounds.Min.Y, bounds.Dx()

	cScale := float64(bW) / float64(c.Width())

	cColor := hudOpts.HudColor(_colorCrosshair)
	cColorScale := ebiten.ColorScale{}
	cColorScale.ScaleWithColor(cColor)

	// render camera dot at center screen as guide for where camera is currently looking
	midX, midY := float64(sW)/2, float64(sH)/2
	var cT, cR float32 = 1, 4 // TODO: calculate line thickness and radius based on crosshair size
	vector.StrokeCircle(screen, float32(midX), float32(midY), cR, cT, cColor, false)

	// render crosshairs at an offset as unit/turret angle/pitch catches up to camera view
	var offX, offY float64
	if c.angleOffset != 0 {
		offX = geom.GetOppositeTriangleLeg(c.angleOffset, c.fovDepth) * float64(sW)
	}
	if c.pitchOffset != 0 {
		offY = geom.GetOppositeTriangleLeg(c.pitchOffset, c.fovDepth) * float64(sH)
	}

	op := &ebiten.DrawImageOptions{
		Filter:     ebiten.FilterNearest,
		ColorScale: cColorScale,
	}
	op.GeoM.Scale(cScale, cScale)
	op.GeoM.Translate(float64(bX)+offX, float64(bY)+offY)
	screen.DrawImage(c.Texture(), op)
}
