package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/transitions"
)

type GameScene struct {
	Game *Game

	transition SceneTransition
}

func NewGameScene(g *Game) *GameScene {
	// load mission resources and launch
	g.initMission()

	// prepare for battle
	gameMenu := createGameMenu(g)
	g.menu = gameMenu
	g.closeMenu()

	// stop menu music and sound effects
	g.audio.StopMusic()
	g.audio.StopSFX()

	// start mission music
	if len(g.mission.MusicPath) > 0 {
		g.audio.StartMusicFromFile("audio/music/" + g.mission.MusicPath)
	}

	// start engine ambience
	g.audio.PlayPowerOnSequence()

	// transition in to start game
	tOpts := &transitions.TransitionOptions{
		InDuration:   5.0,
		HoldDuration: 0,
		OutDuration:  0,
	}
	transition := transitions.NewFade(g.overlayScreen, tOpts, ebiten.GeoM{})

	return &GameScene{
		Game:       g,
		transition: transition,
	}
}

func (g *Game) LeaveGame() {
	// stop mission music and sfx audio
	g.audio.StopSFX()
	g.audio.StopMusic()

	// go to mission debrief
	g.scene = NewMissionDebriefScene(g)
}

func (s *GameScene) Update() error {
	g := s.Game

	if g.osType == osTypeBrowser && ebiten.CursorMode() == ebiten.CursorModeVisible && !g.menu.Active() && !g.menu.Closing() {
		// capture not working sometimes (https://developer.mozilla.org/en-US/docs/Web/API/Pointer_Lock_API#iframe_limitations):
		//   sm_exec.js:349 pointerlockerror event is fired. 'sandbox="allow-pointer-lock"' might be required at an iframe.
		//   This function on browsers must be called as a result of a gestural interaction or orientation change.
		//   localhost/:1 Uncaught (in promise) DOMException: The user has exited the lock before this request was completed.
		g.openMenu()
	}

	if g.menu.Closing() && !g.menu.Active() {
		// reset simple flag to make sure that if we really wanted the menu closed in browser it won't trigger reopen
		gameMenu, ok := g.menu.(*GameMenu)
		if ok {
			gameMenu.closing = false
		}
	}

	// handle input (when paused making sure only to allow input for closing menu so it can be unpaused)
	g.handleInput()

	if !g.paused {
		// Perform logical updates
		g.updateAI()
		g.updatePlayer()
		g.updateProjectiles()
		g.updateSprites()
		g.updateObjectives()

		if g.clutter != nil {
			g.clutter.Update(g, false)
		}

		// handle player weapon updates
		g.updateWeaponCooldowns(g.player.Unit)

		// handle player camera movement
		g.updatePlayerCamera(false)

		if g.InProgress() {
			if s.transition != nil {
				// update transition at start of game
				s.transition.Update()
				if s.transition.Completed() {
					s.transition = nil
				}
			}
		} else {
			if s.transition == nil {
				// transition out to leave game
				tOpts := &transitions.TransitionOptions{
					InDuration:   0.0,
					HoldDuration: 4.0,
					OutDuration:  3.0,
				}
				s.transition = transitions.NewFade(g.overlayScreen, tOpts, ebiten.GeoM{})
			} else {
				// update transition about to leave game
				s.transition.Update()
				if s.transition.Completed() {
					g.LeaveGame()
				}
			}
		}
	}

	// update the menu (if active)
	g.menu.Update()

	return nil
}

func (s *GameScene) Draw(screen *ebiten.Image) {
	g := s.Game

	// Put projectiles together with sprites for raycasting both as sprites
	raycastSprites := g.getRaycastSprites()

	// Update camera (calculate raycast)
	g.camera.Update(raycastSprites)

	// store sprite at raycasted convergence point for next Update
	g.player.convergenceSprite = getSpriteFromInterface(g.camera.GetConvergenceSprite())

	// Render raycast scene
	g.camera.Draw(g.rayScreen)

	// Draw raycast scene on render scene, scaled as needed
	if g.renderScale == 1 {
		g.renderScreen = g.rayScreen
	} else {
		rayOp := &ebiten.DrawImageOptions{}
		rayOp.Filter = ebiten.FilterNearest
		rayOp.GeoM.Scale(1/g.renderScale, 1/g.renderScale)
		g.renderScreen.DrawImage(g.rayScreen, rayOp)
	}

	if g.crtShader || g.lightAmpEngaged || g.player.ejectionPod != nil {
		// use CRT shader over raycasted scene when in ejection pod
		showCurve := (g.lightAmpEngaged || g.player.ejectionPod != nil)
		crtShader.DrawWithOptions(g.overlayScreen, g.renderScreen, showCurve)
	} else {
		g.overlayScreen.DrawImage(g.renderScreen, nil)
	}

	// draw HUD elements to overlay screen
	g.drawHUD(g.overlayScreen)

	if s.transition != nil {
		// draw transition shader to screen
		s.transition.SetImage(g.overlayScreen)
		s.transition.Draw(screen)
	} else {
		// draw HUD overlayed elements directly to screen
		screen.DrawImage(g.overlayScreen, nil)
	}

	// draw menu (if active)
	g.menu.Draw(screen)
}
