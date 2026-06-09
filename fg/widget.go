package fg

import fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"

// WidgetType enumerates the kinds of widget the client knows how to render.
type WidgetType = fugov1.WidgetType

// Widget type identifiers, one per supported widget kind. They mirror the
// WidgetType enum in the protobuf schema.
const (
	WidgetText               = fugov1.WidgetType_TEXT
	WidgetContainer          = fugov1.WidgetType_CONTAINER
	WidgetColumn             = fugov1.WidgetType_COLUMN
	WidgetCenter             = fugov1.WidgetType_CENTER
	WidgetButton             = fugov1.WidgetType_BUTTON
	WidgetRow                = fugov1.WidgetType_ROW
	WidgetStack              = fugov1.WidgetType_STACK
	WidgetExpanded           = fugov1.WidgetType_EXPANDED
	WidgetPadding            = fugov1.WidgetType_PADDING
	WidgetSizedBox           = fugov1.WidgetType_SIZEDBOX
	WidgetImage              = fugov1.WidgetType_IMAGE
	WidgetTextField          = fugov1.WidgetType_TEXTFIELD
	WidgetPositioned         = fugov1.WidgetType_POSITIONED
	WidgetCheckbox           = fugov1.WidgetType_CHECKBOX
	WidgetSwitchWidget       = fugov1.WidgetType_SWITCH_WIDGET
	WidgetSlider             = fugov1.WidgetType_SLIDER
	WidgetListView           = fugov1.WidgetType_LISTVIEW
	WidgetAnimatedContainer  = fugov1.WidgetType_ANIMATEDCONTAINER
	WidgetIcon               = fugov1.WidgetType_ICON
	WidgetDivider            = fugov1.WidgetType_DIVIDER
	WidgetWrap               = fugov1.WidgetType_WRAP
	WidgetGridView           = fugov1.WidgetType_GRIDVIEW
	WidgetAnimatedOpacity    = fugov1.WidgetType_ANIMATEDOPACITY
	WidgetScrollView         = fugov1.WidgetType_SCROLLVIEW
	WidgetGestureDetector    = fugov1.WidgetType_GESTUREDETECTOR
	WidgetAlign              = fugov1.WidgetType_ALIGN
	WidgetRadio              = fugov1.WidgetType_RADIO
	WidgetDropdown           = fugov1.WidgetType_DROPDOWN
	WidgetAnimatedPositioned = fugov1.WidgetType_ANIMATEDPOSITIONED
	WidgetWindowDragArea     = fugov1.WidgetType_WINDOWDRAGAREA
	WidgetCard               = fugov1.WidgetType_CARD
	WidgetScaffold           = fugov1.WidgetType_SCAFFOLD
	WidgetFAB                = fugov1.WidgetType_FLOATINGACTIONBUTTON
	WidgetListTile           = fugov1.WidgetType_LISTTILE
	WidgetChip               = fugov1.WidgetType_CHIP
	WidgetProgress           = fugov1.WidgetType_PROGRESS
	WidgetAppBar             = fugov1.WidgetType_APPBAR
	WidgetNavigationBar      = fugov1.WidgetType_NAVIGATIONBAR
	WidgetTabs               = fugov1.WidgetType_TABS
	WidgetTooltip            = fugov1.WidgetType_TOOLTIP
	WidgetBadge              = fugov1.WidgetType_BADGE
	WidgetAvatar             = fugov1.WidgetType_AVATAR
	WidgetSegmentedButton    = fugov1.WidgetType_SEGMENTEDBUTTON
	WidgetAspectRatio        = fugov1.WidgetType_ASPECTRATIO
	WidgetClipRRect          = fugov1.WidgetType_CLIPRRECT
	WidgetFittedBox          = fugov1.WidgetType_FITTEDBOX
	WidgetFlexible           = fugov1.WidgetType_FLEXIBLE
	WidgetExpansionTile      = fugov1.WidgetType_EXPANSIONTILE
	WidgetPopupMenu          = fugov1.WidgetType_POPUPMENU
)

// Event is a user interaction forwarded from the client to a widget's handler.
type Event struct {
	NodeID    string
	EventType string
	Data      []byte
}

// AppContext is the slice of the application lifecycle that widgets and
// handlers may use: scheduling a re-render and driving navigation.
type AppContext interface {
	Update()
	NavigateTo(route string)
	GoBack()
}

// Widget is implemented by every Fugo widget. Most methods are unexported
// bookkeeping used by the engine; HasHandler and Handle expose event dispatch.
type Widget interface {
	isWidget()
	widgetID() uint32
	setWidgetID(id uint32)
	widgetKey() string
	setWidgetKey(key string)
	widgetChildren() []Widget
	walkNodes(counter *uint32) []*fugov1.WidgetNode
	HasHandler() bool
	Handle(e Event)
}

// baseWidget carries the id/key bookkeeping embedded by every widget and
// provides the no-op handler defaults for widgets that take no events.
type baseWidget struct {
	id  uint32
	key string
}

func (b *baseWidget) widgetID() uint32        { return b.id }
func (b *baseWidget) setWidgetID(id uint32)   { b.id = id }
func (b *baseWidget) widgetKey() string       { return b.key }
func (b *baseWidget) setWidgetKey(key string) { b.key = key }

// HasHandler reports whether the widget has an event handler. The base
// implementation returns false; interactive widgets override it.
func (b *baseWidget) HasHandler() bool { return false }

// Handle dispatches an event to the widget. The base implementation is a no-op.
func (b *baseWidget) Handle(Event) {}

// BuildTree walks root depth-first, assigning stable ids, and returns the
// serialized widget tree together with a map from node id to widget (used to
// route events back to handlers).
func BuildTree(root Widget) (*fugov1.WidgetTree, map[uint32]Widget) {
	var counter uint32
	nodes := root.walkNodes(&counter)

	widgetByID := make(map[uint32]Widget)
	collectIDs(root, widgetByID)

	return &fugov1.WidgetTree{
		Root:  nodes[0].GetId(),
		Nodes: nodes,
	}, widgetByID
}

// BuildTreeOnly is BuildTree without the id→widget map, for callers that only
// need the serialized tree.
func BuildTreeOnly(root Widget) *fugov1.WidgetTree {
	tree, _ := BuildTree(root)

	return tree
}

// BuildTreeWithMerge rebuilds the tree from root and merges the freshly
// assigned id→widget entries into oldMap, returning the merged map. It is used
// each frame so handlers registered on the retained tree stay reachable.
func BuildTreeWithMerge(root Widget, oldMap map[uint32]Widget) (*fugov1.WidgetTree, map[uint32]Widget) {
	tree, newMap := BuildTree(root)
	for k, v := range newMap {
		oldMap[k] = v
	}

	return tree, oldMap
}

// WithKey assigns a stable identity key to w and returns it as a Widget. Prefer
// Keyed when you want to keep chaining concrete setters — it preserves the
// widget's concrete type.
//
// A key labels a node's identity so the diff treats a key change as a real
// change (treesEqual compares keys). Keys are most useful on the children of a
// dynamic list, where they document which item a node represents across frames.
func WithKey(w Widget, key string) Widget {
	w.setWidgetKey(key)

	return w
}

// Keyed assigns a stable identity key to w and returns w unchanged so callers
// can keep chaining concrete setters, e.g.
//
//	fg.Keyed(fg.Text("Alice"), "user-1").FontSize(18)
//
// It is the generic, type-preserving counterpart of WithKey. See WithKey for
// what keys mean.
func Keyed[T Widget](w T, key string) T {
	w.setWidgetKey(key)

	return w
}

// Key reads the identity key previously assigned to w via Keyed/WithKey, or ""
// if none was set.
func Key(w Widget) string {
	return w.widgetKey()
}

func collectIDs(w Widget, m map[uint32]Widget) {
	m[w.widgetID()] = w
	for _, child := range w.widgetChildren() {
		collectIDs(child, m)
	}
}

// walkChildren assigns ids to each child widget (depth-first, via the shared
// counter) and returns their root ids — for the parent's Children field —
// together with the flattened list of all descendant nodes.
func walkChildren(children []Widget, counter *uint32) ([]uint32, []*fugov1.WidgetNode) {
	var childIDs []uint32
	var allNodes []*fugov1.WidgetNode

	for _, child := range children {
		subNodes := child.walkNodes(counter)
		if len(subNodes) > 0 {
			childIDs = append(childIDs, subNodes[0].GetId())
			allNodes = append(allNodes, subNodes...)
		}
	}

	return childIDs, allNodes
}

// oneChild wraps child in a slice, or returns nil for a nil child.
func oneChild(child Widget) []Widget {
	if child == nil {
		return nil
	}

	return []Widget{child}
}

// selfNode builds a parent node and prepends it to its descendants — the common
// tail of a widget's walkNodes.
func selfNode(id uint32, key string, t fugov1.WidgetType, props []byte, childIDs []uint32, descendants []*fugov1.WidgetNode) []*fugov1.WidgetNode {
	return append([]*fugov1.WidgetNode{{
		Id:       id,
		Key:      key,
		Type:     t,
		Props:    props,
		Children: childIDs,
	}}, descendants...)
}
