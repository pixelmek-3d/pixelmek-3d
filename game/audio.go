package game

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/adrianbrad/queue"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
	"github.com/solarlune/resound"
	"github.com/solarlune/resound/effects"

	log "github.com/sirupsen/logrus"
)

type AudioMainSource int

const (
	AUDIO_INTERFACE AudioMainSource = iota
	AUDIO_ENGINE
	AUDIO_STOMP_LEFT
	AUDIO_STOMP_RIGHT
	AUDIO_JUMP_JET
	_AUDIO_MAIN_SOURCE_COUNT
)

type AudioInterfaceResource string

const (
	AUDIO_BUTTON_AFF    AudioInterfaceResource = "audio/sfx/button-aff.ogg"
	AUDIO_BUTTON_NEG    AudioInterfaceResource = "audio/sfx/button-neg.ogg"
	AUDIO_BUTTON_OVER   AudioInterfaceResource = "audio/sfx/button-over.ogg"
	AUDIO_CLICK_AFF     AudioInterfaceResource = "audio/sfx/click-aff.ogg"
	AUDIO_CLICK_NEG     AudioInterfaceResource = "audio/sfx/click-neg.ogg"
	AUDIO_SELECT_TARGET AudioInterfaceResource = "audio/sfx/select-target.ogg"
)

var (
	bgmVolume   float64
	sfxVolume   float64
	sfxChannels int
)

type AudioHandler struct {
	bgm    *BGMHandler
	sfx    *SFXHandler
	sfxMap *sync.Map // map[string][]byte
}

type BGMHandler struct {
	channel *resound.DSPChannel
	player  *resound.Player
}

type SFXHandler struct {
	mainSources   []*SFXSource
	entitySources *sync.Map // map[model.Entity]*SFXSource
	extSources    *queue.Priority[*SFXSource]
	_extSFXCount  *sync.Map // map[string]float64
}

type SFXSource struct {
	channel *resound.DSPChannel
	player  *resound.Player
	volume  float64

	_sfxFile            string
	_pausedWhilePlaying bool
	_sfxType            _sfxTypeHint
}

type _sfxTypeHint int

const (
	_SFX_HINT_NONE _sfxTypeHint = iota
	_SFX_HINT_ENGINE
	_SFX_HINT_POWER_ON
	_SFX_HINT_POWER_OFF
)

func init() {
	audio.NewContext(resources.SampleRate)
}

// NewAudioHandler creates a new audio handler instance
func NewAudioHandler() *AudioHandler {
	a := &AudioHandler{}

	a.bgm = &BGMHandler{}
	a.bgm.channel = resound.NewDSPChannel()
	a.bgm.channel.AddEffect("volume", effects.NewVolume())
	a.SetMusicVolume(bgmVolume)

	a.sfxMap = &sync.Map{}
	a.sfx = &SFXHandler{}
	a.sfx.mainSources = make([]*SFXSource, _AUDIO_MAIN_SOURCE_COUNT)
	a.sfx.entitySources = &sync.Map{}

	a.sfx.mainSources[AUDIO_INTERFACE] = NewSoundEffectSource(0.8)
	// engine audio source file setup later since it is a looping ambient source
	// TODO: increase engine noise level a bit when running or at high heat levels
	a.sfx.mainSources[AUDIO_ENGINE] = NewSoundEffectSource(0.3)
	// stomp audio track to be initialized based on player unit selection
	a.sfx.mainSources[AUDIO_STOMP_LEFT] = NewSoundEffectSource(0.5)
	a.sfx.mainSources[AUDIO_STOMP_LEFT].SetPan(-0.5)
	a.sfx.mainSources[AUDIO_STOMP_RIGHT] = NewSoundEffectSource(0.5)
	a.sfx.mainSources[AUDIO_STOMP_RIGHT].SetPan(0.5)

	a.sfx.mainSources[AUDIO_JUMP_JET] = NewSoundEffectSource(0.7)
	a.sfx.mainSources[AUDIO_JUMP_JET].LoadSFX(a, "audio/sfx/jet-thrust.ogg")

	a.SetSFXChannels(sfxChannels)
	a.SetSFXVolume(sfxVolume)

	return a
}

// NewSoundEffectSource creates a new sound effect channel
func NewSoundEffectSource(sourceVolume float64) *SFXSource {
	s := &SFXSource{volume: sourceVolume}
	s.channel = resound.NewDSPChannel()
	s.channel.AddEffect("volume", effects.NewVolume().SetStrength(sourceVolume))
	s.channel.AddEffect("pan", effects.NewPan())
	return s
}

// LoadSFX loads a new sound effect player into the sound effect channel
func (s *SFXSource) LoadSFX(a *AudioHandler, sfxFile string) error {
	// make sure current source is closed before loading a new one
	s.Close()

	// use cache of audio if possible
	var audioBytes []byte
	iAudioBytes, found := a.sfxMap.Load(sfxFile)
	if audioBytesCheck, ok := iAudioBytes.([]byte); found && ok {
		audioBytes = audioBytesCheck
	} else {
		var err error
		audioBytes, err = resources.ReadFile(sfxFile)
		if err != nil {
			log.Error("Error reading sound effect file: " + sfxFile)
			return err
		}
		a.sfxMap.Store(sfxFile, audioBytes)
	}

	stream, _, err := resources.NewAudioStream(audioBytes, sfxFile)
	if err != nil {
		log.Error("Error playing sound effect file: " + sfxFile)
		return err
	}

	s.player, err = resound.NewPlayer(stream)
	if err != nil {
		return err
	}
	s.player.SetDSPChannel(s.channel)
	s.player.SetBufferSize(time.Millisecond * 100)
	s._sfxFile = sfxFile

	return nil
}

// LoadLoopSFX loads a new looping sound effect player into the sound effect channel
func (s *SFXSource) LoadLoopSFX(a *AudioHandler, sfxFile string) error {
	// make sure current source is closed before loading a new one
	s.Close()

	// use cache of audio if possible
	var audioBytes []byte
	iAudioBytes, found := a.sfxMap.Load(sfxFile)
	if audioBytesCheck, ok := iAudioBytes.([]byte); found && ok {
		audioBytes = audioBytesCheck
	} else {
		var err error
		audioBytes, err = resources.ReadFile(sfxFile)
		if err != nil {
			log.Error("Error reading looping sound effect file: " + sfxFile)
			return err
		}
		a.sfxMap.Store(sfxFile, audioBytes)
	}

	stream, length, err := resources.NewAudioStream(audioBytes, sfxFile)
	if err != nil {
		log.Error("Error playing looping sound effect file: " + sfxFile)
		return err
	}

	loop := audio.NewInfiniteLoop(stream, length)
	s.player, err = resound.NewPlayer(loop)
	if err != nil {
		return err
	}
	s.player.SetDSPChannel(s.channel)
	s.player.SetBufferSize(time.Millisecond * 100)
	s._sfxFile = sfxFile

	return nil
}

// UpdateVolume updates the volume of the sound channel taking into account relative volume modifier
func (s *SFXSource) UpdateVolume() {
	if vol, ok := s.channel.Effects["volume"].(*effects.Volume); ok {
		vol.SetStrength(sfxVolume * s.volume)
	} else {
		s.channel.AddEffect("volume", effects.NewVolume().SetStrength(sfxVolume*s.volume))
	}
}

// SetSourceVolume sets the relative volume modifier of the sound channel
func (s *SFXSource) SetSourceVolume(sourceVolume float64) {
	s.volume = sourceVolume
	s.UpdateVolume()
}

// SetPan sets the left/right panning percent of the sound channel
func (s *SFXSource) SetPan(panPercent float64) {
	if pan, ok := s.channel.Effects["pan"].(*effects.Pan); ok {
		pan.SetPan(panPercent)
	} else {
		s.channel.AddEffect("pan", effects.NewPan().SetPan(panPercent))
	}
}

// IsPlaying returns true if the sound effect is currently playing
func (s *SFXSource) IsPlaying() bool {
	if s.player != nil {
		return s.player.IsPlaying()
	}
	return false
}

// Play starts playing the sound effect player from the beginning of the effect
func (s *SFXSource) Play() {
	s._pausedWhilePlaying = false
	if s.player != nil {
		s.player.Rewind()
		s.player.Play()
	}
}

// Pause pauses the sound effect player
func (s *SFXSource) Pause() {
	if s.player != nil {
		s._pausedWhilePlaying = s.player.IsPlaying()
		s.player.Pause()
	}
}

// Resume resumes the sound effect player without rewinding
func (s *SFXSource) Resume() {
	if s.player != nil && s._pausedWhilePlaying {
		s.player.Play()
	}
}

// Close stops and closes the sound effect player
func (s *SFXSource) Close() {
	s._pausedWhilePlaying = false
	if s.player != nil {
		s.player.Close()
		s.player = nil
	}
}

// PlaySFX plays given external sound effect file
func (a *AudioHandler) PlaySFX(sfxFile string, sourceVolume, panPercent float64) {
	if sfxVolume <= 0 || sourceVolume <= 0 {
		return
	}

	// get and close the lowest priority source for reuse
	source, _ := a.sfx.extSources.Get()
	if source == nil {
		source = NewSoundEffectSource(0.0)
	} else {
		// decrement count of the previous sound effect
		a.sfx._updateExtSFXCount(source._sfxFile, -1)
		source.Close()
	}

	source.SetSourceVolume(sourceVolume)
	source.SetPan(panPercent)

	source.LoadSFX(a, sfxFile)
	source.Play()

	// increment count of this sound effect
	a.sfx._updateExtSFXCount(sfxFile, 1)
	a.sfx.extSources.Offer(source)
}

// PlayLoopEntitySFX plays given looping sound effect as emitted from an Entity object, if not already playing
func (a *AudioHandler) PlayLoopEntitySFX(sfxFile string, entity model.Entity, sourceVolume, panPercent float64) {
	if sfxVolume <= 0 || sourceVolume <= 0 {
		return
	}

	// check if entity is already playing a looping source to update instead of playing as new source
	var source *SFXSource
	a.sfx.entitySources.Range(func(k, v interface{}) bool {
		if entity == k.(model.Entity) {
			source = v.(*SFXSource)
			return false
		}
		return true
	})

	// get and close the lowest priority source for reuse
	if source == nil {
		source = NewSoundEffectSource(0.0)
	} else if source._sfxFile != sfxFile {
		// close out the source to play a new source
		source.Close()
	}

	// update volume and panning, even if continuing to play current loop
	source.SetSourceVolume(sourceVolume)
	source.SetPan(panPercent)

	if source._sfxFile == sfxFile {
		// TODO: any better way to get the volume and pan to update in a playing loop?
		source.player.Pause()
		source.player.Play()
	} else {
		source.LoadLoopSFX(a, sfxFile)
		source.Play()
	}

	// store the sound effect against the Entity
	a.sfx.entitySources.Store(entity, source)
}

// StopLoopEntitySFX stops given looping sound effect as emitted from an Entity object
func (a *AudioHandler) StopLoopEntitySFX(sfxFile string, entity model.Entity) {
	var source *SFXSource
	a.sfx.entitySources.Range(func(k, v interface{}) bool {
		if entity == k.(model.Entity) {
			source = v.(*SFXSource)
			return false
		}
		return true
	})

	if source != nil && sfxFile == source._sfxFile {
		source.Close()
		source._sfxFile = ""
	}
}

// _updateExtSFXCount is used to keep track of duplicate sound effects being played to prioritize channel reuse
func (s *SFXHandler) _updateExtSFXCount(sfxFile string, countDiff int) {
	var newCount float64
	if value, ok := s._extSFXCount.Load(sfxFile); ok {
		if count, ok := value.(float64); ok {
			newCount = count
		}
	}
	newCount += float64(countDiff)
	if newCount < 0 {
		newCount = 0
	}
	s._extSFXCount.Store(sfxFile, newCount)
}

// SetMusicVolume sets volume of background music
func (a *AudioHandler) SetMusicVolume(strength float64) {
	bgmVolume = strength
	v := a.bgm.channel.Effects["volume"].(*effects.Volume)
	v.SetStrength(strength)

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
	extInit := make([]*SFXSource, 0, numChannels)
	for _ = range numChannels {
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
		a.sfxSourcePriorityCompare,
		queue.WithCapacity(numChannels),
	)
	a.sfx._extSFXCount = &sync.Map{}
}

func (a *AudioHandler) sfxSourcePriorityCompare(elem, other *SFXSource) bool {
	// give higher priority rating to sources that are still playing and with higher volume
	var elemRating, otherRating float64
	if elem.player != nil && elem.player.IsPlaying() {
		elemRating = elem.player.Volume()
	}
	if other.player != nil && other.player.IsPlaying() {
		otherRating = other.player.Volume()
	}

	// give lower priority rating to sources that have multiple of the same sound effect currently playing
	if elemRating > 0 {
		if value, ok := a.sfx._extSFXCount.Load(elem._sfxFile); ok {
			elemCount, ok := value.(float64)
			if ok {
				elemRating *= 1 / (elemCount + 1)
			}
		}
	}
	if otherRating > 0 {
		if value, ok := a.sfx._extSFXCount.Load(other._sfxFile); ok {
			otherCount, ok := value.(float64)
			if ok {
				otherRating *= 1 / (otherCount + 1)
			}
		}
	}

	return elemRating < otherRating
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
		}
	}
	a.sfx.entitySources.Range(func(_, v interface{}) bool {
		s := v.(*SFXSource)
		s.Close()
		return true
	})
	for s := range a.sfx.extSources.Iterator() {
		s.Close()
		a.sfx.extSources.Offer(s)
	}
}

// PauseSFX pauses all sound effect sources
func (a *AudioHandler) PauseSFX() {
	for _, s := range a.sfx.mainSources {
		s.Pause()
	}
	a.sfx.entitySources.Range(func(_, v interface{}) bool {
		s := v.(*SFXSource)
		s.Pause()
		return true
	})
	for s := range a.sfx.extSources.Iterator() {
		s.Pause()
		a.sfx.extSources.Offer(s)
	}
}

// ResumeSFX resumes play of all sound effect sources
func (a *AudioHandler) ResumeSFX() {
	for _, s := range a.sfx.mainSources {
		s.Resume()
	}
	a.sfx.entitySources.Range(func(_, v interface{}) bool {
		s := v.(*SFXSource)
		s.Resume()
		return true
	})
	for s := range a.sfx.extSources.Iterator() {
		s.Resume()
		a.sfx.extSources.Offer(s)
	}
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
		log.Errorf("Error loading music: %v\n", err)
		a.bgm.player = nil
		return
	}

	bgm := audio.NewInfiniteLoop(stream, length)
	a.bgm.player, err = resound.NewPlayer(bgm)
	if err != nil {
		log.Errorf("Error starting music player: %v\n", err)
		return
	}
	a.bgm.player.SetDSPChannel(a.bgm.channel)
	a.bgm.player.SetBufferSize(time.Millisecond * 100)
	a.bgm.player.Play()
}

// StartEngineAmbience starts the ambient engine audio loop
func (a *AudioHandler) StartEngineAmbience() {
	engine := a.sfx.mainSources[AUDIO_ENGINE]
	engine._sfxType = _SFX_HINT_ENGINE
	if engine.player != nil {
		engine.player.Close()
	}

	// TODO: different ambient angine sound for different tonnages/unit types
	stream, length, err := resources.NewAudioStreamFromFile("audio/sfx/ambience-engine.ogg")
	if err != nil {
		log.Errorf("Error loading engine ambience file: %v\n", err)
		engine.player = nil
		return
	}

	engAmb := audio.NewInfiniteLoop(stream, length)
	engine.player, err = resound.NewPlayer(engAmb)
	if err != nil {
		log.Errorf("Error starting engine ambience player: %v\n", err)
		return
	}
	engine.player.SetDSPChannel(engine.channel)
	engine.player.SetBufferSize(time.Millisecond * 50)
	engine.player.Play()
}

// StopEngineAmbience stop the ambient engine audio loop
func (a *AudioHandler) StopEngineAmbience() {
	engine := a.sfx.mainSources[AUDIO_ENGINE]
	if engine._sfxType == _SFX_HINT_ENGINE && engine.player != nil && engine.player.IsPlaying() {
		engine._sfxType = _SFX_HINT_NONE
		engine.player.Close()
	}
}

// IsEngineAmbience indicates whether the current engine audio is the ambience loop
func (a *AudioHandler) EngineAmbience() _sfxTypeHint {
	return a.sfx.mainSources[AUDIO_ENGINE]._sfxType
}

// PlayPowerOnSequence plays the power on sound using the engine audio source
func (a *AudioHandler) PlayPowerOnSequence() {
	engine := a.sfx.mainSources[AUDIO_ENGINE]
	engine._sfxType = _SFX_HINT_POWER_ON
	if engine.player != nil {
		engine.player.Close()
	}

	// TODO: different power on sounds for different tonnages/unit types
	engine.LoadSFX(a, "audio/sfx/power-on.ogg")
	engine.player.Play()
}

// PlayPowerOffSequence plays the power down sound using the engine audio source
func (a *AudioHandler) PlayPowerOffSequence() {
	engine := a.sfx.mainSources[AUDIO_ENGINE]
	engine._sfxType = _SFX_HINT_POWER_OFF
	if engine.player != nil {
		engine.player.Close()
	}

	engine.LoadSFX(a, "audio/sfx/power-off.ogg")
	engine.player.Play()
}

func (a *AudioHandler) SetStompSFX(sfxFile string) {
	a.sfx.mainSources[AUDIO_STOMP_LEFT].LoadSFX(a, sfxFile)
	a.sfx.mainSources[AUDIO_STOMP_RIGHT].LoadSFX(a, sfxFile)
}

func StompSFXForMech(m *model.Mech) (string, error) {
	if m == nil {
		return "", fmt.Errorf("can not get stomp SFX for nil mech")
	}
	mechClass := m.Class()
	mechStompFile := fmt.Sprintf("audio/sfx/stomp-%d.ogg", mechClass)
	return mechStompFile, nil
}

func JumpJetSFXForMech(m *model.Mech) (string, error) {
	if m == nil {
		return "", fmt.Errorf("can not get jump jet SFX for nil mech")
	}
	return "audio/sfx/jet-thrust.ogg", nil
}

// PlayButtonAudio plays the indicated button audio channel resource
func (a *AudioHandler) PlayButtonAudio(buttonResource AudioInterfaceResource) {
	sfxSource := a.sfx.mainSources[AUDIO_INTERFACE]
	sfxSource.LoadSFX(a, string(buttonResource))
	sfxSource.Play()
}

// IsButtonAudioPlaying returns true if the button audio channel is still playing
func (a *AudioHandler) IsButtonAudioPlaying() bool {
	sfxSource := a.sfx.mainSources[AUDIO_INTERFACE]
	return sfxSource.player != nil && sfxSource.player.IsPlaying()
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
		go a.PlaySFX(weapon.Audio(), 1.0, panPercent)
	}
}

// PlayExternalWeaponFireAudio plays weapon fire audio fired by units other than the player
func (a *AudioHandler) PlayExternalWeaponFireAudio(g *Game, weapon model.Weapon, extUnit model.Unit) {
	if len(weapon.Audio()) > 0 {
		// TODO: introduce volume modifier based on weapon type, classification, and size
		extPos, extPosZ := extUnit.Pos(), extUnit.PosZ()
		a.PlayExternalAudio(g, weapon.Audio(), extPos.X, extPos.Y, extPosZ, 10, 1.0)
	}
}

// PlayProjectileImpactAudio plays projectile impact audio near the player
func (a *AudioHandler) PlayProjectileImpactAudio(g *Game, p *render.ProjectileSprite) {
	impactAudio := p.ImpactAudio()
	if len(impactAudio) > 0 {
		// TODO: introduce volume modifier based on projectile's weapon type, classification, and size
		extPos, extPosZ := p.Pos(), p.PosZ()
		a.PlayExternalAudio(g, impactAudio, extPos.X, extPos.Y, extPosZ, 10, 1.0)
	}
}

// PlayEffectAudio plays effect audio near the player
func (a *AudioHandler) PlayEffectAudio(g *Game, p *render.EffectSprite) {
	fxAudio := p.AudioFile
	if len(fxAudio) > 0 {
		extPos, extPosZ := p.Pos(), p.PosZ()
		a.PlayExternalAudio(g, fxAudio, extPos.X, extPos.Y, extPosZ, 10, 0.7)
	}
}

// PlayExternalAudio plays audio that may be near the player taking into account distance/direction for volume/panning
// intensityDist - distance of 100% sound intensity before volume begins to dropoff at a rate of 1/d^2
// maxVolume - the maximum volume percent to be perceived by the player
func (a *AudioHandler) PlayExternalAudio(g *Game, sfxFile string, extPosX, extPosY, extPosZ, intensityDist, maxVolume float64) {
	camPos, _, camHeading, _ := g.player.CameraPosition()

	extLine := geom3d.Line3d{
		X1: camPos.X, Y1: camPos.Y, Z1: g.player.cameraZ,
		X2: extPosX, Y2: extPosY, Z2: extPosZ,
	}
	extDist := extLine.Distance()
	extHeading := extLine.Heading()

	relHeading := -model.AngleDistance(camHeading, extHeading)
	relPercent := 1 - (geom.HalfPi-relHeading)/geom.HalfPi

	extVolume := geom.Clamp(math.Pow(intensityDist/extDist, 2), 0.0, maxVolume)
	if extVolume > 0.05 {
		go g.audio.PlaySFX(sfxFile, extVolume, relPercent)
	}
}

// PlayEntityAudioLoop plays audio that may be near the player emitted from an Entity object
// intensityDist - distance of 100% sound intensity before volume begins to dropoff at a rate of 1/d^2
// maxVolume - the maximum volume percent to be perceived by the player
func (a *AudioHandler) PlayEntityAudioLoop(g *Game, sfxFile string, entity model.Entity, intensityDist, maxVolume float64) {
	camPos, _, camHeading, _ := g.player.CameraPosition()

	extPosX, extPosY, extPosZ := entity.Pos().X, entity.Pos().Y, entity.PosZ()
	extLine := geom3d.Line3d{
		X1: camPos.X, Y1: camPos.Y, Z1: g.player.cameraZ,
		X2: extPosX, Y2: extPosY, Z2: extPosZ,
	}
	extDist := extLine.Distance()
	extHeading := extLine.Heading()

	relHeading := -model.AngleDistance(camHeading, extHeading)
	relPercent := 1 - (geom.HalfPi-relHeading)/geom.HalfPi

	extVolume := geom.Clamp(math.Pow(intensityDist/extDist, 2), 0.0, maxVolume)
	if extVolume > 0.05 {
		go g.audio.PlayLoopEntitySFX(sfxFile, entity, extVolume, relPercent)
	}
}

// StopEntityAudioLoop stops audio emitted from an Entity object that may have been playing
func (a *AudioHandler) StopEntityAudioLoop(g *Game, sfxFile string, entity model.Entity) {
	g.audio.StopLoopEntitySFX(sfxFile, entity)
}
