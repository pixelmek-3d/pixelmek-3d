package model

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go"
)

type Effect struct {
	*Sprite
	LoopCount int
}

func NewAnimatedEffect(
	x, y, scale float64, animationRate int, img *ebiten.Image, columns, rows int, anchor raycaster.SpriteAnchor, loopCount int,
) *Effect {
	mapColor := color.RGBA{0, 0, 0, 0}
	e := &Effect{
		Sprite:    NewAnimatedSprite(x, y, scale, animationRate, img, mapColor, columns, rows, anchor, 0),
		LoopCount: loopCount,
	}

	return e
}
