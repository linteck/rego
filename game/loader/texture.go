package loader

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type TextureHandler struct {
	mapObj         *Map
	Textures       []*ebiten.Image
	FloorTex       *image.RGBA
	RenderFloorTex bool
}

func NewTextureHandler(mapObj *Map, textureCapacity int) *TextureHandler {
	t := &TextureHandler{
		mapObj:         mapObj,
		Textures:       make([]*ebiten.Image, textureCapacity),
		RenderFloorTex: true,
	}
	return t
}

func (t *TextureHandler) TextureAt(x, y, levelNum, side int) *ebiten.Image {
	texNum := -1

	mapLevel := t.mapObj.Level(levelNum)
	if mapLevel == nil {
		return nil
	}

	mapWidth := len(mapLevel)
	if mapWidth == 0 {
		return nil
	}
	mapHeight := len(mapLevel[0])
	if mapHeight == 0 {
		return nil
	}

	if x >= 0 && x < mapWidth && y >= 0 && y < mapHeight {
		texNum = mapLevel[x][y] - 1 // 1 subtracted from it so that texture 0 can be used
	}

	if side == 0 {
		//--some supid hacks to make the houses render correctly--//
		// this corrects textures on two sides of house since the textures are not symmetrical
		if texNum == 3 {
			texNum = 4
		} else if texNum == 4 {
			texNum = 3
		}

		if texNum == 1 {
			texNum = 4
		} else if texNum == 2 {
			texNum = 3
		}

		// make the ebitengine splash only show on one side
		if texNum == 5 {
			texNum = 0
		}
	}

	if texNum < 0 {
		return nil
	}
	return t.Textures[texNum]
}

func (t *TextureHandler) FloorTextureAt(x, y int) *image.RGBA {
	// x/y could be used to render different floor texture at given coords,
	// but for this demo we will just be rendering the same texture everywhere.
	if t.RenderFloorTex {
		return t.FloorTex
	}
	return nil
}
