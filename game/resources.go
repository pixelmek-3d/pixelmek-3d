package game

import (
	"image"
	"image/color"
	"log"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/harbdog/pixelmek-3d/game/model"
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
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
	g.clutterSprites = make(map[*model.Sprite]struct{}, 2048)
	g.mechSprites = make(map[*model.MechSprite]struct{}, 128)

	// keep a map of textures by name to only load duplicate entries once
	g.tex.texMap = make(map[string]*ebiten.Image, 128)

	// load textured flooring
	if g.mapObj.Flooring.Default != "" {
		g.tex.floorTexDefault = getRGBAFromFile(g.mapObj.Flooring.Default)
	}

	// keep track of floor texture positions by name so they can be matched on later
	var floorTexNames [][]string

	// load texture floor pathing
	if len(g.mapObj.Flooring.Pathing) > 0 {
		g.tex.floorTexMap = make([][]*image.RGBA, g.mapWidth)
		floorTexNames = make([][]string, g.mapWidth)
		for x := 0; x < g.mapWidth; x++ {
			g.tex.floorTexMap[x] = make([]*image.RGBA, g.mapHeight)
			floorTexNames[x] = make([]string, g.mapHeight)
		}
		// create map grid of path image textures for the X/Y coords indicated
		for _, pathing := range g.mapObj.Flooring.Pathing {
			tex := getRGBAFromFile(pathing.Image)

			// create filled rectangle paths
			for _, rect := range pathing.Rects {
				x0, y0, x1, y1 := rect[0][0], rect[0][1], rect[1][0], rect[1][1]
				for x := x0; x <= x1; x++ {
					for y := y0; y <= y1; y++ {
						g.tex.floorTexMap[x][y] = tex
						floorTexNames[x][y] = pathing.Image
					}
				}
			}

			// create line segment paths
			for _, segments := range pathing.Lines {
				var prevPoint *geom.Vector2
				for _, seg := range segments {
					point := &geom.Vector2{X: float64(seg[0]), Y: float64(seg[1])}

					if prevPoint != nil {
						// fill in path for line segment from previous to next point
						line := geom.Line{X1: prevPoint.X, Y1: prevPoint.Y, X2: point.X, Y2: point.Y}

						// use the angle of the line to then find every coordinate along the line path
						angle := line.Angle()
						dist := geom.Distance(line.X1, line.Y1, line.X2, line.Y2)
						for d := 0.0; d <= dist; d += 0.1 {
							nLine := geom.LineFromAngle(line.X1, line.Y1, angle, d)
							g.tex.floorTexMap[int(nLine.X2)][int(nLine.Y2)] = tex
						}
					}

					prevPoint = point
				}
			}
		}
	}

	// load clutter sprites and randomly distribute as needed
	if len(g.mapObj.Clutter) > 0 { // 0.25 // 0.05 // 0.5  //0.07
		for _, clutter := range g.mapObj.Clutter {
			var clutterImg *ebiten.Image
			if eImg, ok := g.tex.texMap[clutter.Image]; ok {
				clutterImg = eImg
			} else {
				clutterImg = getSpriteFromFile(clutter.Image)
				g.tex.texMap[clutter.Image] = clutterImg
			}

			// FIXME: this will create too many to loop through (500x500 map = 250,000 cells!)
			//        instead may need to have some dynamic feature where a certain number are made
			//        and then just update their position as player moves?
			for x := 1; x < 25; x += 1 {
				for y := 1; y < 25; y += 1 {
					cSprite := model.NewSprite(
						float64(x), float64(y), clutter.Scale, clutterImg, color.RGBA{}, raycaster.AnchorBottom, 0,
					)
					g.addClutterSprite(cSprite)
				}
			}
		}
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

		for _, position := range s.Positions {
			sprite := model.NewSprite(
				position[0], position[1], 1.0, spriteImg, color.RGBA{0, 255, 0, 196}, raycaster.AnchorBottom, 0,
			)
			g.addSprite(sprite)
		}
	}
}

// loadSprites loads all mission sprite reources
func (g *Game) loadSprites() {

	// TODO: load mission sprites from yaml file

	// mechImg := getSpriteFromFile("mechs/timberwolf.png")
	// mechTemplate := model.NewMechSprite(0, 0, mechImg, 0.01)

	// // for i := 1.5; i <= 19.5; i++ {
	// // 	for j := 16.0; j < 24; j++ {
	// // 		mech := model.NewMechSpriteFromMech(i, j, mechTemplate)
	// // 		g.addMechSprite(mech)
	// // 	}
	// // }
	// mech := model.NewMechSpriteFromMech(5, 18, mechTemplate)
	// g.addMechSprite(mech)
	// mech2 := model.NewMechSpriteFromMech(7, 18, mechTemplate)
	// g.addMechSprite(mech2)
}

func (g *Game) addSprite(sprite *model.Sprite) {
	g.sprites[sprite] = struct{}{}
}

func (g *Game) addClutterSprite(clutter *model.Sprite) {
	g.clutterSprites[clutter] = struct{}{}
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
