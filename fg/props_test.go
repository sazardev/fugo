package fg

import (
	"testing"

	"github.com/sazardev/fugo/style"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

func nodeOfType(tree *fugov1.WidgetTree, typ fugov1.WidgetType) *fugov1.WidgetNode {
	for _, n := range tree.GetNodes() {
		if n.GetType() == typ {
			return n
		}
	}

	return nil
}

// TestTextPropsRoundTrip ensures weight and alignment reach the wire. Both were
// settable on TextWidget but never marshaled in walkNodes, so they silently
// never rendered.
func TestTextPropsRoundTrip(t *testing.T) {
	txt := Text("hi").FontSize(20).Weight(style.WeightBold).Align(style.AlignRight)

	tree, _ := BuildTree(txt)

	node := nodeOfType(tree, fugov1.WidgetType_TEXT)
	if node == nil {
		t.Fatal("no TEXT node produced")
	}

	var props fugov1.TextProps
	if err := proto.Unmarshal(node.GetProps(), &props); err != nil {
		t.Fatalf("unmarshal TextProps: %v", err)
	}

	if props.GetFontWeight() != int32(style.WeightBold) {
		t.Errorf("font weight = %d, want %d", props.GetFontWeight(), int32(style.WeightBold))
	}

	if props.GetTextAlign() != int32(style.AlignRight) {
		t.Errorf("text align = %d, want %d", props.GetTextAlign(), int32(style.AlignRight))
	}

	if props.GetFontSize() != 20 {
		t.Errorf("font size = %v, want 20", props.GetFontSize())
	}
}

// TestContainerPaddingRoundTrip ensures all four padding edges reach the wire.
// walkNodes previously marshaled only the top edge, so asymmetric padding was
// silently collapsed.
func TestContainerPaddingRoundTrip(t *testing.T) {
	// EdgeOnly(top, right, bottom, left)
	c := Container(Text("x")).Pad(style.EdgeOnly(1, 2, 3, 4))

	tree, _ := BuildTree(c)

	node := nodeOfType(tree, fugov1.WidgetType_CONTAINER)
	if node == nil {
		t.Fatal("no CONTAINER node produced")
	}

	var props fugov1.ContainerProps
	if err := proto.Unmarshal(node.GetProps(), &props); err != nil {
		t.Fatalf("unmarshal ContainerProps: %v", err)
	}

	got := [4]float64{props.GetPadTop(), props.GetPadRight(), props.GetPadBottom(), props.GetPadLeft()}
	if got != [4]float64{1, 2, 3, 4} {
		t.Errorf("padding edges [top right bottom left] = %v, want [1 2 3 4]", got)
	}
}
