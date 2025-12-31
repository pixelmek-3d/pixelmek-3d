package game

import (
	"math"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
	"github.com/pixelmek-3d/pixelmek-3d/game/texture"
)

func (g *Game) LoadMission(missionFile string) (*model.Mission, error) {
	mission, err := model.LoadMission(missionFile)
	if err != nil {
		return nil, err
	}
	g.mission = mission
	return mission, err
}

func (g *Game) initMission() {
	if g.mission == nil {
		panic("g.mission must be set before initMission!")
	}

	missionMap := g.mission.Map()

	// reload texture handler
	if g.tex != nil {
		g.initRenderFloorTex = g.tex.RenderFloorTex()
	}
	g.tex = texture.NewTextureHandler(missionMap)
	g.tex.SetRenderFloorTex(g.initRenderFloorTex)

	// clear mission sprites
	g.sprites.Clear()

	g.collisionMap = missionMap.GenerateWallCollisionLines(clipDistance)
	g.mapWidth, g.mapHeight = missionMap.Size()

	// load map and mission content
	g.loadContent()

	// initialize objectives
	g.objectives = NewObjectivesHandler(g, g.mission.Objectives)

	// init player at DZ
	pX, pY, pDegrees := g.mission.DropZone.Position[0], g.mission.DropZone.Position[1], g.mission.DropZone.Heading
	pHeading := model.CardinalToAngle(pDegrees)
	g.player.SetPos(&geom.Vector2{X: pX, Y: pY})
	g.player.SetHeading(pHeading)
	g.player.SetTargetHeading(pHeading)
	g.player.SetTurretAngle(pHeading)
	g.player.cameraAngle = pHeading
	g.player.cameraPitch = 0

	// init player as powered off but booting up
	g.player.SetPowered(model.POWER_ON)

	// init player armament for display
	if armament := g.GetHUDElement(HUD_ARMAMENT); armament != nil {
		armament.(*render.Armament).SetWeapons(g.player.Armament())
	}

	// initial mouse position to establish delta
	g.mouseX, g.mouseY = math.MinInt32, math.MinInt32

	//--init camera and renderer--//
	g.camera = raycaster.NewCamera(g.renderWidth, g.renderHeight, resources.TexSize, g.mission.Map(), g.tex)
	g.camera.SetRenderDistance(g.renderDistance)
	g.camera.SetAlwaysSetSpriteScreenRect(true)

	if len(g.mission.Map().FloorBox.Image) > 0 {
		g.camera.SetFloorTexture(resources.GetTextureFromFile(g.mission.Map().FloorBox.Image))
	}
	if len(g.mission.Map().SkyBox.Image) > 0 {
		g.camera.SetSkyTexture(resources.GetTextureFromFile(g.mission.Map().SkyBox.Image))
	}

	// init camera lighting from map settings
	g.lightFalloff = g.mission.Map().Lighting.Falloff
	g.globalIllumination = g.mission.Map().Lighting.Illumination
	g.minLightRGB, g.maxLightRGB = g.mission.Map().Lighting.LightRGB()

	g.camera.SetLightFalloff(g.lightFalloff)
	g.camera.SetGlobalIllumination(g.globalIllumination)
	g.camera.SetLightRGB(*g.minLightRGB, *g.maxLightRGB)

	// initialize camera to player position
	g.updatePlayerCamera(true)
	g.setFovAngle(g.fovDegrees)
	g.fovDepth = g.camera.FovDepth()

	g.zoomFovDepth = 2.0

	// initialize clutter
	if g.clutter != nil {
		g.clutter.Update(g, true)
	}

	// initialize AI
	g.ai = NewAIHandler(g)
}
