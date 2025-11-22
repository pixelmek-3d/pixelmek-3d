package missionimage

import (
	"errors"
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/mapimage"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/shapes"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/sprites"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
	"github.com/pixelmek-3d/pixelmek-3d/game/texture"

	log "github.com/sirupsen/logrus"
)

type MissionImageOptions struct {
	RenderDropZone      bool
	RenderNavPoints     bool
	RenderEnemyUnits    bool
	RenderFriendlyUnits bool
}

type missionImage struct {
	mission     *model.Mission
	image       *ebiten.Image
	res         *model.ModelResources
	tex         *texture.TextureHandler
	mapOpts     mapimage.MapImageOptions
	missionOpts MissionImageOptions
}

func NewMissionImage(m *model.Mission, res *model.ModelResources, tex *texture.TextureHandler, mapOpts mapimage.MapImageOptions, missionOpts MissionImageOptions) (*ebiten.Image, error) {
	if m == nil || tex == nil {
		return nil, errors.New("mission image called with nil mission or texture handler")
	}
	mapImage, err := mapimage.NewMapImage(m.Map(), tex, mapOpts)
	if err != nil {
		return nil, err
	}

	missionImage := &missionImage{
		mission:     m,
		image:       mapImage,
		res:         res,
		tex:         tex,
		mapOpts:     mapOpts,
		missionOpts: missionOpts,
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
	for _, u := range m.Mechs {
		if (u.Team >= 0 && !missionOpts.RenderEnemyUnits) || (u.Team < 0 && !missionOpts.RenderFriendlyUnits) {
			continue
		}
		if err := renderMissionUnit[model.Mech](missionImage, u); err != nil {
			return nil, err
		}
	}
	for _, u := range m.Vehicles {
		if (u.Team >= 0 && !missionOpts.RenderEnemyUnits) || (u.Team < 0 && !missionOpts.RenderFriendlyUnits) {
			continue
		}
		if err := renderMissionUnit[model.Vehicle](missionImage, u); err != nil {
			return nil, err
		}
	}
	for _, u := range m.VTOLs {
		if (u.Team >= 0 && !missionOpts.RenderEnemyUnits) || (u.Team < 0 && !missionOpts.RenderFriendlyUnits) {
			continue
		}
		if err := renderMissionUnit[model.VTOL](missionImage, u); err != nil {
			return nil, err
		}
	}
	for _, u := range m.Infantry {
		if (u.Team >= 0 && !missionOpts.RenderEnemyUnits) || (u.Team < 0 && !missionOpts.RenderFriendlyUnits) {
			continue
		}
		if err := renderMissionUnit[model.Infantry](missionImage, u); err != nil {
			return nil, err
		}
	}
	for _, u := range m.Emplacements {
		if (u.Team >= 0 && !missionOpts.RenderEnemyUnits) || (u.Team < 0 && !missionOpts.RenderFriendlyUnits) {
			continue
		}
		if err := renderMissionUnit[model.Emplacement](missionImage, u); err != nil {
			return nil, err
		}
	}

	return mapImage, nil
}

func renderMissionUnit[T model.AnyUnitModel](m *missionImage, missionUnit model.MissionUnitInterface) error {
	var s *sprites.Sprite

	var t T
	switch any(t).(type) {
	case model.Mech:
		r, err := m.res.GetMechResource(missionUnit.GetUnit())
		if err != nil {
			log.Error("error loading unit resource: ", missionUnit.GetUnit())
		}
		u := model.NewMech(r)
		position := missionUnit.GetPosition()
		u.SetPos(&position)

		relPath := fmt.Sprintf("%s/%s", model.MechResourceType, r.Image)
		img := resources.GetSpriteFromFile(relPath)
		s = sprites.NewMechSprite(u, u.PixelScale(), img).Sprite
	case model.Vehicle:
		r, err := m.res.GetVehicleResource(missionUnit.GetUnit())
		if err != nil {
			log.Error("error loading unit resource: ", missionUnit.GetUnit())
		}
		u := model.NewVehicle(r)
		position := missionUnit.GetPosition()
		u.SetPos(&position)

		relPath := fmt.Sprintf("%s/%s", model.VehicleResourceType, r.Image)
		img := resources.GetSpriteFromFile(relPath)
		s = sprites.NewVehicleSprite(u, u.PixelScale(), img).Sprite
	case model.Infantry:
		r, err := m.res.GetInfantryResource(missionUnit.GetUnit())
		if err != nil {
			log.Error("error loading unit resource: ", missionUnit.GetUnit())
		}
		u := model.NewInfantry(r)
		position := missionUnit.GetPosition()
		u.SetPos(&position)

		relPath := fmt.Sprintf("%s/%s", model.InfantryResourceType, r.Image)
		img := resources.GetSpriteFromFile(relPath)
		s = sprites.NewInfantrySprite(u, u.PixelScale(), img).Sprite
	case model.VTOL:
		r, err := m.res.GetVTOLResource(missionUnit.GetUnit())
		if err != nil {
			log.Error("error loading unit resource: ", missionUnit.GetUnit())
		}
		u := model.NewVTOL(r)
		position := missionUnit.GetPosition()
		u.SetPos(&position)

		relPath := fmt.Sprintf("%s/%s", model.VTOLResourceType, r.Image)
		img := resources.GetSpriteFromFile(relPath)
		s = sprites.NewVTOLSprite(u, u.PixelScale(), img).Sprite
	case model.Emplacement:
		r, err := m.res.GetEmplacementResource(missionUnit.GetUnit())
		if err != nil {
			log.Error("error loading unit resource: ", missionUnit.GetUnit())
		}
		u := model.NewEmplacement(r)
		position := missionUnit.GetPosition()
		u.SetPos(&position)

		relPath := fmt.Sprintf("%s/%s", model.EmplacementResourceType, r.Image)
		img := resources.GetSpriteFromFile(relPath)
		s = sprites.NewEmplacementSprite(u, u.PixelScale(), img).Sprite
	default:
		return fmt.Errorf("mission unit model type not implemented: %T", t)
	}

	pxPerCell := m.mapOpts.PxPerCell
	_, mapHeight := m.mission.Map().Size()
	u := s.Unit()
	uColor := color.NRGBA{R: 255, G: 0, B: 0, A: 255}
	if u.Team() < 0 {
		uColor = color.NRGBA{R: 0, G: 255, B: 0, A: 255}
	}

	uX, uY := float32(u.Pos().X), float32(u.Pos().Y)
	vector.DrawFilledCircle(m.image, uX*float32(pxPerCell), (float32(mapHeight)-uY)*float32(pxPerCell), float32(pxPerCell)/2, uColor, false)

	// render scaled unit image only if pxPerCell is high enough for detail
	if pxPerCell >= 16 {
		x, y := float64(uX), float64(uY)
		uW, uH := float64(u.PixelWidth()), float64(u.PixelHeight())
		pxScaleX, pxScaleY := u.PixelScale()*float64(pxPerCell)/float64(uW), u.PixelScale()*float64(pxPerCell)/float64(uH)
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest
		op.GeoM.Scale(pxScaleY, pxScaleY)
		op.GeoM.Translate((x*float64(pxPerCell))-(pxScaleX*uW/2), ((float64(mapHeight)-y)*float64(pxPerCell))-(pxScaleY*uH))
		m.image.DrawImage(s.Texture(), op)
	}
	return nil
}
