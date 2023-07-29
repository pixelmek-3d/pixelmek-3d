package game

import (
	"time"

	"github.com/adrianbrad/queue"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/harbdog/pixelmek-3d/game/resources"
	"github.com/solarlune/resound"

	log "github.com/sirupsen/logrus"
)

type AudioMainSource int

const (
	AUDIO_ENGINE AudioMainSource = iota
	AUDIO_STOMP_LEFT
	AUDIO_STOMP_RIGHT
	_AUDIO_MAIN_SOURCE_COUNT
)

var (
	bgmVolume   float64
	sfxVolume   float64
	sfxChannels int = 8 // TODO: improve channel reuse and then make this a config setting
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
	mainSources []*SFXSource
	extSources  *queue.Priority[*SFXSource]
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
	a.sfx.mainSources = make([]*SFXSource, _AUDIO_MAIN_SOURCE_COUNT)
	// engine audio source file setup later since it is a looping ambient source
	a.sfx.mainSources[AUDIO_ENGINE] = NewSoundEffectSource(0.3)
	// TODO: different stomp sounds for different tonnages
	stompSound := "audio/sfx/stomp.ogg"
	a.sfx.mainSources[AUDIO_STOMP_LEFT] = NewSoundEffectSourceFromFile(stompSound, 0.6)
	a.sfx.mainSources[AUDIO_STOMP_LEFT].SetPan(-0.5)
	a.sfx.mainSources[AUDIO_STOMP_RIGHT] = NewSoundEffectSourceFromFile(stompSound, 0.6)
	a.sfx.mainSources[AUDIO_STOMP_RIGHT].SetPan(0.5)

	extInit := make([]*SFXSource, 0, sfxChannels)
	for i := 0; i < sfxChannels; i++ {
		extInit = append(extInit, NewSoundEffectSource(0.0))
	}

	a.sfx.extSources = queue.NewPriority(
		extInit,
		func(elem, other *SFXSource) bool {
			// give higher priority rating to sources that are still playing and with higher volume
			var elemRating, otherRating float64
			if elem.player != nil && elem.player.IsPlaying() {
				elemRating = elem.player.Volume()
			}
			if other.player != nil && other.player.IsPlaying() {
				otherRating = other.player.Volume()
			}

			return elemRating < otherRating
		},
		queue.WithCapacity(sfxChannels),
	)

	a.SetSFXVolume(sfxVolume)

	return a
}

func NewSoundEffectSource(sourceVolume float64) *SFXSource {
	s := &SFXSource{volume: sourceVolume}
	s.channel = resound.NewDSPChannel()
	s.channel.Add("volume", resound.NewVolume(nil).SetStrength(sourceVolume))
	s.channel.Add("pan", resound.NewPan(nil))
	return s
}

func NewSoundEffectSourceFromFile(sourceFile string, sourceVolume float64) *SFXSource {
	s := NewSoundEffectSource(sourceVolume)

	stream, _, err := resources.NewAudioStreamFromFile(sourceFile)
	if err != nil {
		log.Error("Error loading sound effect file:")
		log.Error(err)
		s.player = nil
		return s
	}

	s.player = s.channel.CreatePlayer(stream)
	s.player.SetBufferSize(time.Millisecond * 200)

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

func (s *SFXSource) Play() {
	if s.player != nil {
		s.channel.Active = true
		s.player.Rewind()
		s.player.Play()
	}
}

func (s *SFXSource) Close() {
	if s.player != nil {
		s.player.Close()
		s.player = nil
	}
	s.channel.Active = false
}

func (s *SFXSource) SetPan(panPercent float64) {
	if pan, ok := s.channel.Effects["pan"].(*resound.Pan); ok {
		pan.SetPan(panPercent)
	} else {
		s.channel.Add("pan", resound.NewPan(nil).SetPan(panPercent))
	}
}

func (a *AudioHandler) PlaySFX(sfxFile string) {
	// get and close the lowest priority source for reuse
	source, _ := a.sfx.extSources.Get()
	if source == nil {
		source = NewSoundEffectSource(0.0)
	} else {
		source.Close()
	}

	// TODO: after done trying stuff, cache somewhere for reuse
	stream, _, err := resources.NewAudioStreamFromFile(sfxFile)
	if err != nil {
		log.Error("Error playing sound effect file: " + sfxFile)
		return
	}

	source.player = source.channel.CreatePlayer(stream)
	source.player.SetBufferSize(time.Millisecond * 100)
	source.player.Play()

	a.sfx.extSources.Offer(source)
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
	for _, s := range a.sfx.mainSources {
		s.UpdateVolume()
	}

	for s := range a.sfx.extSources.Iterator() {
		s.UpdateVolume()
		a.sfx.extSources.Offer(s)
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
	for _, s := range a.sfx.mainSources {
		if s.player != nil {
			s.player.Close()
			//s.player = nil // do not want to have to reinitialize player sources
		}
	}
}

func (a *AudioHandler) PauseSFX() {
	for _, s := range a.sfx.mainSources {
		if s.player != nil {
			s.player.Pause()
		}
	}
}

func (a *AudioHandler) ResumeSFX() {
	for _, s := range a.sfx.mainSources {
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
	engine := a.sfx.mainSources[AUDIO_ENGINE]
	if engine.player != nil {
		engine.player.Close()
	}

	// TODO: different ambient angine sound for different tonnages
	stream, length, err := resources.NewAudioStreamFromFile("audio/sfx/ambience-engine.ogg")
	if err != nil {
		log.Error("Error loading engine ambience file:")
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