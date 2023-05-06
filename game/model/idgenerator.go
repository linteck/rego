package model

const WALL_ID = 0

type IdRx <-chan ID

func NewGenerator(start ID) IdRx {
	c := make(chan ID)
	id := start
	go func() {
		for {
			c <- id
			id += 1
		}
	}()
	return c
}

// ID start from 100
var IdGen = NewGenerator(100)
