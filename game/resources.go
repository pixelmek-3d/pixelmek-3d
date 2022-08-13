package game

import (
	"fmt"
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
	if g.mission.Map().Flooring.Default != "" {
		g.tex.floorTexDefault = newFloorTexture(g.mission.Map().Flooring.Default)
	}

	// keep track of floor texture positions by name so they can be matched on later
	var floorTexNames [][]string

	// load texture floor pathing
	if len(g.mission.Map().Flooring.Pathing) > 0 {
		g.tex.floorTexMap = make([][]*FloorTexture, g.mapWidth)
		floorTexNames = make([][]string, g.mapWidth)
		for x := 0; x < g.mapWidth; x++ {
			g.tex.floorTexMap[x] = make([]*FloorTexture, g.mapHeight)
			floorTexNames[x] = make([]string, g.mapHeight)
		}
		// create map grid of path image textures for the X/Y coords indicated
		for _, pathing := range g.mission.Map().Flooring.Pathing {
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
	if len(g.mission.Map().Clutter) > 0 {
		g.clutter = NewClutterHandler()

		for _, clutter := range g.mission.Map().Clutter {
			var clutterImg *ebiten.Image
			if _, ok := g.tex.texMap[clutter.Image]; !ok {
				clutterImg = getSpriteFromFile(clutter.Image)
				g.tex.texMap[clutter.Image] = clutterImg
			}
		}
	}

	// load textures mapped by path
	for _, tex := range g.mission.Map().Textures {
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
	for _, s := range g.mission.Map().Sprites {
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

	// load non-static mission sprites
	g.loadMissionSprites()

	// load all other game sprites
	g.loadGameSprites()
}

// loadMissionSprites loads all mission sprite reources
func (g *Game) loadMissionSprites() {
	// TODO: move these to predefined mech sprites from their own data source files
	mechSpriteTemplates := make(map[string]*model.MechSprite, len(g.mission.Mechs))

	for _, missionMech := range g.mission.Mechs {
		if _, ok := mechSpriteTemplates[missionMech.Image]; !ok {
			mechRelPath := fmt.Sprintf("mechs/%s", missionMech.Image)
			mechImg := getSpriteFromFile(mechRelPath)
			mechSpriteTemplates[missionMech.Image] = model.NewMechSprite(0, 0, missionMech.Scale, mechImg, 0.3)
		}

		mechTemplate := mechSpriteTemplates[missionMech.Image]
		posX, posY := missionMech.Position[0], missionMech.Position[1]
		mech := model.NewMechSpriteFromMech(posX, posY, mechTemplate)

		// TODO: give mission mechs a bit more of a brain
		if len(missionMech.PatrolPath) > 0 {
			mech.PatrolPath = missionMech.PatrolPath
			mech.SetMechAnimation(model.ANIMATE_STRUT)
			mech.AnimationRate = 3
		} else {
			mech.SetMechAnimation(model.ANIMATE_IDLE)
			mech.AnimationRate = 7
		}

		g.sprites.addMechSprite(mech)
	}
}

// loadGameSprites loads all other game sprite reources
func (g *Game) loadGameSprites() {
	// TODO: move these to predefined projectile sprites from their own data source files
	redLaserImg := getSpriteFromFile("projectiles/beams_red.png")
	lifespanSeconds := 4.0 * float64(ebiten.MaxTPS()) // TODO: determine based on max distance for travel
	redLaserProjectile := model.NewAnimatedProjectile(
		0, 0, 0.2, lifespanSeconds, redLaserImg, color.RGBA{}, 1, 3, 4, raycaster.AnchorCenter, 0.01,
	)

	// give projectile angle facing textures by row index
	var redLaserFacingMap = map[float64]int{
		geom.Radians(0):   0,
		geom.Radians(30):  1,
		geom.Radians(90):  2,
		geom.Radians(150): 1,
		geom.Radians(180): 0,
		geom.Radians(210): 1,
		geom.Radians(270): 2,
		geom.Radians(330): 1,
	}
	redLaserProjectile.SetTextureFacingMap(redLaserFacingMap)

	// give projectile impact effect
	laserImpactImg := getSpriteFromFile("effects/laser_impact_sheet.png")
	redExplosionEffect := model.NewAnimatedEffect(
		0, 0, 0.1, laserImpactImg, 8, 3, 1, raycaster.AnchorCenter, 1,
	)
	redLaserProjectile.ImpactEffect = *redExplosionEffect

	g.player.TestProjectile = redLaserProjectile
}
