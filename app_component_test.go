package fugo

import (
	"testing"

	"github.com/sazardev/fugo/fg"
)

type greeter struct{ name string }

func (g *greeter) Render(_ *Context) fg.Widget {
	return fg.Text("hi " + g.name)
}

func TestComponentSatisfiesInterface(t *testing.T) {
	var c Component = &greeter{name: "fugo"}

	tree, _ := fg.BuildTree(c.Render(nil))
	if len(tree.GetNodes()) != 1 {
		t.Errorf("expected 1 node from component render, got %d", len(tree.GetNodes()))
	}
}
