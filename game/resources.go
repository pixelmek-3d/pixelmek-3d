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
	g.sprites = NewSpriteHandler()

	// keep a map of textures by name to only load duplicate entries once
	g.tex.texMap = make(map[string]*ebiten.Image, 128)

	// load textured flooring
	if g.mapObj.Flooring.Default != "" {
		g.tex.floorTexDefault = newFloorTexture(g.mapObj.Flooring.Default)
	}

	// keep track of floor texture positions by name so they can be matched on later
	var floorTexNames [][]string

	// load texture floor pathing
	if len(g.mapObj.Flooring.Pathing) > 0 {
		g.tex.floorTexMap = make([][]*FloorTexture, g.mapWidth)
		floorTexNames = make([][]string, g.mapWidth)
		for x := 0; x < g.mapWidth; x++ {
			g.tex.floorTexMap[x] = make([]*FloorTexture, g.mapHeight)
			floorTexNames[x] = make([]string, g.mapHeight)
		}
		// create map grid of path image textures for the X/Y coords indicated
		for _, pathing := range g.mapObj.Flooring.Pathing {
			tex := newFloorTexture(pathing.Image)

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

	// load clutter sprites mapped by path
	if len(g.mapObj.Clutter) > 0 {
		g.clutter = NewClutterHandler()

		for _, clutter := range g.mapObj.Clutter {
			var clutterImg *ebiten.Image
			if _, ok := g.tex.texMap[clutter.Image]; !ok {
				clutterImg = getSpriteFromFile(clutter.Image)
				g.tex.texMap[clutter.Image] = clutterImg
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

		if s.Scale == 0.0 {
			// default unset scale to 1.0
			s.Scale = 1.0
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
				position[0], position[1], s.Scale, spriteImg, color.RGBA{0, 255, 0, 196}, raycaster.AnchorBottom, 0,
			)
			g.sprites.addSprite(sprite)
		}
	}
}

// loadSprites loads all mission sprite reources
func (g *Game) loadSprites() {

	// TODO: load mission sprites from yaml file

	tbrImg := getSpriteFromFile("mechs/timberwolf.png")
	tbrTemplate := model.NewMechSprite(0, 0, 0.75, tbrImg, 0.01)

	whkImg := getSpriteFromFile("mechs/warhawk.png")
	whkTemplate := model.NewMechSprite(0, 0, 0.8, whkImg, 0.01)

	// testing a few  of them
	mech0 := model.NewMechSpriteFromMech(13, 15, whkTemplate)
	mech0.SetMechAnimation(model.ANIMATE_STATIC)
	g.sprites.addMechSprite(mech0)

	mech1 := model.NewMechSpriteFromMech(15, 15, whkTemplate)
	mech1.SetMechAnimation(model.ANIMATE_STRUT)
	mech1.AnimationRate = 5
	mech1.Velocity = 0.0025
	g.sprites.addMechSprite(mech1)

	mech2 := model.NewMechSpriteFromMech(17, 15, whkTemplate)
	mech2.SetMechAnimation(model.ANIMATE_IDLE)
	mech2.AnimationRate = 7
	g.sprites.addMechSprite(mech2)

	// testing lots of them
	for i := 1.5; i <= 19.5; i++ {
		for j := 22.0; j < 24; j++ {
			mech := model.NewMechSpriteFromMech(i, j, tbrTemplate)

			mech.Velocity = 0.01

			if false && int(j)%2 == 0 {
				mech.SetMechAnimation(model.ANIMATE_IDLE)
				mech.AnimationRate = 7
			} else {
				mech.SetMechAnimation(model.ANIMATE_STRUT)
				// TODO: set AnimationRate based on mech velocity (1 is fastest for running light mechs)
				//       2 could be for medium mech at run speed, 3 for heavy, 4 for assault,
				//       higher values if mech is moving but not at run speed.
				mech.AnimationRate = 2
			}

			if int(i)%2 == 0 {
				mech.SetAnimationReversed(true)
			}

			if mech.NumAnimationFrames() > 1 {
				mech.SetAnimationFrame(int(i) % mech.NumAnimationFrames())
			}

			g.sprites.addMechSprite(mech)
		}
	}
}
