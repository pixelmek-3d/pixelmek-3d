package game

import (
	"image"
	"image/color"
	"log"
	"path/filepath"
	"sort"
	"strconv"

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
	g.sprites = make(map[*model.Sprite]struct{}, 128)

	// load textured flooring
	if floor, ok := g.mapObj.Textures["floor"]; ok {
		g.tex.floorTex = getRGBAFromFile(floor.Image)
	}

	// find highest texture index to determine texture slice length
	textureIndices := make([]int, len(g.mapObj.Textures))
	for k := range g.mapObj.Textures {
		if texIndex, err := strconv.Atoi(k); err == nil {
			textureIndices = append(textureIndices, texIndex)
		}
	}

	sort.Ints(textureIndices)
	highestIndex := textureIndices[len(textureIndices)-1]
	g.tex.textures = make([]*ebiten.Image, highestIndex+1)

	for _, i := range textureIndices {
		k := strconv.Itoa(i)
		kTex := g.mapObj.Textures[k]
		if len(kTex.Image) == 0 {
			continue
		}

		g.tex.textures[i] = getTextureFromFile(kTex.Image)
	}

	// load static sprites
	spriteMap := make(map[string]*ebiten.Image, len(g.mapObj.Sprites))
	for _, s := range g.mapObj.Sprites {
		if len(s.Image) == 0 {
			continue
		}

		var spriteImg *ebiten.Image
		if eImg, ok := spriteMap[s.Image]; ok {
			spriteImg = eImg
		} else {
			spriteImg = getSpriteFromFile(s.Image)
			spriteMap[s.Image] = spriteImg
		}

		sprite := model.NewSprite(
			s.Position[0], s.Position[1], 1.0, spriteImg, color.RGBA{0, 255, 0, 196}, texWidth, 0, 0,
		)
		g.addSprite(sprite)
	}
}

// loadSprites loads all mission sprite reources
func (g *Game) loadSprites() {
	// TODO: load mission sprites
}

func (g *Game) addSprite(sprite *model.Sprite) {
	g.sprites[sprite] = struct{}{}
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
