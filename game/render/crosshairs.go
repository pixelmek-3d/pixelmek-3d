package render

import (
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/harbdog/raycaster-go/geom"
)

var (
	_colorCrosshair = _colorDefaultGreen
)

type Crosshairs struct {
	HUDSprite
	fovHorizontal float64
	fovVertical   float64
	angleOffset   float64
	pitchOffset   float64
}

func NewCrosshairs(
	img *ebiten.Image, scale float64, columns, rows, crosshairIndex int,
) *Crosshairs {
	c := &Crosshairs{
		HUDSprite: NewHUDSpriteFromSheet(img, scale, columns, rows, crosshairIndex),
	}

	return c
}

func (c *Crosshairs) SetFocalAngles(fovHorizontal, fovVertical float64) {
	c.fovHorizontal = fovHorizontal
	c.fovVertical = fovVertical
}

func (c *Crosshairs) SetOffsets(angleOffset, pitchOffset float64) {
	c.angleOffset = angleOffset
	c.pitchOffset = pitchOffset
}

func (c *Crosshairs) Draw(bounds image.Rectangle, hudOpts *DrawHudOptions) {
	screen := hudOpts.Screen
	sW, sH := screen.Bounds().Dx(), screen.Bounds().Dy()
	bX, bY, bW, bH := bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy()

	cScale := float64(bW) / float64(c.Width())

	cColor := hudOpts.HudColor(_colorCrosshair)
	cColorScale := ebiten.ColorScale{}
	cColorScale.ScaleWithColor(cColor)

	// render camera reticle at center screen as guide for where camera is currently looking
	// with dynamic reticle centered position based on angle/pitch offset
	var glX, glY float32 = float32(bX), float32(bY) + (float32(bH) / 2)
	var grX, grY float32 = float32(bX + bW), float32(bY) + (float32(bH) / 2)
	var gW, gT float32 = float32(bW), 3
	var gOffset float32 = 4 * gW * (float32(math.Abs(c.angleOffset)+math.Abs(c.pitchOffset)) / geom.Pi)
	vector.StrokeLine(screen, glX-gW-gOffset, glY, glX-gOffset, glY, gT, cColor, false)
	vector.StrokeLine(screen, grX+gW+gOffset, grY, grX+gOffset, grY, gT, cColor, false)

	// render crosshairs at an offset as unit/turret angle/pitch catches up to camera view
	var offX, offY float64
	if c.angleOffset != 0 {
		// calculate the length of the FOV triangle as base to find length of opposite leg using angle offset
		oppHorizontal := geom.GetOppositeTriangleBase(c.fovHorizontal/2, float64(sW)/2)
		offX = geom.GetOppositeTriangleLeg(c.angleOffset, oppHorizontal)
	}
	if c.pitchOffset != 0 {
		oppVertical := geom.GetOppositeTriangleBase(c.fovVertical/2, float64(sH)/2)
		offY = geom.GetOppositeTriangleLeg(c.pitchOffset, oppVertical)
	}

	op := &ebiten.DrawImageOptions{
		Filter:     ebiten.FilterNearest,
		ColorScale: cColorScale,
	}
	op.GeoM.Scale(cScale, cScale)
	op.GeoM.Translate(float64(bX)+offX, float64(bY)+offY)
	screen.DrawImage(c.Texture(), op)
}
