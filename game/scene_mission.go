package game

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type MissionScene struct {
	Game *Game
}

func NewMissionScene(g *Game) *MissionScene {
	// prepare for battle
	gameMenu := createGameMenu(g)
	g.menu = gameMenu
	g.closeMenu()

	// TODO: load mission

	return &MissionScene{
		Game: g,
	}
}

func (s *MissionScene) Update() error {
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
		g.updatePlayer()
		g.updateProjectiles()
		g.updateSprites()

		if g.clutter != nil {
			g.clutter.Update(g, false)
		}

		// handle player weapon updates
		g.updateWeaponCooldowns(g.player.Unit)

		// handle player camera movement
		g.updatePlayerCamera(false)
	}

	// update the menu (if active)
	g.menu.Update()

	return nil
}

func (s *MissionScene) Draw(screen *ebiten.Image) {
	g := s.Game

	// Put projectiles together with sprites for raycasting both as sprites
	raycastSprites := g.getRaycastSprites()

	// Update camera (calculate raycast)
	g.camera.Update(raycastSprites)

	// Render raycast scene
	g.camera.Draw(g.rayScene)

	// draw raycasted scene
	op := &ebiten.DrawImageOptions{}
	if g.renderScale < 1 {
		op.Filter = ebiten.FilterNearest
		op.GeoM.Scale(1/g.renderScale, 1/g.renderScale)
	}
	screen.DrawImage(g.rayScene, op)

	// store raycasted convergence point for next Update
	g.player.convergenceDistance = g.camera.GetConvergenceDistance()
	g.player.convergencePoint = g.camera.GetConvergencePoint()
	g.player.convergenceSprite = getSpriteFromInterface(g.camera.GetConvergenceSprite())

	// draw HUD elements
	g.drawHUD(screen)

	// draw menu (if active)
	g.menu.Draw(screen)
}
