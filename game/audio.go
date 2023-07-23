package game

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/harbdog/pixelmek-3d/game/resources"
	"github.com/solarlune/resound"

	log "github.com/sirupsen/logrus"
)

type AudioHandler struct {
	bgmChannel *resound.DSPChannel
	bgmPlayer  *resound.DSPPlayer
}

func init() {
	audio.NewContext(resources.SampleRate)
}

func NewAudioHandler() *AudioHandler {
	a := &AudioHandler{}

	a.bgmChannel = resound.NewDSPChannel()
	a.bgmChannel.Add("volume", resound.NewVolume(nil))

	return a
}

func (a *AudioHandler) MusicVolume() float64 {
	v := a.bgmChannel.Effects["volume"].(*resound.Volume)
	return v.Strength()
}

func (a *AudioHandler) SetMusicVolume(strength float64) {
	v := a.bgmChannel.Effects["volume"].(*resound.Volume)
	v.SetStrength(strength)
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
