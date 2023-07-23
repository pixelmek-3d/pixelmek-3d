package game

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/harbdog/pixelmek-3d/game/resources"
	"github.com/solarlune/resound"

	log "github.com/sirupsen/logrus"
)

var (
	bgmVolume float64
	sfxVolume float64
)

type AudioHandler struct {
	bgmChannel *resound.DSPChannel
	bgmPlayer  *resound.DSPPlayer

	engineChannel *resound.DSPChannel
	enginePlayer  *resound.DSPPlayer
}

func init() {
	audio.NewContext(resources.SampleRate)
}

func NewAudioHandler() *AudioHandler {
	a := &AudioHandler{}
	a.bgmChannel = resound.NewDSPChannel()
	a.engineChannel = resound.NewDSPChannel()

	a.bgmChannel.Add("volume", resound.NewVolume(nil))
	a.SetMusicVolume(bgmVolume)

	a.engineChannel.Add("volume", resound.NewVolume(nil))
	a.SetMusicVolume(sfxVolume)

	return a
}

func (a *AudioHandler) MusicVolume() float64 {
	return bgmVolume
}

func (a *AudioHandler) SetMusicVolume(strength float64) {
	bgmVolume = strength
	v := a.bgmChannel.Effects["volume"].(*resound.Volume)
	v.SetStrength(bgmVolume)
}

func (a *AudioHandler) PlayMusicFromFile(path string) {
	if a.bgmPlayer != nil {
		a.bgmPlayer.Close()
	}

	stream, length, err := resources.NewAudioStreamFromFile(path)
	if err != nil {
		log.Error("Error loading music:")
		log.Error(err)
		a.bgmPlayer = nil
		return
	}

	bgm := audio.NewInfiniteLoop(stream, length)
	vol := resound.NewVolume(bgm)
	a.bgmPlayer = a.bgmChannel.CreatePlayer(vol)
	a.bgmPlayer.SetBufferSize(time.Millisecond * 100)
	a.bgmPlayer.Play()
}

func (a *AudioHandler) PlayEngineAmbience() {
	if a.enginePlayer != nil {
		a.enginePlayer.Close()
	}

	stream, length, err := resources.NewAudioStreamFromFile("audio/sfx/ambience-engine.ogg")
	if err != nil {
		log.Error("Error loading engine ambience:")
		log.Error(err)
		a.enginePlayer = nil
		return
	}

	engAmb := audio.NewInfiniteLoop(stream, length)
	vol := resound.NewVolume(engAmb)
	a.enginePlayer = a.engineChannel.CreatePlayer(vol)
	a.enginePlayer.SetBufferSize(time.Millisecond * 50)
	a.enginePlayer.Play()
}
