package fg

import fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"

type Column struct {
	items []Widget
	baseWidget
}

func Column(items ...Widget) *Column {
	return &Column{items: items}
}

func (c *Column) isWidget() {}

func (c *Column) widgetChildren() []Widget { return c.items }

func (c *Column) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	c.id = *counter

	var childIDs []uint32

	var allNodes []*fugov1.WidgetNode

	for _, child := range c.widgetChildren() {
		subNodes := child.walkNodes(counter)
		if len(subNodes) > 0 {
			childIDs = append(childIDs, subNodes[0].GetId())
			allNodes = append(allNodes, subNodes...)
		}
	}

	self := &fugov1.WidgetNode{
		Id:       c.id,
		Key:      c.key,
		Type:     fugov1.WidgetType_COLUMN,
		Children: childIDs,
	}

	return append([]*fugov1.WidgetNode{self}, allNodes...)
}

type Center struct {
	child Widget
	baseWidget
}

func Center(child Widget) *Center {
	return &Center{child: child}
}

func (c *Center) isWidget() {}

func (c *Center) widgetChildren() []Widget {
	if c.child != nil {
		return []Widget{c.child}
	}

	return nil
}

func (c *Center) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	c.id = *counter

	var childIDs []uint32

	var allNodes []*fugov1.WidgetNode

	for _, child := range c.widgetChildren() {
		subNodes := child.walkNodes(counter)
		if len(subNodes) > 0 {
			childIDs = append(childIDs, subNodes[0].GetId())
			allNodes = append(allNodes, subNodes...)
		}
	}

	self := &fugov1.WidgetNode{
		Id:       c.id,
		Key:      c.key,
		Type:     fugov1.WidgetType_CENTER,
		Children: childIDs,
	}

	return append([]*fugov1.WidgetNode{self}, allNodes...)
}
