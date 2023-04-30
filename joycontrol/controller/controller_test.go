package controller

import (
	"testing"
)

func TestButtonAction(t *testing.T) {
	c := NewController()

	c.Press("UP")
	b := c.Dump()
	t.Log(b)
	c.Release("UP")
	b = c.Dump()
	t.Log(b)
}
