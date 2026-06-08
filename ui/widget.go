package ui

import fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"

type WidgetType = fugov1.WidgetType

const (
	WidgetText      = fugov1.WidgetType_TEXT
	WidgetContainer = fugov1.WidgetType_CONTAINER
	WidgetColumn    = fugov1.WidgetType_COLUMN
	WidgetCenter    = fugov1.WidgetType_CENTER
	WidgetButton    = fugov1.WidgetType_BUTTON
)

type Event struct {
	NodeID    string
	EventType string
	Data      []byte
}

type AppContext interface {
	Update()
}

type Widget interface {
	isWidget()
	widgetID() uint32
	setWidgetID(id uint32)
	widgetChildren() []Widget
	walkNodes(counter *uint32) []*fugov1.WidgetNode
	HasHandler() bool
	Handle(e Event)
}

type baseWidget struct {
	id uint32
}

func (b *baseWidget) widgetID() uint32      { return b.id }
func (b *baseWidget) setWidgetID(id uint32) { b.id = id }
func (b *baseWidget) HasHandler() bool      { return false }
func (b *baseWidget) Handle(Event)          {}

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

func collectIDs(w Widget, m map[uint32]Widget) {
	m[w.widgetID()] = w
	for _, child := range w.widgetChildren() {
		collectIDs(child, m)
	}
}
