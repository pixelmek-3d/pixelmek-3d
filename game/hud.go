package game

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/pixelmek-3d/game/render"
)

func (g *Game) initInteractiveTypes() {
	g.interactiveSpriteTypes = map[SpriteType]struct{}{
		MechSpriteType:     {},
		VehicleSpriteType:  {},
		VTOLSpriteType:     {},
		InfantrySpriteType: {},
	}
}

func (g *Game) isInteractiveType(spriteType SpriteType) bool {
	if _, containsType := g.interactiveSpriteTypes[spriteType]; containsType {
		return true
	}
	return false
}

func (g *Game) drawCompass(screen *ebiten.Image) {
	if g.compass == nil {
		return
	}

	g.compass.Update(g.player.Angle())

	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.ColorM.ScaleWithColor(g.hudRGBA)

	compassScale := g.compass.Scale() * g.renderScale * g.hudScale
	op.GeoM.Scale(compassScale, compassScale)
	op.GeoM.Translate(
		float64(g.width)/2-float64(g.compass.Width())*compassScale/2,
		float64(2*g.compass.Height())*compassScale/2,
	)
	screen.DrawImage(g.compass.Texture(), op)
}

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
		float64(g.width)/2-float64(g.crosshairs.Width())*crosshairScale/2,
		float64(g.height)/2-float64(g.crosshairs.Height())*crosshairScale/2,
	)
	screen.DrawImage(g.crosshairs.Texture(), op)
}

func (g *Game) drawTargetReticle(screen *ebiten.Image) {
	if g.reticle == nil {
		return
	}

	rScale := g.reticle.Scale() * g.renderScale * g.hudScale
	rOff := rScale * float64(g.reticle.Width()) / 2

	var op *ebiten.DrawImageOptions

	// setup some common draw modifications
	geoM := ebiten.GeoM{}
	geoM.Scale(rScale, rScale)

	colorM := ebiten.ColorM{}
	colorM.ScaleWithColor(g.hudRGBA)

	for spriteType, spriteMap := range g.sprites.sprites {
		if !g.isInteractiveType(spriteType) {
			// only show on certain sprite types (skip projectiles, effects, etc.)
			continue
		}

		spriteMap.Range(func(k, _ interface{}) bool {
			var rect *image.Rectangle
			switch spriteType {
			case MechSpriteType:
				s := k.(*render.MechSprite)
				rect = s.ScreenRect()

			case VehicleSpriteType:
				s := k.(*render.VehicleSprite)
				rect = s.ScreenRect()

			case VTOLSpriteType:
				s := k.(*render.VTOLSprite)
				rect = s.ScreenRect()

			case InfantrySpriteType:
				s := k.(*render.InfantrySprite)
				rect = s.ScreenRect()
			}

			if rect == nil {
				return true
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

			return true
		})
	}
}
