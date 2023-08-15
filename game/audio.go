package game

import (
	"time"

	"github.com/adrianbrad/queue"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/pixelmek-3d/game/render"
	"github.com/harbdog/pixelmek-3d/game/resources"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
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
	sfxChannels int // TODO: improve channel performance by reusing players from the same audio file
)

type AudioHandler struct {
	bgm    *BGMHandler
	sfx    *SFXHandler
	sfxMap map[string][]byte
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

// NewAudioHandler creates a new audio handler instance
func NewAudioHandler() *AudioHandler {
	a := &AudioHandler{}

	a.bgm = &BGMHandler{}
	a.bgm.channel = resound.NewDSPChannel()
	a.bgm.channel.Add("volume", resound.NewVolume(nil))
	a.SetMusicVolume(bgmVolume)

	a.sfxMap = make(map[string][]byte, 64)
	a.sfx = &SFXHandler{}
	a.sfx.mainSources = make([]*SFXSource, _AUDIO_MAIN_SOURCE_COUNT)
	// engine audio source file setup later since it is a looping ambient source
	// TODO: increase engine noise level a bit when running or at high heat levels
	a.sfx.mainSources[AUDIO_ENGINE] = NewSoundEffectSource(0.3)
	// stomp audio track to be initialized based on player unit selection
	a.sfx.mainSources[AUDIO_STOMP_LEFT] = NewSoundEffectSource(0.5)
	a.sfx.mainSources[AUDIO_STOMP_LEFT].SetPan(-0.5)
	a.sfx.mainSources[AUDIO_STOMP_RIGHT] = NewSoundEffectSource(0.5)
	a.sfx.mainSources[AUDIO_STOMP_RIGHT].SetPan(0.5)

	a.SetSFXChannels(sfxChannels)
	a.SetSFXVolume(sfxVolume)

	return a
}

// NewSoundEffectSource creates a new sound effect channel
func NewSoundEffectSource(sourceVolume float64) *SFXSource {
	s := &SFXSource{volume: sourceVolume}
	s.channel = resound.NewDSPChannel()
	s.channel.Add("volume", resound.NewVolume(nil).SetStrength(sourceVolume))
	s.channel.Add("pan", resound.NewPan(nil))
	return s
}

// LoadSFX loads a new sound effect player into the sound effect channel
func (s *SFXSource) LoadSFX(a *AudioHandler, sfxFile string) error {
	// make sure current source is closed before loading a new one
	s.Close()

	// use cache of audio if possible
	audioBytes, found := a.sfxMap[sfxFile]
	if !found {
		var err error
		audioBytes, err = resources.ReadFile(sfxFile)
		if err != nil {
			log.Error("Error reading sound effect file: " + sfxFile)
			return err
		}
		a.sfxMap[sfxFile] = audioBytes
	}

	stream, _, err := resources.NewAudioStream(audioBytes, sfxFile)
	if err != nil {
		log.Error("Error playing sound effect file: " + sfxFile)
		return err
	}

	s.player = s.channel.CreatePlayer(stream)
	s.player.SetBufferSize(time.Millisecond * 100)

	return nil
}

// UpdateVolume updates the volume of the sound channel taking into account relative volume modifier
func (s *SFXSource) UpdateVolume() {
	v := s.channel.Effects["volume"].(*resound.Volume)
	v.SetStrength(sfxVolume * s.volume)
}

// SetSourceVolume sets the relative volume modifier of the sound channel
func (s *SFXSource) SetSourceVolume(sourceVolume float64) {
	s.volume = sourceVolume
	s.UpdateVolume()
}

// SetPan sets the left/right panning percent of the sound channel
func (s *SFXSource) SetPan(panPercent float64) {
	if pan, ok := s.channel.Effects["pan"].(*resound.Pan); ok {
		pan.SetPan(panPercent)
	} else {
		s.channel.Add("pan", resound.NewPan(nil).SetPan(panPercent))
	}
}

// Play starts playing the sound effect player from the beginning of the effect
func (s *SFXSource) Play() {
	if s.player != nil {
		s.player.Rewind()
		s.player.Play()
	}
}

// Close stops and closes the sound effect player
func (s *SFXSource) Close() {
	if s.player != nil {
		s.player.Close()
		s.player = nil
	}
}

// PlaySFX plays given external sound effect file
func (a *AudioHandler) PlaySFX(sfxFile string, sourceVolume, panPercent float64) {
	// get and close the lowest priority source for reuse
	source, _ := a.sfx.extSources.Get()
	if source == nil {
		source = NewSoundEffectSource(0.0)
	} else {
		source.Close()
	}

	source.SetSourceVolume(sourceVolume)
	source.SetPan(panPercent)

	source.LoadSFX(a, sfxFile)
	source.Play()

	a.sfx.extSources.Offer(source)
}

// SetMusicVolume sets volume of background music
func (a *AudioHandler) SetMusicVolume(strength float64) {
	bgmVolume = strength
	v := a.bgm.channel.Effects["volume"].(*resound.Volume)
	v.SetStrength(bgmVolume)

	if bgmVolume == 0 {
		a.PauseMusic()
	} else {
		a.ResumeMusic()
	}
}

// SetSFXVolume sets volume of all sound effect sources
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

// SetSFXChannels sets max number of external sound effect channels
func (a *AudioHandler) SetSFXChannels(numChannels int) {
	sfxChannels = numChannels

	extInit := make([]*SFXSource, 0, sfxChannels)
	for i := 0; i < sfxChannels; i++ {
		// reuse existing channels if available
		if a.sfx.extSources != nil && !a.sfx.extSources.IsEmpty() {
			s, _ := a.sfx.extSources.Get()
			extInit = append(extInit, s)
		} else {
			extInit = append(extInit, NewSoundEffectSource(0.0))
		}
	}

	if a.sfx.extSources != nil && !a.sfx.extSources.IsEmpty() {
		// close out any excess channels as necessary
		for s := range a.sfx.extSources.Iterator() {
			s.Close()
		}
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
}

// IsMusicPlaying return true if background music is currently playing
func (a *AudioHandler) IsMusicPlaying() bool {
	return a.bgm.player != nil && a.bgm.player.IsPlaying()
}

// StopMusic stops and closes the background music source
func (a *AudioHandler) StopMusic() {
	if a.bgm.player != nil {
		a.bgm.player.Close()
		a.bgm.player = nil
	}
}

// PauseMusic pauses play of background music
func (a *AudioHandler) PauseMusic() {
	if a.bgm.player != nil && a.bgm.player.IsPlaying() {
		a.bgm.player.Pause()
	}
}

// ResumeMusic resumes play of background music
func (a *AudioHandler) ResumeMusic() {
	if a.bgm.player != nil && !a.bgm.player.IsPlaying() {
		a.bgm.player.Play()
	}
}

// StopSFX stops and closes all sound effect sources
func (a *AudioHandler) StopSFX() {
	for _, s := range a.sfx.mainSources {
		if s.player != nil {
			s.player.Close()
			//s.player = nil // do not want to have to reinitialize main sources
		}
	}
	// TODO: stop extSources
}

// PauseSFX pauses all sound effect sources
func (a *AudioHandler) PauseSFX() {
	for _, s := range a.sfx.mainSources {
		if s.player != nil {
			s.player.Pause()
		}
	}
	// TODO: pause extSources
}

// ResumeSFX resumes play of all sound effect sources
func (a *AudioHandler) ResumeSFX() {
	for _, s := range a.sfx.mainSources {
		if s.player != nil {
			s.player.Play()
		}
	}
	// TODO: resume extSources
}

// StartMenuMusic starts main menu background music audio loop
func (a *AudioHandler) StartMenuMusic() {
	a.StartMusicFromFile("audio/music/crossing_horizon.mp3")
}

// StartMusicFromFile starts background music audio loop
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

// StartEngineAmbience starts the ambient engine audio loop
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

func (a *AudioHandler) SetStompSFX(sfxFile string) {
	a.sfx.mainSources[AUDIO_STOMP_LEFT].LoadSFX(a, sfxFile)
	a.sfx.mainSources[AUDIO_STOMP_RIGHT].LoadSFX(a, sfxFile)
}

// PlayLocalWeaponFireAudio plays weapon fire audio intended only if fired by the player unit
func (a *AudioHandler) PlayLocalWeaponFireAudio(weapon model.Weapon) {
	if len(weapon.Audio()) > 0 {
		var panPercent float64
		offsetX := -weapon.Offset().X
		switch {
		case offsetX < 0:
			// pan left
			panPercent = geom.Clamp(offsetX-0.4, -0.8, 0)
		case offsetX > 0:
			// pan right
			panPercent = geom.Clamp(offsetX+0.4, 0, 0.8)
		}
		a.PlaySFX(weapon.Audio(), 1.0, panPercent)
	}
}

// PlayExternalWeaponFireAudio plays weapon fire audio fired by units other than the player
func (a *AudioHandler) PlayExternalWeaponFireAudio(g *Game, weapon model.Weapon, extUnit model.Unit) {
	if len(weapon.Audio()) > 0 {
		// determine distance and angle of external fire

		// TODO: refactor distance calculation with similar from PlayProjectileImpactAudio
		// TODO: introduce volume modifier based on weapon type, classification, and size
		extPos, extPosZ := extUnit.Pos(), extUnit.PosZ()
		playerPos := g.player.Pos()
		playerHeading := g.player.Heading() + g.player.TurretAngle()

		pLine := geom3d.Line3d{
			X1: playerPos.X, Y1: playerPos.Y, Z1: g.player.cameraZ,
			X2: extPos.X, Y2: extPos.Y, Z2: extPosZ,
		}
		pDist := pLine.Distance()
		pHeading := pLine.Heading()

		relHeading := -model.AngleDistance(playerHeading, pHeading)
		relPercent := 1 - (geom.HalfPi-relHeading)/geom.HalfPi

		pVolume := (20 - pDist) / 20
		if pVolume > 0.05 {
			g.audio.PlaySFX(weapon.Audio(), pVolume, relPercent)
		}
	}
}

// PlayProjectileImpactAudio plays projectile impact audio near the player
func (a *AudioHandler) PlayProjectileImpactAudio(g *Game, p *render.ProjectileSprite) {
	if len(p.ImpactAudio) > 0 {
		// determine distance and player camera relative direction of impact for volume and panning
		extPos, extPosZ := p.Pos(), p.PosZ()
		playerPos := g.player.Pos()
		playerHeading := g.player.Heading() + g.player.TurretAngle()

		impactLine := geom3d.Line3d{
			X1: playerPos.X, Y1: playerPos.Y, Z1: g.player.cameraZ,
			X2: extPos.X, Y2: extPos.Y, Z2: extPosZ,
		}
		impactDist := impactLine.Distance()
		impactHeading := impactLine.Heading()

		relHeading := -model.AngleDistance(playerHeading, impactHeading)
		relPercent := 1 - (geom.HalfPi-relHeading)/geom.HalfPi

		impactVolume := (20 - impactDist) / 20
		if impactVolume > 0.05 {
			g.audio.PlaySFX(p.ImpactAudio, impactVolume, relPercent)
		}
	}
}
