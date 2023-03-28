package joycontrol

import (
	"testing"
)

func TestButtonAction(t *testing.T) {
	c := NewController()

	c.Press("UP")
	b := c.dump()
	t.Log(b)
	c.Release("UP")
	b = c.dump()
	t.Log(b)
}
