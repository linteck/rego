package model

type ICooldownFlag interface {
	set()
	get() bool
	cooldown()
}

type cooldownFlag struct {
	counter     int
	counterInit int
	flag        bool
}

func (c *cooldownFlag) set() {
	if c.counter <= 0 {
		c.flag = true
		c.counter = c.counterInit
	} else {
		// ignore seting
	}
}
func (c *cooldownFlag) get() bool {
	f := c.flag
	c.flag = false
	return f
}

func (c *cooldownFlag) cooldown() {
	if c.counter > 0 {
		c.counter -= 1
	}
}

type ICooldownInt interface {
	add(int) int
	cooldown()
}

type cooldownInt struct {
	counter     int
	counterInit int
	value       int
}

func (c *cooldownInt) add(v int) int {
	if c.counter <= 0 {
		c.value += v
		c.counter = c.counterInit
	} else {
		// ignore
	}
	return c.value
}
func (c *cooldownInt) get() int {
	return c.value
}

func (c *cooldownInt) cooldown() {
	if c.counter > 0 {
		c.counter -= 1
	}
}
