package missionimage

import (
	"errors"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/mapimage"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/shapes"
	"github.com/pixelmek-3d/pixelmek-3d/game/texture"
)

type MissionImageOptions struct {
	RenderDropZone      bool
	RenderNavPoints     bool
	RenderEnemyUnits    bool
	RenderFriendlyUnits bool
}

func NewMissionImage(m *model.Mission, tex *texture.TextureHandler, mapOpts mapimage.MapImageOptions, missionOpts MissionImageOptions) (*ebiten.Image, error) {
	if m == nil || tex == nil {
		return nil, errors.New("mission image called with nil mission or texture handler")
	}
	mapImage, err := mapimage.NewMapImage(m.Map(), tex, mapOpts)
	if err != nil {
		return nil, err
	}
	_, mapHeight := m.Map().Size()
	pxPerCell := mapOpts.PxPerCell

	if missionOpts.RenderDropZone {
		// draw player drop zone
		playerX, playerY := float32(m.DropZone.Position[0]), float32(m.DropZone.Position[1])
		vector.StrokeCircle(mapImage, playerX*float32(pxPerCell), (float32(mapHeight)-playerY)*float32(pxPerCell), float32(pxPerCell), 1, color.NRGBA{R: 0, G: 255, B: 0, A: 255}, false)
	}

	if missionOpts.RenderNavPoints {
		// draw mission nav points
		for _, navPoint := range m.NavPoints {
			navX, navY := float32(navPoint.Position[0]), float32(navPoint.Position[1])
			midX, midY := navX*float32(pxPerCell), ((float32(mapHeight) - navY) * float32(pxPerCell))
			navColor := color.NRGBA{R: 255, G: 206, B: 0, A: 255}
			shapes.StrokeDiamond(mapImage, midX, midY, float32(pxPerCell), float32(pxPerCell), 1, navColor, false)
		}
	}

	// draw mission units
	for _, unit := range m.Mechs {
		if (unit.Team >= 0 && !missionOpts.RenderEnemyUnits) || (unit.Team < 0 && !missionOpts.RenderFriendlyUnits) {
			continue
		}
		uX, uY := float32(unit.Position[0]), float32(unit.Position[1])
		uColor := color.NRGBA{R: 255, G: 0, B: 0, A: 255}
		if unit.Team < 0 {
			uColor = color.NRGBA{R: 0, G: 255, B: 0, A: 255}
		}
		// TODO: render scaled unit image if pxPerCell >= 16?
		vector.DrawFilledCircle(mapImage, uX*float32(pxPerCell), (float32(mapHeight)-uY)*float32(pxPerCell), float32(pxPerCell)/2, uColor, false)
	}
	// TODO: common draw func for all unit types
	for _, unit := range m.Vehicles {
		if (unit.Team >= 0 && !missionOpts.RenderEnemyUnits) || (unit.Team < 0 && !missionOpts.RenderFriendlyUnits) {
			continue
		}
		uX, uY := float32(unit.Position[0]), float32(unit.Position[1])
		uColor := color.NRGBA{R: 255, G: 0, B: 0, A: 255}
		if unit.Team < 0 {
			uColor = color.NRGBA{R: 0, G: 255, B: 0, A: 255}
		}
		vector.DrawFilledCircle(mapImage, uX*float32(pxPerCell), (float32(mapHeight)-uY)*float32(pxPerCell), float32(pxPerCell)/2, uColor, false)
	}
	for _, unit := range m.VTOLs {
		if (unit.Team >= 0 && !missionOpts.RenderEnemyUnits) || (unit.Team < 0 && !missionOpts.RenderFriendlyUnits) {
			continue
		}
		uX, uY := float32(unit.Position[0]), float32(unit.Position[1])
		uColor := color.NRGBA{R: 255, G: 0, B: 0, A: 255}
		if unit.Team < 0 {
			uColor = color.NRGBA{R: 0, G: 255, B: 0, A: 255}
		}
		vector.DrawFilledCircle(mapImage, uX*float32(pxPerCell), (float32(mapHeight)-uY)*float32(pxPerCell), float32(pxPerCell)/2, uColor, false)
	}
	for _, unit := range m.Infantry {
		if (unit.Team >= 0 && !missionOpts.RenderEnemyUnits) || (unit.Team < 0 && !missionOpts.RenderFriendlyUnits) {
			continue
		}
		uX, uY := float32(unit.Position[0]), float32(unit.Position[1])
		uColor := color.NRGBA{R: 255, G: 0, B: 0, A: 255}
		if unit.Team < 0 {
			uColor = color.NRGBA{R: 0, G: 255, B: 0, A: 255}
		}
		vector.DrawFilledCircle(mapImage, uX*float32(pxPerCell), (float32(mapHeight)-uY)*float32(pxPerCell), float32(pxPerCell)/2, uColor, false)
	}
	for _, unit := range m.Emplacements {
		if (unit.Team >= 0 && !missionOpts.RenderEnemyUnits) || (unit.Team < 0 && !missionOpts.RenderFriendlyUnits) {
			continue
		}
		uX, uY := float32(unit.Position[0]), float32(unit.Position[1])
		uColor := color.NRGBA{R: 255, G: 0, B: 0, A: 255}
		if unit.Team < 0 {
			uColor = color.NRGBA{R: 0, G: 255, B: 0, A: 255}
		}
		vector.DrawFilledCircle(mapImage, uX*float32(pxPerCell), (float32(mapHeight)-uY)*float32(pxPerCell), float32(pxPerCell)/2, uColor, false)
	}

	return mapImage, nil
}
