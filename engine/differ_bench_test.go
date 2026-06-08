package engine_test

import (
	"testing"

	"github.com/sazardev/fugo/engine"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
)

func makeTree(nodeCount int) *fugov1.WidgetTree {
	nodes := make([]*fugov1.WidgetNode, nodeCount)
	for i := range nodeCount {
		props := make([]byte, 32)
		for j := range props {
			props[j] = byte(i + j)
		}

		var children []uint32
		if i > 0 && i%5 == 0 {
			children = []uint32{uint32(i - 2), uint32(i - 1)}
		}

		nodes[i] = &fugov1.WidgetNode{
			Id:       uint32(i + 1),
			Type:     fugov1.WidgetType_TEXT,
			Props:    props,
			Children: children,
		}
	}

	return &fugov1.WidgetTree{Root: 1, Nodes: nodes}
}

func BenchmarkDiff_100(b *testing.B) {
	oldTree := makeTree(100)
	newTree := makeTree(100)

	b.ResetTimer()
	for range b.N {
		engine.Diff(oldTree, newTree)
	}
}

func BenchmarkDiff_500(b *testing.B) {
	oldTree := makeTree(500)
	newTree := makeTree(500)

	b.ResetTimer()
	for range b.N {
		engine.Diff(oldTree, newTree)
	}
}

func BenchmarkDiff_1000(b *testing.B) {
	oldTree := makeTree(1000)
	newTree := makeTree(1000)

	b.ResetTimer()
	for range b.N {
		engine.Diff(oldTree, newTree)
	}
}

func BenchmarkDiff_100_WithChange(b *testing.B) {
	oldTree := makeTree(100)
	newTree := makeTree(100)
	newTree.Nodes[50].Props = []byte("changed")

	b.ResetTimer()
	for range b.N {
		engine.Diff(oldTree, newTree)
	}
}

func BenchmarkDiff_1000_WithChange(b *testing.B) {
	oldTree := makeTree(1000)
	newTree := makeTree(1000)
	newTree.Nodes[500].Props = []byte("changed")

	b.ResetTimer()
	for range b.N {
		engine.Diff(oldTree, newTree)
	}
}

func BenchmarkDiff_FullCreate_1000(b *testing.B) {
	tree := makeTree(1000)

	b.ResetTimer()
	for range b.N {
		engine.Diff(nil, tree)
	}
}
