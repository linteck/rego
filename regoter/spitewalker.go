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
	x := rand.Intn(xlimit / 4)
	y := rand.Intn(ylimit)
	return iregoter.RgPosition{X: x, Y: y}
}

func (s *SpiteWalker) Update() (iregoter.RgPosition, *ebiten.Image) {
	s.count += 1
	if s.count%5 == 0 {
		s.position.X += rand.Intn(5)
		if s.position.X+frameWidth >= screenWidth || s.position.Y+frameHeight >= screenHeight {
			s.position = RandomPosition(screenWidth, screenHeight)
		}
	}
	i := (s.count / 5) % frameCount
	sx, sy := frameOX+i*frameWidth, frameOY
	//logger.Print("sub image (", sx, ", ", sy, ")")
	currImg := s.fullimg.SubImage(image.Rect(sx, sy, sx+frameWidth, sy+frameHeight)).(*ebiten.Image)
	return s.position, currImg
}

func NewSpiteWalker(coreMsgbox chan<- iregoter.IRegoterEvent) *Regoter[*SpiteWalker] {
	fullimg := LoadImage()
	p := RandomPosition(screenWidth, screenHeight)
	t := &SpiteWalker{position: p, fullimg: fullimg}
	r := NewRegoter(coreMsgbox, t)
	return r
}
