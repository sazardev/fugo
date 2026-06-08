package fg

import fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"

type WidgetType = fugov1.WidgetType

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

type Event struct {
	NodeID    string
	EventType string
	Data      []byte
}

type AppContext interface {
	Update()
	NavigateTo(route string)
	GoBack()
}

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

type baseWidget struct {
	id  uint32
	key string
}

func (b *baseWidget) widgetID() uint32        { return b.id }
func (b *baseWidget) setWidgetID(id uint32)   { b.id = id }
func (b *baseWidget) widgetKey() string       { return b.key }
func (b *baseWidget) setWidgetKey(key string) { b.key = key }
func (b *baseWidget) HasHandler() bool        { return false }
func (b *baseWidget) Handle(Event)            {}

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

func BuildTreeOnly(root Widget) *fugov1.WidgetTree {
	tree, _ := BuildTree(root)

	return tree
}

func BuildTreeWithMerge(root Widget, oldMap map[uint32]Widget) (*fugov1.WidgetTree, map[uint32]Widget) {
	tree, newMap := BuildTree(root)
	for k, v := range newMap {
		oldMap[k] = v
	}

	return tree, oldMap
}

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
