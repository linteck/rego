package model

import (
	"image"
	"image/color"
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

	typesNeedDraw := []RegoterEnum{
		RegoterEnumSprite,
		//RegoterEnumProjectile,
		//RegoterEnumEffect,
	}
	// draw sprite screen indicators to show we know where it was raycasted (must occur after camera.Update)
	for _, t := range typesNeedDraw {
		sl := g.rgs[t]
		for _, val := range sl {
			if val.sprite != nil {
				m.Set(int(val.entity.Position.X), int(val.entity.Position.Y), val.entity.MapColor)
			}
		}
	}

	// player position
	pl := g.rgs[RegoterEnumPlayer]
	for _, p := range pl {
		m.Set(int(p.entity.Position.X), int(p.entity.Position.Y), p.entity.MapColor)
	}

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
