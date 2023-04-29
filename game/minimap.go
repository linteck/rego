package game

import (
	"image"
	"image/color"
	"sort"

	"lintech/rego/iregoter"
)

func (g *Core) miniMap() *image.RGBA {
	m := image.NewRGBA(image.Rect(0, 0, g.mapWidth, g.mapHeight))

	// wall/world positions
	worldMap := g.mapObj.Level(0)
	for x, row := range worldMap {
		for y := range row {
			c := g.getMapColor(x, y)
			if c.A == 255 {
				c.A = 142
			}
			m.Set(x, y, c)
		}
	}

	// sprite positions, sort by color to avoid random color getting chosen as last when using map keys
	sl := g.rgs[iregoter.RegoterEnumSprite]
	sprites := make([]*iregoter.Entity, 0, sl.Len())
	sl.ForEach(func(_ iregoter.ID, s regoterInCore) {
		sprites = append(sprites, s.entity)
	})
	sort.Slice(sprites, func(i, j int) bool {
		iComp := (sprites[i].MapColor.R + sprites[i].MapColor.G + sprites[i].MapColor.B)
		jComp := (sprites[j].MapColor.R + sprites[j].MapColor.G + sprites[j].MapColor.B)
		return iComp < jComp
	})

	for _, sprite := range sprites {
		if sprite.MapColor.A > 0 {

			m.Set(int(sprite.Position.X), int(sprite.Position.Y), sprite.MapColor)
		}
	}

	// projectile positions
	pl := g.rgs[iregoter.RegoterEnumProjectile]
	projectiles := make([]*iregoter.Entity, 0, pl.Len())
	pl.ForEach(func(_ iregoter.ID, p regoterInCore) {
		projectiles = append(projectiles, p.entity)
	})
	sort.Slice(projectiles, func(i, j int) bool {
		iComp := (projectiles[i].MapColor.R + projectiles[i].MapColor.G + projectiles[i].MapColor.B)
		jComp := (projectiles[j].MapColor.R + projectiles[j].MapColor.G + projectiles[j].MapColor.B)
		return iComp < jComp
	})

	for _, projectile := range projectiles {
		if projectile.MapColor.A > 0 {

			m.Set(int(projectile.Position.X), int(projectile.Position.Y), projectile.MapColor)
		}
	}

	// player position
	player := g.rgs[iregoter.RegoterEnumPlayer].Iterate().Value().entity
	m.Set(int(player.Position.X), int(player.Position.Y), player.MapColor)

	return m
}

func (g *Core) getMapColor(x, y int) color.RGBA {
	worldMap := g.mapObj.Level(0)
	switch worldMap[x][y] {
	case 0:
		return color.RGBA{43, 30, 24, 255}
	case 1:
		return color.RGBA{100, 89, 73, 255}
	case 2:
		return color.RGBA{51, 32, 0, 196}
	case 3:
		return color.RGBA{56, 36, 0, 196}
	case 6:
		// ebitengine splash logo color!
		return color.RGBA{219, 86, 32, 255}
	default:
		return color.RGBA{255, 194, 32, 255}
	}
}
