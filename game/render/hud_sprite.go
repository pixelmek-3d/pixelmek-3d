package render

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/colors"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/sprites"
)

var (
	_colorDefaultRed    = colors.DefaultRed
	_colorDefaultGreen  = colors.DefaultGreen
	_colorDefaultBlue   = colors.DefaultBlue
	_colorDefaultYellow = colors.DefaultYellow
	_colorEnemy         = colors.Enemy
	_colorFriendly      = colors.Friendly
)

type HUDSprite interface {
	Width() int
	Height() int
	Rect() image.Rectangle
	Scale() float64
	SetScale(float64)

	Texture() *ebiten.Image
	NumTextureFrames() int
	SetTextureFrame(int)
}

type BasicHUD struct {
	w, h           int
	scale          float64
	texNum, lenTex int
	textures       []*ebiten.Image
}

type DrawHudOptions struct {
	Screen           *ebiten.Image
	HudRect          image.Rectangle
	MarginX, MarginY int
	UseCustomColor   bool
	Color            color.NRGBA
	HudUnit          model.Unit
}

// HudColor gets the color that should be used by a HUD element based on its default or custom user color setting
func (o *DrawHudOptions) HudColor(defaultColor color.NRGBA) color.NRGBA {
	if o.UseCustomColor {
		return o.Color
	} else {
		// apply custom alpha to default color
		hudColor := defaultColor
		hudColor.A = o.Color.A
		return hudColor
	}
}

func NewHUDSprite(img *ebiten.Image, scale float64) *BasicHUD {
	b := &BasicHUD{
		scale: scale,
	}

	if img != nil {
		b.w, b.h = img.Bounds().Dx(), img.Bounds().Dy()
		b.textures, _ = sprites.GetSpriteSheetSlices(img, 1, 1)
	}

	return b
}

func NewHUDSpriteFromSheet(img *ebiten.Image, scale float64, columns, rows, frameIndex int) *BasicHUD {
	b := &BasicHUD{
		scale:  scale,
		texNum: frameIndex,
	}

	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	wFloat, hFloat := float64(w)/float64(columns), float64(h)/float64(rows)
	b.w, b.h = int(wFloat), int(hFloat)

	b.textures, _ = sprites.GetSpriteSheetSlices(img, columns, rows)

	return b
}

func (h *BasicHUD) Width() int {
	return h.w
}

func (h *BasicHUD) Height() int {
	return h.h
}

func (h *BasicHUD) Rect() image.Rectangle {
	return image.Rect(0, 0, h.w, h.h)
}

func (h *BasicHUD) Scale() float64 {
	return h.scale
}

func (h *BasicHUD) SetScale(scale float64) {
	h.scale = scale
}

func (h *BasicHUD) Texture() *ebiten.Image {
	return h.textures[h.texNum]
}

func (h *BasicHUD) SetTextureFrame(texNum int) {
	h.texNum = texNum
}

func (h *BasicHUD) NumTextureFrames() int {
	return h.lenTex
}
