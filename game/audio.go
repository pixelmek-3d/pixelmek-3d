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
	bgm *BGMHandler
	sfx *SFXHandler
}

type BGMHandler struct {
	channel *resound.DSPChannel
	player  *resound.DSPPlayer
}

type SFXHandler struct {
	engineChannel *resound.DSPChannel
	enginePlayer  *resound.DSPPlayer
}

func init() {
	audio.NewContext(resources.SampleRate)
}

func NewAudioHandler() *AudioHandler {
	a := &AudioHandler{}

	a.bgm = &BGMHandler{}
	a.bgm.channel = resound.NewDSPChannel()
	a.bgm.channel.Add("volume", resound.NewVolume(nil))
	a.SetMusicVolume(bgmVolume)

	a.sfx = &SFXHandler{}
	a.sfx.engineChannel = resound.NewDSPChannel()
	a.sfx.engineChannel.Add("volume", resound.NewVolume(nil))
	a.SetSFXVolume(sfxVolume)

	return a
}

func (a *AudioHandler) MusicVolume() float64 {
	return bgmVolume
}

func (a *AudioHandler) SetMusicVolume(strength float64) {
	bgmVolume = strength
	v := a.bgm.channel.Effects["volume"].(*resound.Volume)
	v.SetStrength(bgmVolume)
}

func (a *AudioHandler) SFXVolume() float64 {
	return sfxVolume
}

func (a *AudioHandler) SetSFXVolume(strength float64) {
	sfxVolume = strength
	v := a.sfx.engineChannel.Effects["volume"].(*resound.Volume)
	v.SetStrength(bgmVolume)
}

func (a *AudioHandler) IsMusicPlaying() bool {
	return a.bgm.player != nil && a.bgm.player.IsPlaying()
}

func (a *AudioHandler) StopMusic() {
	if a.bgm.player != nil {
		a.bgm.player.Close()
		a.bgm.player = nil
	}
}

func (a *AudioHandler) PauseMusic() {
	if a.bgm.player != nil && a.bgm.player.IsPlaying() {
		a.bgm.player.Pause()
	}
}

func (a *AudioHandler) ResumeMusic() {
	if a.bgm.player != nil && !a.bgm.player.IsPlaying() {
		a.bgm.player.Play()
	}
}

func (a *AudioHandler) StopSFX() {
	if a.sfx.enginePlayer != nil {
		a.sfx.enginePlayer.Close()
		a.sfx.enginePlayer = nil
	}
}

func (a *AudioHandler) PauseSFX() {
	if a.sfx.enginePlayer != nil && a.sfx.enginePlayer.IsPlaying() {
		a.sfx.enginePlayer.Pause()
	}
}

func (a *AudioHandler) ResumeSFX() {
	if a.sfx.enginePlayer != nil && !a.sfx.enginePlayer.IsPlaying() {
		a.sfx.enginePlayer.Play()
	}
}

func (a *AudioHandler) StartMenuMusic() {
	a.StartMusicFromFile("audio/music/soundflakes_crossing-horizon.mp3")
}

func (a *AudioHandler) StartMusicFromFile(path string) {
	if a.bgm.player != nil {
		a.StopMusic()
	}

	stream, length, err := resources.NewAudioStreamFromFile(path)
	if err != nil {
		log.Error("Error loading music:")
		log.Error(err)
		return
	}

	bgm := audio.NewInfiniteLoop(stream, length)
	vol := resound.NewVolume(bgm)
	a.bgm.player = a.bgm.channel.CreatePlayer(vol)
	a.bgm.player.SetBufferSize(time.Millisecond * 100)
	a.bgm.player.Play()
}

func (a *AudioHandler) StartEngineAmbience() {
	if a.sfx.enginePlayer != nil {
		a.sfx.enginePlayer.Close()
	}

	stream, length, err := resources.NewAudioStreamFromFile("audio/sfx/ambience-engine.ogg")
	if err != nil {
		log.Error("Error loading engine ambience:")
		log.Error(err)
		a.sfx.enginePlayer = nil
		return
	}

	engAmb := audio.NewInfiniteLoop(stream, length)
	vol := resound.NewVolume(engAmb)
	a.sfx.enginePlayer = a.sfx.engineChannel.CreatePlayer(vol)
	a.sfx.enginePlayer.SetBufferSize(time.Millisecond * 50)
	a.sfx.enginePlayer.Play()
}
