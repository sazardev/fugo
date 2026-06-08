package fg

import fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"

type ColumnWidget struct {
	items []Widget
	baseWidget
}

func Column(items ...Widget) *ColumnWidget {
	return &ColumnWidget{items: items}
}

func (c *ColumnWidget) isWidget() {}

func (c *ColumnWidget) widgetChildren() []Widget { return c.items }

func (c *ColumnWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
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

type CenterWidget struct {
	child Widget
	baseWidget
}

func Center(child Widget) *CenterWidget {
	return &CenterWidget{child: child}
}

func (c *CenterWidget) isWidget() {}

func (c *CenterWidget) widgetChildren() []Widget {
	if c.child != nil {
		return []Widget{c.child}
	}

	return nil
}

func (c *CenterWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
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
