package fg

import (
	"github.com/sazardev/fugo/flog"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// SemanticsWidget wraps its child with an explicit accessibility annotation
// for screen readers. Build with Semantics.
type SemanticsWidget struct {
	child  Widget
	label  string
	hint   string
	button bool
	header bool
	baseWidget
}

// Semantics wraps child with an explicit accessibility annotation for screen
// readers — use it when Flutter can't infer a good one on its own (icon-only
// buttons, images, or a composition of several widgets that reads as one
// control).
func Semantics(child Widget) *SemanticsWidget {
	return &SemanticsWidget{child: child}
}

// Label sets the semantic label announced for this subtree.
func (w *SemanticsWidget) Label(text string) *SemanticsWidget {
	w.label = text

	return w
}

// Hint sets the semantic hint announced for this subtree.
func (w *SemanticsWidget) Hint(text string) *SemanticsWidget {
	w.hint = text

	return w
}

// Button marks this subtree as a single button control.
func (w *SemanticsWidget) Button(v bool) *SemanticsWidget {
	w.button = v

	return w
}

// Header marks this subtree as a heading.
func (w *SemanticsWidget) Header(v bool) *SemanticsWidget {
	w.header = v

	return w
}

func (w *SemanticsWidget) isWidget()                {}
func (w *SemanticsWidget) widgetChildren() []Widget { return oneChild(w.child) }

func (w *SemanticsWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	w.id = *counter
	ids, nodes := walkChildren(w.widgetChildren(), counter)
	props, err := proto.Marshal(&fugov1.SemanticsProps{
		Label:  w.label,
		Hint:   w.hint,
		Button: w.button,
		Header: w.header,
	})
	if err != nil {
		flog.Errorf("marshal SemanticsProps: %v", err)
	}

	return selfNode(w.id, w.key, fugov1.WidgetType_SEMANTICS, props, ids, nodes)
}
