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

	crosshairScale := g.crosshairs.Scale()
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
	rScale := r.Scale()

	var op *ebiten.DrawImageOptions

	for s := range g.sprites.mechSprites {
		rect := s.ScreenRect()
		if rect == nil {
			continue
		}

		// top left corner
		g.reticle.SetTextureFrame(0)
		op = &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest
		op.GeoM.Scale(rScale, rScale)
		op.GeoM.Translate(float64(rect.Min.X-r.W/2), float64(rect.Min.Y-r.W/2))
		screen.DrawImage(g.reticle.Texture(), op)

		// top right corner
		g.reticle.SetTextureFrame(1)
		op = &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest
		op.GeoM.Scale(rScale, rScale)
		op.GeoM.Translate(float64(rect.Max.X-r.W/2), float64(rect.Min.Y-r.W/2))
		screen.DrawImage(g.reticle.Texture(), op)

		// bottom left corner
		g.reticle.SetTextureFrame(2)
		op = &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest
		op.GeoM.Scale(rScale, rScale)
		op.GeoM.Translate(float64(rect.Min.X-r.W/2), float64(rect.Max.Y-r.W/2))
		screen.DrawImage(g.reticle.Texture(), op)

		// bottom right corner
		g.reticle.SetTextureFrame(3)
		op = &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest
		op.GeoM.Scale(rScale, rScale)
		op.GeoM.Translate(float64(rect.Max.X-r.W/2), float64(rect.Max.Y-r.W/2))
		screen.DrawImage(g.reticle.Texture(), op)
	}
}
