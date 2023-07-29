package game

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/harbdog/pixelmek-3d/game/resources"
	"github.com/solarlune/resound"

	log "github.com/sirupsen/logrus"
)

type PlayerSource int

const (
	AUDIO_ENGINE PlayerSource = iota
	AUDIO_STOMP
	_AUDIO_PLAYER_SOURCE_COUNT
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
	playerSources []*SFXSource
}

type SFXSource struct {
	channel *resound.DSPChannel
	player  *resound.DSPPlayer
	volume  float64
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
	a.sfx.playerSources = make([]*SFXSource, _AUDIO_PLAYER_SOURCE_COUNT)
	a.sfx.playerSources[AUDIO_ENGINE] = NewSoundEffectSource(0.3)
	a.sfx.playerSources[AUDIO_STOMP] = NewSoundEffectSource(0.6)
	a.SetSFXVolume(sfxVolume)

	return a
}

func NewSoundEffectSource(sourceVolume float64) *SFXSource {
	s := &SFXSource{volume: sourceVolume}
	s.channel = resound.NewDSPChannel()
	s.channel.Add("volume", resound.NewVolume(nil).SetStrength(sourceVolume))
	return s
}

func (s *SFXSource) UpdateVolume() {
	v := s.channel.Effects["volume"].(*resound.Volume)
	v.SetStrength(sfxVolume * s.volume)
}

func (s *SFXSource) SetSourceVolume(sourceVolume float64) {
	s.volume = sourceVolume
	s.UpdateVolume()
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
	for _, s := range a.sfx.playerSources {
		s.UpdateVolume()
	}
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
	for _, s := range a.sfx.playerSources {
		if s.player != nil {
			s.player.Close()
			s.player = nil
		}
	}
}

func (a *AudioHandler) PauseSFX() {
	for _, s := range a.sfx.playerSources {
		if s.player != nil {
			s.player.Pause()
		}
	}
}

func (a *AudioHandler) ResumeSFX() {
	for _, s := range a.sfx.playerSources {
		if s.player != nil {
			s.player.Play()
		}
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
	engine := a.sfx.playerSources[AUDIO_ENGINE]
	if engine.player != nil {
		engine.player.Close()
	}

	stream, length, err := resources.NewAudioStreamFromFile("audio/sfx/ambience-engine.ogg")
	if err != nil {
		log.Error("Error loading engine ambience:")
		log.Error(err)
		engine.player = nil
		return
	}

	engAmb := audio.NewInfiniteLoop(stream, length)
	vol := resound.NewVolume(engAmb)
	engine.player = engine.channel.CreatePlayer(vol)
	engine.player.SetBufferSize(time.Millisecond * 50)
	engine.player.Play()
}
