package sprites

import (
	"fmt"
	"sync"

	"github.com/harbdog/raycaster-go"
)

type SpriteType int

const (
	MapSpriteType SpriteType = iota
	MechSpriteType
	VehicleSpriteType
	VTOLSpriteType
	InfantrySpriteType
	EmplacementSpriteType
	ProjectileSpriteType
	EffectSpriteType
	TotalSpriteTypes
)

type SpriteHandler struct {
	sprites map[SpriteType]*sync.Map

	MechSpriteTemplates        map[string]*MechSprite
	VehicleSpriteTemplates     map[string]*VehicleSprite
	VTOLSpriteTemplates        map[string]*VTOLSprite
	InfantrySpriteTemplates    map[string]*InfantrySprite
	EmplacementSpriteTemplates map[string]*EmplacementSprite
	ProjectileSpriteTemplates  map[string]*ProjectileSprite
}

func NewSpriteHandler() *SpriteHandler {
	s := &SpriteHandler{
		sprites:                    make(map[SpriteType]*sync.Map, TotalSpriteTypes),
		MechSpriteTemplates:        make(map[string]*MechSprite),
		VehicleSpriteTemplates:     make(map[string]*VehicleSprite),
		VTOLSpriteTemplates:        make(map[string]*VTOLSprite),
		InfantrySpriteTemplates:    make(map[string]*InfantrySprite),
		EmplacementSpriteTemplates: make(map[string]*EmplacementSprite),
		ProjectileSpriteTemplates:  make(map[string]*ProjectileSprite),
	}
	s.sprites[MechSpriteType] = &sync.Map{}
	s.sprites[VehicleSpriteType] = &sync.Map{}
	s.sprites[VTOLSpriteType] = &sync.Map{}
	s.sprites[InfantrySpriteType] = &sync.Map{}
	s.sprites[EmplacementSpriteType] = &sync.Map{}
	s.sprites[MapSpriteType] = &sync.Map{}
	s.sprites[ProjectileSpriteType] = &sync.Map{}
	s.sprites[EffectSpriteType] = &sync.Map{}

	return s
}

func (s *SpriteHandler) SpriteTypes() []SpriteType {
	typesArr := make([]SpriteType, 0, len(s.sprites))
	for spriteType := range s.sprites {
		typesArr = append(typesArr, spriteType)
	}
	return typesArr
}

func (s *SpriteHandler) Range(f func(key, value any) bool) {
	for _, spriteMap := range s.sprites {
		spriteMap.Range(f)
	}
}

func (s *SpriteHandler) RangeByType(spriteType SpriteType, f func(key, value any) bool) {
	if spriteMap, ok := s.sprites[spriteType]; ok {
		spriteMap.Range(f)
	}
}

func (s *SpriteHandler) Clear() {
	for spriteType := range s.sprites {
		s.sprites[spriteType] = &sync.Map{}
	}
}

func GetSpriteType(sInterface raycaster.Sprite) SpriteType {
	switch interfaceType := sInterface.(type) {
	case *Sprite:
		return MapSpriteType
	case *MechSprite:
		return MechSpriteType
	case *VehicleSprite:
		return VehicleSpriteType
	case *VTOLSprite:
		return VTOLSpriteType
	case *InfantrySprite:
		return InfantrySpriteType
	case *EmplacementSprite:
		return EmplacementSpriteType
	case *ProjectileSprite:
		return ProjectileSpriteType
	case *EffectSprite:
		return EffectSpriteType
	default:
		panic(fmt.Errorf("get SpriteType from sprite not implemented: %v", interfaceType))
	}
}

func (s *SpriteHandler) AddMapSprite(sprite *Sprite) {
	s.sprites[MapSpriteType].Store(sprite, struct{}{})
}

func (s *SpriteHandler) DeleteMapSprite(sprite *Sprite) {
	s.sprites[MapSpriteType].Delete(sprite)
}

func (s *SpriteHandler) AddMechSprite(mech *MechSprite) {
	s.sprites[MechSpriteType].Store(mech, struct{}{})
}

func (s *SpriteHandler) DeleteMechSprite(mech *MechSprite) {
	s.sprites[MechSpriteType].Delete(mech)
}

func (s *SpriteHandler) AddVehicleSprite(vehicle *VehicleSprite) {
	s.sprites[VehicleSpriteType].Store(vehicle, struct{}{})
}

func (s *SpriteHandler) DeleteVehicleSprite(vehicle *VehicleSprite) {
	s.sprites[VehicleSpriteType].Delete(vehicle)
}

func (s *SpriteHandler) AddVTOLSprite(vtol *VTOLSprite) {
	s.sprites[VTOLSpriteType].Store(vtol, struct{}{})
}

func (s *SpriteHandler) DeleteVTOLSprite(vtol *VTOLSprite) {
	s.sprites[VTOLSpriteType].Delete(vtol)
}

func (s *SpriteHandler) AddInfantrySprite(infantry *InfantrySprite) {
	s.sprites[InfantrySpriteType].Store(infantry, struct{}{})
}

func (s *SpriteHandler) DeleteInfantrySprite(infantry *InfantrySprite) {
	s.sprites[InfantrySpriteType].Delete(infantry)
}

func (s *SpriteHandler) AddEmplacementSprite(emplacement *EmplacementSprite) {
	s.sprites[EmplacementSpriteType].Store(emplacement, struct{}{})
}

func (s *SpriteHandler) DeleteEmplacementSprite(emplacement *EmplacementSprite) {
	s.sprites[EmplacementSpriteType].Delete(emplacement)
}

func (s *SpriteHandler) AddProjectile(projectile *ProjectileSprite) {
	s.sprites[ProjectileSpriteType].Store(projectile, struct{}{})
}

func (s *SpriteHandler) DeleteProjectile(projectile *ProjectileSprite) {
	s.sprites[ProjectileSpriteType].Delete(projectile)
}

func (s *SpriteHandler) AddEffect(effect *EffectSprite) {
	s.sprites[EffectSpriteType].Store(effect, struct{}{})
}

func (s *SpriteHandler) DeleteEffect(effect *EffectSprite) {
	s.sprites[EffectSpriteType].Delete(effect)
}
