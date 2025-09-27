package game

import (
	"image"
	"math"
	"sort"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/harbdog/raycaster-go/geom3d"
	"github.com/pixelmek-3d/pixelmek-3d/game/model"
	"github.com/pixelmek-3d/pixelmek-3d/game/render/sprites"
)

type EntityCollision struct {
	entity     model.Entity
	collision  *geom.Vector2
	collisionZ float64
}

func init() {
	collisonSpriteTypes = map[sprites.SpriteType]bool{
		sprites.MapSpriteType:         true,
		sprites.MechSpriteType:        true,
		sprites.VehicleSpriteType:     true,
		sprites.VTOLSpriteType:        true,
		sprites.InfantrySpriteType:    true,
		sprites.EmplacementSpriteType: true,
	}
}

func (g *Game) isCollisionType(spriteType sprites.SpriteType) bool {
	if _, containsType := collisonSpriteTypes[spriteType]; containsType {
		return true
	}
	return false
}

// checks walls for line of sight from source to target entity
func (g *Game) lineOfSight(source, target model.Entity) bool {
	if source == nil || target == nil {
		return false
	}

	srcX, srcY := source.Pos().X, source.Pos().Y
	tgtX, tgtY := target.Pos().X, target.Pos().Y

	line := geom.Line{X1: srcX, Y1: srcY, X2: tgtX, Y2: tgtY}
	lRect := image.Rect(int(srcX), int(srcY), int(tgtX), int(tgtY))
	for _, borderLine := range g.collisionMap {
		// quick check for nearby wall cells instead of all of them
		bRect := image.Rect(int(borderLine.X1), int(borderLine.Y1), int(borderLine.X2), int(borderLine.Y2))
		if (bRect.Min.X > lRect.Max.X && bRect.Min.Y > lRect.Max.Y) ||
			(bRect.Max.X < lRect.Min.X && bRect.Max.Y < lRect.Min.Y) {
			continue
		}

		_, _, intersects := geom.LineIntersection(line, *borderLine)
		if intersects {
			return false
		}
	}

	return true
}

// checks for valid move from current position, returns valid (x, y) position, whether a collision
// was encountered, and a list of entity collisions that may have been encountered
func (g *Game) getValidMove(entity model.Entity, moveX, moveY, moveZ float64, checkAlternate bool) (*geom.Vector2, float64, bool, []*EntityCollision) {
	position := entity.Pos()
	posX, posY, posZ := position.X, position.Y, entity.PosZ()

	entityCollisionRadius := entity.CollisionRadius()
	entityCollisionHeight := entity.CollisionHeight()
	if entityCollisionRadius <= 0 || entityCollisionHeight <= 0 {
		return &geom.Vector2{X: posX, Y: posY}, posZ, false, []*EntityCollision{}
	}
	if posX == moveX && posY == moveY && posZ == moveZ {
		return &geom.Vector2{X: posX, Y: posY}, posZ, false, []*EntityCollision{}
	}

	newX, newY, newZ := moveX, moveY, moveZ
	moveLine := geom.Line{X1: posX, Y1: posY, X2: newX, Y2: newY}

	// use line distance to determine fast approximation of which x/y positions to skip collision checks against
	moveLine3d := geom3d.Line3d{X1: posX, Y1: posY, Z1: posZ, X2: newX, Y2: newY, Z2: newZ}
	checkDist := moveLine3d.Distance()

	intersectPoints := []geom.Vector2{}
	collisionEntities := []*EntityCollision{}

	// check wall collisions
	for _, borderLine := range g.collisionMap {
		if px, py, ok := geom.LineIntersection(moveLine, *borderLine); ok {
			intersectPoints = append(intersectPoints, geom.Vector2{X: px, Y: py})
		}
	}

	// check sprite against player collision
	if entity != g.player.Unit && entity.Parent() != g.player.Unit && !entity.IsDestroyed() {
		// only check for collision if player is somewhat nearby
		playerPosition := g.player.Pos()
		playerCollisionRadius := g.player.CollisionRadius()
		if model.PointInProximity(checkDist, newX, newY, playerPosition.X, playerPosition.Y) {
			// quick check if intersects in Z-plane
			zIntersect := zEntityIntersection(newZ, entity, g.player.Unit)

			// check if movement line intersects with combined collision radii
			combinedCircle := geom.Circle{X: playerPosition.X, Y: playerPosition.Y, Radius: playerCollisionRadius + entityCollisionRadius}
			combinedIntersects := geom.LineCircleIntersection(moveLine, combinedCircle, true)

			if zIntersect >= 0 && len(combinedIntersects) > 0 {
				playerCircle := geom.Circle{X: playerPosition.X, Y: playerPosition.Y, Radius: playerCollisionRadius}
				for _, chkPoint := range combinedIntersects {
					// intersections from combined circle radius indicate center point to check intersection toward sprite collision circle
					chkLine := geom.Line{X1: chkPoint.X, Y1: chkPoint.Y, X2: playerPosition.X, Y2: playerPosition.Y}
					intersectPoints = append(intersectPoints, geom.LineCircleIntersection(chkLine, playerCircle, true)...)

					for _, intersect := range intersectPoints {
						collisionEntities = append(
							collisionEntities, &EntityCollision{entity: g.player.Unit, collision: &intersect, collisionZ: zIntersect},
						)
					}
				}
			}
		}
	}

	// check sprite collisions
	for _, spriteType := range g.sprites.SpriteTypes() {
		if !g.isCollisionType(spriteType) {
			// only check collision against certain sprite types (skip projectiles, effects, etc.)
			continue
		}
		g.sprites.RangeByType(spriteType, func(k, _ any) bool {
			spriteInterface := k.(raycaster.Sprite)
			sEntity := getEntityFromInterface(spriteInterface)
			if entity == sEntity || entity.Parent() == sEntity || sEntity.CollisionRadius() <= 0 || sEntity.IsDestroyed() {
				return true
			}

			sEntityPosition := sEntity.Pos()
			sEntityCr := sEntity.CollisionRadius()

			// only check intersection of nearby sprites instead of all of them
			if !model.PointInProximity(checkDist, newX, newY, sEntityPosition.X, sEntityPosition.Y) {
				return true
			}

			// quick check if intersects in Z-plane
			zIntersect := zEntityIntersection(newZ, entity, sEntity)
			if zIntersect < 0 {
				return true
			}

			// check if movement line intersects with combined collision radii
			combinedCircle := geom.Circle{X: sEntityPosition.X, Y: sEntityPosition.Y, Radius: sEntityCr + entityCollisionRadius}
			combinedIntersects := geom.LineCircleIntersection(moveLine, combinedCircle, true)

			circleHit := false
			if len(combinedIntersects) > 0 {
				spriteCircle := geom.Circle{X: sEntityPosition.X, Y: sEntityPosition.Y, Radius: sEntityCr}
				for _, chkPoint := range combinedIntersects {
					// intersections from combined circle radius indicate center point to check intersection toward sprite collision circle
					chkLine := geom.Line{X1: chkPoint.X, Y1: chkPoint.Y, X2: sEntityPosition.X, Y2: sEntityPosition.Y}
					chkLineIntersects := geom.LineCircleIntersection(chkLine, spriteCircle, true)
					intersectPoints = append(intersectPoints, chkLineIntersects...)

					for _, intersect := range chkLineIntersects {
						circleHit = true
						collisionEntities = append(
							collisionEntities, &EntityCollision{entity: sEntity, collision: &intersect, collisionZ: zIntersect},
						)
					}
				}
			}

			if !circleHit {
				// check if move point could be inside the circle without touching it
				chkLine := geom.Line{X1: newX, Y1: newY, X2: sEntityPosition.X, Y2: sEntityPosition.Y}
				chkLineDist := chkLine.Distance()
				if chkLineDist <= combinedCircle.Radius {
					chkPoint := geom.Vector2{X: newX, Y: newY}
					intersectPoints = append(intersectPoints, chkPoint)
					collisionEntities = append(
						collisionEntities, &EntityCollision{entity: sEntity, collision: &chkPoint, collisionZ: zIntersect},
					)
				}
			}

			return true
		})
	}

	// sort collisions by distance to current entity position
	sort.Slice(collisionEntities, func(i, j int) bool {
		distI := geom.Distance2(posX, posY, collisionEntities[i].collision.X, collisionEntities[i].collision.Y)
		distJ := geom.Distance2(posX, posY, collisionEntities[j].collision.X, collisionEntities[j].collision.Y)
		return distI < distJ
	})

	isCollision := len(intersectPoints) > 0

	if isCollision {
		if checkAlternate {
			zDiff := math.Abs(moveZ - posZ)
			if zDiff > 0 {
				// if some Z movement, try to move only in Z (useful vs. walls)
				zP, zZ, zCollide, _ := g.getValidMove(entity, posX, posY, moveZ, false)
				if !zCollide {
					return zP, zZ, isCollision, collisionEntities
				} else {
					// Z-only resulted in collision, try without any Z (useful when on top of something)
					zP, zZ, _, _ = g.getValidMove(entity, moveX, moveY, posZ, true)
					return zP, zZ, isCollision, collisionEntities
				}
			}

			if len(collisionEntities) > 0 {
				// attempt escaping stuck on an entity: find the closest intersection point and angle away from it
				min := math.Inf(1)
				minI := -1
				for i, p := range intersectPoints {
					d2 := geom.Distance2(posX, posY, p.X, p.Y)
					if d2 < min {
						min = d2
						minI = i
					}
				}

				// use the closest intersecting point to determine a safe distance to make the move
				intersectLine := geom.Line{X1: posX, Y1: posY, X2: intersectPoints[minI].X, Y2: intersectPoints[minI].Y}
				intersectAngle := intersectLine.Angle()

				// generate new move line using calculated angle and safe distance from intersecting point
				var escapeAngle float64
				escapeDist := geom.Clamp(checkDist, 0, entityCollisionRadius-0.001)
				if model.AngleDistance(entity.Heading(), intersectAngle) >= 0 {
					escapeAngle = intersectAngle - geom.HalfPi
				} else {
					escapeAngle = intersectAngle + geom.HalfPi
				}

				escapeLine := geom.LineFromAngle(posX, posY, escapeAngle, escapeDist)
				escapeX, escapeY := escapeLine.X2, escapeLine.Y2
				//log.Debugf("[%0.3f, %0.3f] escape entity collision (%0.3f@%0.2f)", escapeY, escapeY, escapeDist, geom.Degrees(escapeAngle))

				nP, nZ, _, _ := g.getValidMove(entity, escapeX, escapeY, moveZ, false)
				return nP, nZ, isCollision, collisionEntities

			} else {
				// attempt escaping stuck on a wall: check move in most favored direction, then the other, based on heading
				identityLine := geom.LineFromAngle(0.0, 0.0, entity.Heading(), 1.0)
				identityX, identityY := math.Abs(identityLine.X2), math.Abs(identityLine.Y2)
				if identityX >= identityY {
					// try to move only X, then only Y
					// log.Debugf("[%0.3f, %0.3f] first, trying only X (%0.3f)", posX, posY, moveX)
					nP, nZ, xCollide, _ := g.getValidMove(entity, moveX, posY, moveZ, false)
					if !xCollide {
						return nP, nZ, isCollision, collisionEntities
					} else {
						if geom.NearlyEqual(posY, moveY, 0.001) {
							// try small moveY offset to avoid getting stuck on wall collision
							dY := entity.Velocity() / 2
							if moveY-posY == 0 {
								if math.Signbit(moveX) {
									dY *= -1
								}
							} else {
								if math.Signbit(moveY - posY) {
									dY *= -1
								}
							}
							moveY = posY + dY
						}
						// log.Debugf("[%0.3f, %0.3f] second, trying only Y (%0.3f)", posX, posY, moveY)
						nP, nZ, _, _ = g.getValidMove(entity, posX, moveY, moveZ, false)
						return nP, nZ, isCollision, collisionEntities
					}
				} else {
					// try to move only Y, then only X
					// log.Debugf("[%0.3f, %0.3f] first, trying only Y (%0.3f)", posX, posY, moveY)
					nP, nZ, yCollide, _ := g.getValidMove(entity, posX, moveY, moveZ, false)
					if !yCollide {
						return nP, nZ, isCollision, collisionEntities
					} else {
						if geom.NearlyEqual(posX, moveX, 0.001) {
							// try small moveX offset to avoid getting stuck on wall collision
							dX := entity.Velocity() / 2
							if moveX-posX == 0 {
								if math.Signbit(moveY) {
									dX *= -1
								}
							} else {
								if math.Signbit(moveX - posX) {
									dX *= -1
								}
							}
							moveX = posX + dX
						}
						// log.Debugf("[%0.3f, %0.3f] second, trying only X (%0.3f)", posX, posY, moveX)
						nP, nZ, _, _ = g.getValidMove(entity, moveX, posY, moveZ, false)
						return nP, nZ, isCollision, collisionEntities
					}
				}
			}
		} else {
			// looks like it cannot move
			// log.Debugf("[%0.3f, %0.3f] collision: cannot move", posX, posY)
			return &geom.Vector2{X: posX, Y: posY}, posZ, isCollision, collisionEntities
		}
	}

	// prevent index out of bounds errors
	ix := int(newX)
	iy := int(newY)

	switch {
	case ix < 0 || newX < 0:
		newX = clipDistance
		ix = 0
	case ix >= g.mapWidth:
		newX = float64(g.mapWidth) - clipDistance
		ix = int(newX)
	}

	switch {
	case iy < 0 || newY < 0:
		newY = clipDistance
		iy = 0
	case iy >= g.mapHeight:
		newY = float64(g.mapHeight) - clipDistance
		iy = int(newY)
	}

	worldMap := g.mission.Map().Level(0)
	if worldMap[ix][iy] <= 0 {
		posX = newX
		posY = newY
	} else {
		isCollision = true
	}

	// prevent going under the floor
	// TODO: prevent going above flight ceiling (set in map yaml?)
	posZ = newZ
	zMin, _ := zEntityMinMax(0, entity)
	zMin = math.Abs(zMin)
	if posZ < zMin {
		posZ = zMin
		isCollision = true
	}

	return &geom.Vector2{X: posX, Y: posY}, posZ, isCollision, collisionEntities
}

// zEntityIntersection returns the best positionZ intersection point on the target from the source (-1 if no intersection)
func zEntityIntersection(sourceZ float64, source, target model.Entity) float64 {
	srcMinZ, srcMaxZ := zEntityMinMax(sourceZ, source)
	tgtMinZ, tgtMaxZ := zEntityMinMax(target.PosZ(), target)

	var intersectZ float64 = -1
	if srcMinZ > tgtMaxZ || tgtMinZ > srcMaxZ {
		// no intersection
		return intersectZ
	}

	// find best simple intersection within the target range
	midZ := srcMinZ + (srcMaxZ-srcMinZ)/2
	intersectZ = geom.Clamp(midZ, tgtMinZ, tgtMaxZ)

	if srcMinZ > 0.1 && tgtMinZ > 0.1 {
		return intersectZ
	}

	return intersectZ
}

// zEntityMinMax calculates the minZ/maxZ used for basic collision checking in the Z-plane
func zEntityMinMax(positionZ float64, entity model.Entity) (float64, float64) {
	var minZ, maxZ float64
	collisionHeight := entity.CollisionHeight()

	switch entity.Anchor() {
	case raycaster.AnchorBottom:
		minZ, maxZ = positionZ, positionZ+collisionHeight
	case raycaster.AnchorCenter:
		minZ, maxZ = positionZ-collisionHeight/2, positionZ+collisionHeight/2
	case raycaster.AnchorTop:
		minZ, maxZ = positionZ-collisionHeight, positionZ
	}

	return minZ, maxZ
}
