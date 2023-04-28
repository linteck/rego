package regoter

import (
	"image"
	"lintech/rego/iregoter"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

type SpiteWalker struct {
	position iregoter.RgPosition
	fullimg  *ebiten.Image
	currImg  *ebiten.Image
	count    int
}

const (
	screenWidth  = 320
	screenHeight = 240

	frameOX     = 0
	frameOY     = 32
	frameWidth  = 32
	frameHeight = 32
	frameCount  = 8
)

func RandomPosition(xlimit int, ylimit int) iregoter.RgPosition {
	x := rand.Intn(xlimit)
	y := rand.Intn(ylimit)
	return iregoter.RgPosition{X: x, Y: y}
}

func (s *SpiteWalker) Update() iregoter.RegoterUpdatedInfo {
	changed := false
	s.count += 1
	if s.count%5 == 0 {
		changed = true
		s.position.X += rand.Intn(5)
		if s.position.X+frameWidth >= screenWidth {
			s.position.X = 0
		}
		if s.position.Y+frameHeight >= screenHeight {
			s.position.Y = rand.Intn(screenHeight - frameHeight)
		}
		i := (s.count / 5)
		s.currImg = GetSubImage(s.fullimg, i)
	}
	info := iregoter.RegoterUpdatedInfo{Position: s.position, Img: s.currImg,
		Visiable: true, Deleted: false, Changed: changed}
	return info
}

func NewSpiteWalker(coreMsgbox chan<- iregoter.IRegoterEvent) *Regoter[*SpiteWalker] {
	p := RandomPosition(screenWidth/4, screenHeight-frameHeight)
	fullimg := LoadImage()
	currImg := GetSubImage(fullimg, 0)
	t := &SpiteWalker{position: p, fullimg: fullimg, currImg: currImg}
	r := NewRegoter(coreMsgbox, t)
	return r
}

func GetSubImage(fullimg *ebiten.Image, index int) *ebiten.Image {
	i := index % frameCount
	sx, sy := frameOX+i*frameWidth, frameOY
	//logger.Print("sub image (", sx, ", ", sy, ")")
	img := fullimg.SubImage(image.Rect(sx, sy, sx+frameWidth, sy+frameHeight)).(*ebiten.Image)
	return img
}
