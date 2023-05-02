package game

import (
	"math"
	"sort"

	"lintech/rego/game/loader"
	"lintech/rego/iregoter"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
)

type EntityCollision struct {
	entity     *iregoter.Entity
	collision  *geom.Vector2
	collisionZ float64
}

// checks for valid move from current position, returns valid (x, y) position, whether a collision
// was encountered, and a list of entity collisions that may have been encountered
func (g *Core) getValidMove(entity *iregoter.Entity,
	moveX, moveY, moveZ float64, checkAlternate bool) (*geom.Vector2, bool, []*EntityCollision) {

	posX, posY, posZ := entity.Position.X, entity.Position.Y, entity.Position.Z
	if posX == moveX && posY == moveY && posZ == moveZ {
		return &geom.Vector2{X: posX, Y: posY}, false, []*EntityCollision{}
	}

	newX, newY, newZ := moveX, moveY, moveZ
	moveLine := geom.Line{X1: posX, Y1: posY, X2: newX, Y2: newY}

	intersectPoints := []geom.Vector2{}
	collisionEntities := []*EntityCollision{}

	// check wall collisions
	for _, borderLine := range g.collisionMap {
		// TODO: only check intersection of nearby wall cells instead of all of them
		if px, py, ok := geom.LineIntersection(moveLine, borderLine); ok {
			intersectPoints = append(intersectPoints, geom.Vector2{X: px, Y: py})
		}
	}

	// check sprite against player collision
	player := g.rgs[iregoter.RegoterEnumPlayer].Iterate().Value().sprite
	if entity.RgId != player.Entity.RgId && entity.Collidable && entity.CollisionRadius > 0 {
		// TODO: only check for collision if player is somewhat nearby

		// quick check if intersects in Z-plane
		zIntersect := zEntityIntersection(newZ, entity, player.Entity)

		// check if movement line intersects with combined collision radii
		combinedCircle := geom.Circle{X: player.Entity.Position.X, Y: player.Entity.Position.Y,
			Radius: player.Entity.CollisionRadius + entity.CollisionRadius}
		combinedIntersects := geom.LineCircleIntersection(moveLine, combinedCircle, true)

		if zIntersect >= 0 && len(combinedIntersects) > 0 {
			playerCircle := geom.Circle{X: player.Entity.Position.X, Y: player.Entity.Position.Y, Radius: player.Entity.CollisionRadius}
			for _, chkPoint := range combinedIntersects {
				// intersections from combined circle radius indicate center point to check intersection toward sprite collision circle
				chkLine := geom.Line{X1: chkPoint.X, Y1: chkPoint.Y, X2: player.Entity.Position.X, Y2: player.Entity.Position.Y}
				intersectPoints = append(intersectPoints, geom.LineCircleIntersection(chkLine, playerCircle, true)...)

				for _, intersect := range intersectPoints {
					collisionEntities = append(
						collisionEntities, &EntityCollision{entity: player.Entity, collision: &intersect, collisionZ: zIntersect},
					)
				}
			}
		}
	}

	// check sprite collisions
	g.rgs[iregoter.RegoterEnumSprite].ForEach(
		func(i iregoter.ID, r regoterInCore) {
			sprite := r.sprite
			// TODO: only check intersection of nearby sprites instead of all of them
			if entity.RgId == sprite.Entity.RgId || entity.Collidable || entity.CollisionRadius <= 0 || sprite.Entity.CollisionRadius <= 0 {
				return
			}

			// quick check if intersects in Z-plane
			zIntersect := zEntityIntersection(newZ, entity, sprite.Entity)

			// check if movement line intersects with combined collision radii
			combinedCircle := geom.Circle{X: sprite.Entity.Position.X, Y: sprite.Entity.Position.Y,
				Radius: sprite.Entity.CollisionRadius + entity.CollisionRadius}
			combinedIntersects := geom.LineCircleIntersection(moveLine, combinedCircle, true)

			if zIntersect >= 0 && len(combinedIntersects) > 0 {
				spriteCircle := geom.Circle{X: sprite.Entity.Position.X, Y: sprite.Entity.Position.Y, Radius: sprite.Entity.CollisionRadius}
				for _, chkPoint := range combinedIntersects {
					// intersections from combined circle radius indicate center point to check intersection toward sprite collision circle
					chkLine := geom.Line{X1: chkPoint.X, Y1: chkPoint.Y, X2: sprite.Entity.Position.X, Y2: sprite.Entity.Position.Y}
					intersectPoints = append(intersectPoints, geom.LineCircleIntersection(chkLine, spriteCircle, true)...)

					for _, intersect := range intersectPoints {
						collisionEntities = append(
							collisionEntities, &EntityCollision{entity: sprite.Entity, collision: &intersect, collisionZ: zIntersect},
						)
					}
				}
			}
		})

	// sort collisions by distance to current entity position
	sort.Slice(collisionEntities, func(i, j int) bool {
		distI := geom.Distance2(posX, posY, collisionEntities[i].collision.X, collisionEntities[i].collision.Y)
		distJ := geom.Distance2(posX, posY, collisionEntities[j].collision.X, collisionEntities[j].collision.Y)
		return distI < distJ
	})

	isCollision := len(intersectPoints) > 0

	if isCollision {
		if checkAlternate {
			// find the point closest to the start position
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
			moveLine = geom.Line{X1: posX, Y1: posY, X2: intersectPoints[minI].X, Y2: intersectPoints[minI].Y}
			dist := math.Sqrt(min)
			angle := moveLine.Angle()

			// generate new move line using calculated angle and safe distance from intersecting point
			moveLine = geom.LineFromAngle(posX, posY, angle, dist-0.01)

			newX, newY = moveLine.X2, moveLine.Y2

			// if either X or Y direction was already intersecting, attempt move only in the adjacent direction
			xDiff := math.Abs(newX - posX)
			yDiff := math.Abs(newY - posY)
			// Bug? Shoulde be <0.001 ?
			if xDiff <= 0.001 || yDiff <= 0.001 {
				switch {
				case xDiff <= 0.001:
					// no more room to move in X, try to move only Y
					// fmt.Printf("\t[@%v,%v] move to (%v,%v) try adjacent move to {%v,%v}\n",
					// 	c.pos.X, c.pos.Y, moveX, moveY, posX, moveY)
					return g.getValidMove(entity, posX, moveY, posZ, false)
				case yDiff <= 0.001:
					// no more room to move in Y, try to move only X
					// fmt.Printf("\t[@%v,%v] move to (%v,%v) try adjacent move to {%v,%v}\n",
					// 	c.pos.X, c.pos.Y, moveX, moveY, moveX, posY)
					return g.getValidMove(entity, moveX, posY, posZ, false)
				default:
					// try the new position
					// TODO: need some way to try a potentially valid shorter move without checkAlternate while also avoiding infinite loop
					return g.getValidMove(entity, newX, newY, posZ, false)
				}
			} else {
				// looks like it cannot move
				return &geom.Vector2{X: posX, Y: posY}, isCollision, collisionEntities
			}
		} else {
			// looks like it cannot move
			return &geom.Vector2{X: posX, Y: posY}, isCollision, collisionEntities
		}
	}

	// prevent index out of bounds errors
	ix := int(newX)
	iy := int(newY)

	switch {
	case ix < 0 || newX < 0:
		newX = loader.ClipDistance
		ix = 0
	case ix >= g.mapWidth:
		newX = float64(g.mapWidth) - loader.ClipDistance
		ix = int(newX)
	}

	switch {
	case iy < 0 || newY < 0:
		newY = loader.ClipDistance
		iy = 0
	case iy >= g.mapHeight:
		newY = float64(g.mapHeight) - loader.ClipDistance
		iy = int(newY)
	}

	worldMap := g.mapObj.Level(0)
	if worldMap[ix][iy] <= 0 {
		posX = newX
		posY = newY
	} else {
		isCollision = true
	}

	return &geom.Vector2{X: posX, Y: posY}, isCollision, collisionEntities
}

// zEntityIntersection returns the best Position.Z intersection point on the target from the source (-1 if no intersection)
func zEntityIntersection(sourceZ float64, source, target *iregoter.Entity) float64 {
	srcMinZ, srcMaxZ := zEntityMinMax(sourceZ, source)
	tgtMinZ, tgtMaxZ := zEntityMinMax(target.Position.Z, target)

	var intersectZ float64 = -1
	if srcMinZ > tgtMaxZ || tgtMinZ > srcMaxZ {
		// no intersection
		return intersectZ
	}

	// find best simple intersection within the target range
	midZ := srcMinZ + (srcMaxZ-srcMinZ)/2
	intersectZ = geom.Clamp(midZ, tgtMinZ, tgtMaxZ)

	return intersectZ
}

// zEntityMinMax calculates the minZ/maxZ used for basic collision checking in the Z-plane
func zEntityMinMax(Z float64, entity *iregoter.Entity) (float64, float64) {
	var minZ, maxZ float64
	collisionHeight := entity.CollisionHeight

	switch entity.Anchor {
	case raycaster.AnchorBottom:
		minZ, maxZ = Z, Z+collisionHeight
	case raycaster.AnchorCenter:
		minZ, maxZ = Z-collisionHeight/2, Z+collisionHeight/2
	case raycaster.AnchorTop:
		minZ, maxZ = Z-collisionHeight, Z
	}

	return minZ, maxZ
}
