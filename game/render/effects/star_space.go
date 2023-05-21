package effects

import (
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	scale      = 64
	starsCount = 256
)

type StarSpace struct {
	screenWidth, screenHeight float32
	stars                     [starsCount]Star
}

func NewStarSpace(screenWidth, screenHeight int) *StarSpace {
	s := &StarSpace{
		screenWidth:  float32(screenWidth),
		screenHeight: float32(screenHeight),
	}
	for i := 0; i < starsCount; i++ {
		s.stars[i].Init(screenWidth, screenHeight)
	}
	return s
}

func (s *StarSpace) Update() error {
	x, y := s.screenWidth/2, s.screenHeight/2
	for i := 0; i < starsCount; i++ {
		s.stars[i].Update(float32(x*scale), float32(y*scale))
	}
	return nil
}

func (s *StarSpace) Draw(screen *ebiten.Image) {
	for i := 0; i < starsCount; i++ {
		s.stars[i].Draw(screen)
	}
}

type Star struct {
	screenWidth, screenHeight          float32
	fromx, fromy, tox, toy, brightness float32
}

func (s *Star) Init(screenWidth, screenHeight int) {
	s.screenWidth, s.screenHeight = float32(screenWidth), float32(screenHeight)
	s.tox = rand.Float32() * s.screenWidth * scale
	s.fromx = s.tox
	s.toy = rand.Float32() * s.screenHeight * scale
	s.fromy = s.toy
	s.brightness = rand.Float32() * 0xff
}

func (s *Star) Update(x, y float32) {
	s.fromx = s.tox
	s.fromy = s.toy
	s.tox += (s.tox - x) / 32
	s.toy += (s.toy - y) / 32
	s.brightness += 1
	if 0xff < s.brightness {
		s.brightness = 0xff
	}
	if s.fromx < 0 || s.screenWidth*scale < s.fromx || s.fromy < 0 || s.screenHeight*scale < s.fromy {
		s.Init(int(s.screenWidth), int(s.screenHeight))
	}
}

func (s *Star) Draw(screen *ebiten.Image) {
	c := color.RGBA{
		R: uint8(0xbb * s.brightness / 0xff),
		G: uint8(0xdd * s.brightness / 0xff),
		B: uint8(0xff * s.brightness / 0xff),
		A: 0xff}
	vector.StrokeLine(screen, s.fromx/scale, s.fromy/scale, s.tox/scale, s.toy/scale, 1, c, true)
}
