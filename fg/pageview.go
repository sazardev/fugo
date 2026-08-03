package fg

import (
	"github.com/sazardev/fugo/flog"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// PageViewWidget is a swipeable set of full-size pages. Build one with
// PageView, optionally set the initial page with InitialPage, and switch to
// vertical scrolling with Vertical.
type PageViewWidget struct {
	children        []Widget
	initialPage     int32
	scrollDirection int32 // 0 = horizontal (default), 1 = vertical
	baseWidget
}

// PageView creates a page view containing the given pages.
func PageView(pages ...Widget) *PageViewWidget {
	return &PageViewWidget{children: pages}
}

// InitialPage sets the page shown first and returns the widget for chaining.
func (p *PageViewWidget) InitialPage(i int32) *PageViewWidget {
	p.initialPage = i

	return p
}

// Vertical switches the scroll axis to vertical and returns the widget for chaining.
func (p *PageViewWidget) Vertical() *PageViewWidget {
	p.scrollDirection = 1

	return p
}

func (p *PageViewWidget) isWidget()                {}
func (p *PageViewWidget) widgetChildren() []Widget { return p.children }

func (p *PageViewWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	p.id = *counter

	ids, nodes := walkChildren(p.children, counter)
	props, err := proto.Marshal(&fugov1.PageViewProps{
		InitialPage:     p.initialPage,
		ScrollDirection: p.scrollDirection,
	})
	if err != nil {
		flog.Errorf("marshal PageViewProps: %v", err)
	}

	return selfNode(p.id, p.key, fugov1.WidgetType_PAGEVIEW, props, ids, nodes)
}
