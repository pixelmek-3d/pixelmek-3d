package game

import (
	"github.com/hajimehoshi/ebiten/v2"
)

func (g *Game) drawCrosshairs(screen *ebiten.Image) {
	if g.crosshairs == nil {
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.ColorM.ScaleWithColor(g.hudRGBA)

	crosshairScale := g.crosshairs.Scale() * g.renderScale * g.hudScale
	op.GeoM.Scale(crosshairScale, crosshairScale)
	op.GeoM.Translate(
		float64(g.width)/2-float64(g.crosshairs.W)*crosshairScale/2,
		float64(g.height)/2-float64(g.crosshairs.H)*crosshairScale/2,
	)
	screen.DrawImage(g.crosshairs.Texture(), op)
}

func (g *Game) drawTargetReticle(screen *ebiten.Image) {
	if g.reticle == nil {
		return
	}

	r := g.reticle
	rScale := r.Scale() * g.renderScale * g.hudScale
	rOff := rScale * float64(r.W) / 2

	var op *ebiten.DrawImageOptions

	// setup some common draw modifications
	geoM := ebiten.GeoM{}
	geoM.Scale(rScale, rScale)

	colorM := ebiten.ColorM{}
	colorM.ScaleWithColor(g.hudRGBA)

	for s := range g.sprites.mechSprites {
		rect := s.ScreenRect()
		if rect == nil {
			continue
		}

		minX, minY, maxX, maxY := float64(rect.Min.X), float64(rect.Min.Y), float64(rect.Max.X), float64(rect.Max.Y)

		// top left corner
		g.reticle.SetTextureFrame(0)
		op = &ebiten.DrawImageOptions{ColorM: colorM, GeoM: geoM}
		op.Filter = ebiten.FilterNearest
		op.GeoM.Translate(minX-rOff, minY-rOff)
		screen.DrawImage(g.reticle.Texture(), op)

		// top right corner
		g.reticle.SetTextureFrame(1)
		op = &ebiten.DrawImageOptions{ColorM: colorM, GeoM: geoM}
		op.Filter = ebiten.FilterNearest
		op.GeoM.Translate(maxX-rOff, minY-rOff)
		screen.DrawImage(g.reticle.Texture(), op)

		// bottom left corner
		g.reticle.SetTextureFrame(2)
		op = &ebiten.DrawImageOptions{ColorM: colorM, GeoM: geoM}
		op.Filter = ebiten.FilterNearest
		op.GeoM.Translate(minX-rOff, maxY-rOff)
		screen.DrawImage(g.reticle.Texture(), op)

		// bottom right corner
		g.reticle.SetTextureFrame(3)
		op = &ebiten.DrawImageOptions{ColorM: colorM, GeoM: geoM}
		op.Filter = ebiten.FilterNearest
		op.GeoM.Translate(maxX-rOff, maxY-rOff)
		screen.DrawImage(g.reticle.Texture(), op)
	}
}
