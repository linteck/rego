package model

import (
	"sort"

	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
)

// checks for valid move from current position, returns valid (x, y) position, whether a collision
// was encountered, and a list of entity collisions that may have been encountered
// NOTE: Bug??
// When collision with Wall, the []*EntityCollision is empty!
func (g *Core) getValidMove(entity *Entity,
	moveX, moveY, moveZ float64, checkAlternate bool) (*geom.Vector2, *EntityCollision) {

	posX, posY, posZ := entity.Position.X, entity.Position.Y, entity.Position.Z
	if posX == moveX && posY == moveY && posZ == moveZ {
		return &geom.Vector2{X: posX, Y: posY}, nil
	}

	newX, newY, newZ := moveX, moveY, moveZ
	moveLine := geom.Line{X1: posX, Y1: posY, X2: newX, Y2: newY}

	collisionEntities := make([]*EntityCollision, 0, 10)

	// check wall collisions
	for _, borderLine := range g.collisionMap {
		// TODO: only check intersection of nearby wall cells instead of all of them
		if px, py, ok := geom.LineIntersection(moveLine, borderLine); ok {
			point := geom.Vector2{X: px, Y: py}
			collisionEntities = append(
				// Collistion with wall, RcTx is nil
				collisionEntities, &EntityCollision{
					position: Position{X: point.X, Y: point.Y, Z: newZ},
					peer:     WALL_ID,
					distance: geom.Distance2(posX, posY, point.X, point.Y),
				},
			)
		}
	}

	// check sprite against player collision
	playerInCore := g.getPlayer()
	player := playerInCore.sprite
	if nil != player && entity.RgId != player.Entity.RgId &&
		entity.ParentId != player.Entity.RgId && entity.CollisionRadius > 0 {
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
				intersectPoints := geom.LineCircleIntersection(chkLine, playerCircle, true)
				for _, point := range intersectPoints {
					collisionEntities = append(
						//collisionEntities, &EntityCollision{entity: player.Entity, collision: &intersect, collisionZ: zIntersect},
						collisionEntities, &EntityCollision{
							position: Position{X: point.X, Y: point.Y, Z: zIntersect},
							peer:     playerInCore.entity.RgId,
						},
					)
				}
			}
		}
	}

	// check sprite collisions
	g.rgs[RegoterEnumSprite].ForEach(
		func(i ID, r *regoterInCore) {
			sprite := r.sprite
			// TODO: only check intersection of nearby sprites instead of all of them
			if entity.RgId == sprite.Entity.RgId || entity.ParentId == sprite.Entity.RgId || entity.CollisionRadius <= 0 || sprite.Entity.CollisionRadius <= 0 {
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
					intersectPoints := geom.LineCircleIntersection(chkLine, spriteCircle, true)

					for _, point := range intersectPoints {
						collisionEntities = append(
							//collisionEntities, &EntityCollision{entity: sprite.Entity, collision: &intersect, collisionZ: zIntersect},
							collisionEntities, &EntityCollision{
								position: Position{X: point.X, Y: point.Y, Z: zIntersect},
								peer:     r.entity.RgId,
							},
						)
					}
				}
			}
		})

	isCollision := len(collisionEntities) > 0
	if isCollision {
		// sort collisions by distance to current entity position
		sort.Slice(collisionEntities, func(i, j int) bool {
			return collisionEntities[i].distance < collisionEntities[j].distance
		})
		// If there is collistion, don't move!
		return &geom.Vector2{X: posX, Y: posY}, collisionEntities[0]
	} else {
		return &geom.Vector2{X: newX, Y: newY}, nil
	}

}

// zEntityIntersection returns the best Position.Z intersection point on the target from the source (-1 if no intersection)
func zEntityIntersection(sourceZ float64, source, target *Entity) float64 {
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
func zEntityMinMax(Z float64, entity *Entity) (float64, float64) {
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
