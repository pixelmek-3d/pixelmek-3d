package game

import (
	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
)

// Update camera to match player position and orientation
func (g *Game) updatePlayerCamera(forceUpdate bool) {
	if g.player.debugCameraTarget != nil {
		forceUpdate = true
	}
	if !g.player.moved && !forceUpdate {
		// only update camera position if player moved or forceUpdate set
		return
	}

	// reset player moved flag to only update camera when necessary
	g.player.moved = false

	camPos, camPosZ, camAngle, camPitch := g.player.CameraPosition()

	g.camera.SetPosition(camPos)
	g.camera.SetPositionZ(camPosZ)
	g.camera.SetHeadingAngle(camAngle)
	g.camera.SetPitchAngle(camPitch)
}

// CameraLineTo returns a 2D Line of the current camera position to the given position
func (g *Game) CameraLineTo(posX, posY float64) *geom.Line {
	camX, camY := g.camera.GetPosition().X, g.camera.GetPosition().Y
	return &geom.Line{
		X1: camX, Y1: camY,
		X2: posX, Y2: posY,
	}
}

// CameraLine3dTo returns a 3D Line of the current camera position to the given position
func (g *Game) CameraLine3dTo(posX, posY, posZ float64) *geom3d.Line3d {
	camX, camY, camZ := g.camera.GetPosition().X, g.camera.GetPosition().Y, g.camera.GetPositionZ()
	return &geom3d.Line3d{
		X1: camX, Y1: camY, Z1: camZ,
		X2: posX, Y2: posY, Z2: posZ,
	}
}

// clampToCameraSpriteView returns clamp x/y to appear in front of sprite relative to camera view
func (g *Game) clampToCameraSpriteView(x, y, spriteX, spriteY float64) (clampX, clampY float64) {
	clampX, clampY = x, y

	posLine := g.CameraLineTo(x, y)
	spriteLine := g.CameraLineTo(spriteX, spriteY)

	posDist, spriteDist := posLine.Distance(), spriteLine.Distance()

	if posDist >= spriteDist {
		// clamp the position toward the camera relative to the sprite
		clampLine := geom.LineFromAngle(
			posLine.X1, posLine.Y1, posLine.Angle(), spriteDist-(posDist-spriteDist)-0.01,
		)
		clampX, clampY = clampLine.X2, clampLine.Y2
	}
	return
}
