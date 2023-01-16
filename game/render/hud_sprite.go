package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	_colorDefaultRed    = color.RGBA{R: 225, G: 0, B: 0, A: 255}
	_colorDefaultGreen  = color.RGBA{R: 0, G: 214, B: 0, A: 255}
	_colorDefaultBlue   = color.RGBA{R: 0, G: 0, B: 203, A: 255}
	_colorDefaultYellow = color.RGBA{R: 255, G: 206, B: 0, A: 255}
	_colorEnemy         = color.RGBA{R: 255, G: 0, B: 12, A: 255}
)

type HUDSprite interface {
	Width() int
	Height() int
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
	RenderScale      float64
	MarginX, MarginY int
	UseCustomColor   bool
	Color            color.RGBA
}

func NewHUDSprite(img *ebiten.Image, scale float64) *BasicHUD {
	b := &BasicHUD{
		scale: scale,
	}

	if img != nil {
		b.w, b.h = img.Size()
		b.textures, _ = GetSpriteSheetSlices(img, 1, 1)
	}

	return b
}

func NewHUDSpriteFromSheet(img *ebiten.Image, scale float64, columns, rows, frameIndex int) *BasicHUD {
	b := &BasicHUD{
		scale:  scale,
		texNum: frameIndex,
	}

	w, h := img.Size()
	wFloat, hFloat := float64(w)/float64(columns), float64(h)/float64(rows)
	b.w, b.h = int(wFloat), int(hFloat)

	b.textures, _ = GetSpriteSheetSlices(img, columns, rows)

	return b
}

func (h *BasicHUD) Width() int {
	return h.w
}

func (h *BasicHUD) Height() int {
	return h.h
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
