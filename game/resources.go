package game

import (
	"image"
	"image/color"
	"log"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/harbdog/pixelmek-3d/game/model"
)

func getRGBAFromFile(texFile string) *image.RGBA {
	var rgba *image.RGBA
	resourcePath := filepath.Join("game", "resources", "textures")
	_, tex, err := ebitenutil.NewImageFromFile(filepath.Join(resourcePath, texFile))
	if err != nil {
		log.Fatal(err)
	}
	if tex != nil {
		rgba = image.NewRGBA(image.Rect(0, 0, texWidth, texWidth))
		// convert into RGBA format
		for x := 0; x < texWidth; x++ {
			for y := 0; y < texWidth; y++ {
				clr := tex.At(x, y).(color.RGBA)
				rgba.SetRGBA(x, y, clr)
			}
		}
	}

	return rgba
}

func getTextureFromFile(texFile string) *ebiten.Image {
	resourcePath := filepath.Join("game", "resources", "textures", texFile)
	eImg, _, err := ebitenutil.NewImageFromFile(resourcePath)
	if err != nil {
		log.Fatal(err)
	}
	return eImg
}

func getSpriteFromFile(sFile string) *ebiten.Image {
	resourcePath := filepath.Join("game", "resources", "sprites", sFile)
	eImg, _, err := ebitenutil.NewImageFromFile(resourcePath)
	if err != nil {
		log.Fatal(err)
	}
	return eImg
}

// loadContent loads all map texture and static sprite resources
func (g *Game) loadContent() {
	g.projectiles = make(map[*model.Projectile]struct{}, 1024)
	g.effects = make(map[*model.Effect]struct{}, 1024)
	g.sprites = make(map[*model.Sprite]struct{}, 512)
	g.mechSprites = make(map[*model.MechSprite]struct{}, 128)

	// keep a map of textures by name to only load duplicate entries once
	g.tex.texMap = make(map[string]*ebiten.Image, 128)

	// load textured flooring
	if floor, ok := g.mapObj.Textures["floor"]; ok {
		g.tex.floorTex = getRGBAFromFile(floor.Image)
	}

	// load textures mapped by path
	for _, tex := range g.mapObj.Textures {
		if tex.Image != "" {
			if _, ok := g.tex.texMap[tex.Image]; !ok {
				g.tex.texMap[tex.Image] = getTextureFromFile(tex.Image)
			}
		}

		if tex.SideX != "" {
			if _, ok := g.tex.texMap[tex.SideX]; !ok {
				g.tex.texMap[tex.SideX] = getTextureFromFile(tex.SideX)
			}
		}

		if tex.SideY != "" {
			if _, ok := g.tex.texMap[tex.SideY]; !ok {
				g.tex.texMap[tex.SideY] = getTextureFromFile(tex.SideY)
			}
		}
	}

	// load static sprites
	for _, s := range g.mapObj.Sprites {
		if len(s.Image) == 0 {
			continue
		}

		var spriteImg *ebiten.Image
		if eImg, ok := g.tex.texMap[s.Image]; ok {
			spriteImg = eImg
		} else {
			spriteImg = getSpriteFromFile(s.Image)
			g.tex.texMap[s.Image] = spriteImg
		}

		sprite := model.NewSprite(
			s.Position[0], s.Position[1], 1.0, spriteImg, color.RGBA{0, 255, 0, 196}, 0, 0,
		)
		g.addSprite(sprite)
	}
}

// loadSprites loads all mission sprite reources
func (g *Game) loadSprites() {

	// TODO: load mission sprites from yaml file

	mechImg := getSpriteFromFile("mechs/timberwolf.png")
	mechTemplate := model.NewMechSprite(0, 0, mechImg, 0.01)

	for i := 1.5; i <= 19.5; i++ {
		for j := 16.0; j < 24; j++ {
			mech := model.NewMechSpriteFromMech(i, j, mechTemplate)
			g.addMechSprite(mech)
		}
	}
}

func (g *Game) addSprite(sprite *model.Sprite) {
	g.sprites[sprite] = struct{}{}
}

func (g *Game) addMechSprite(mech *model.MechSprite) {
	g.mechSprites[mech] = struct{}{}
}

// func (g *Game) deleteSprite(sprite *model.Sprite) {
// 	delete(g.sprites, sprite)
// }

func (g *Game) addProjectile(projectile *model.Projectile) {
	g.projectiles[projectile] = struct{}{}
}

func (g *Game) deleteProjectile(projectile *model.Projectile) {
	delete(g.projectiles, projectile)
}

func (g *Game) addEffect(effect *model.Effect) {
	g.effects[effect] = struct{}{}
}

func (g *Game) deleteEffect(effect *model.Effect) {
	delete(g.effects, effect)
}
