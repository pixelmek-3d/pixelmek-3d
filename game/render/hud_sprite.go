package render

import (
	"github.com/hajimehoshi/ebiten/v2"
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

func NewHUDSprite(img *ebiten.Image, scale float64) *BasicHUD {
	b := &BasicHUD{
		scale: scale,
	}

	b.w, b.h = img.Size()
	b.textures, _ = GetSpriteSheetSlices(img, 1, 1)

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