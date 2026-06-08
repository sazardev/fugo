package fg

import fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"

// WidgetType enumerates the kinds of widget the client knows how to render.
type WidgetType = fugov1.WidgetType

// Widget type identifiers, one per supported widget kind. They mirror the
// WidgetType enum in the protobuf schema.
const (
	WidgetText              = fugov1.WidgetType_TEXT
	WidgetContainer         = fugov1.WidgetType_CONTAINER
	WidgetColumn            = fugov1.WidgetType_COLUMN
	WidgetCenter            = fugov1.WidgetType_CENTER
	WidgetButton            = fugov1.WidgetType_BUTTON
	WidgetRow               = fugov1.WidgetType_ROW
	WidgetStack             = fugov1.WidgetType_STACK
	WidgetExpanded          = fugov1.WidgetType_EXPANDED
	WidgetPadding           = fugov1.WidgetType_PADDING
	WidgetSizedBox          = fugov1.WidgetType_SIZEDBOX
	WidgetImage             = fugov1.WidgetType_IMAGE
	WidgetTextField         = fugov1.WidgetType_TEXTFIELD
	WidgetPositioned        = fugov1.WidgetType_POSITIONED
	WidgetCheckbox          = fugov1.WidgetType_CHECKBOX
	WidgetSwitchWidget      = fugov1.WidgetType_SWITCH_WIDGET
	WidgetSlider            = fugov1.WidgetType_SLIDER
	WidgetListView          = fugov1.WidgetType_LISTVIEW
	WidgetAnimatedContainer = fugov1.WidgetType_ANIMATEDCONTAINER
	WidgetIcon              = fugov1.WidgetType_ICON
	WidgetDivider           = fugov1.WidgetType_DIVIDER
	WidgetWrap              = fugov1.WidgetType_WRAP
	WidgetGridView          = fugov1.WidgetType_GRIDVIEW
	WidgetAnimatedOpacity   = fugov1.WidgetType_ANIMATEDOPACITY
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

// WithKey assigns a stable key to w (used by the diff to track identity across
// frames in dynamic lists) and returns w for chaining.
func WithKey(w Widget, key string) Widget {
	w.setWidgetKey(key)

	return w
}

func collectIDs(w Widget, m map[uint32]Widget) {
	m[w.widgetID()] = w
	for _, child := range w.widgetChildren() {
		collectIDs(child, m)
	}
}
