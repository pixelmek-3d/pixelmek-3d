package game

import (
	"github.com/hajimehoshi/ebiten/v2"
)

func (g *Game) drawCrosshairs(screen *ebiten.Image) {
	if g.crosshairs != nil {
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
}
