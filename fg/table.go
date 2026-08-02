package fg

import (
	"github.com/sazardev/fugo/flog"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// TableWidget is a Material table whose cells are arbitrary widgets (unlike
// DataTable, which is limited to text cells). Build one with Table, passing
// the column count and the cells in row-major order: len(cells) must be a
// multiple of columnCount (the caller's responsibility — not validated at
// runtime). Set optional header labels with Columns.
type TableWidget struct {
	columns     []string // optional headers; empty = no header row
	columnCount int32
	cells       []Widget
	baseWidget
}

// Table creates a table with the given column count and cells, provided in
// row-major order (len(cells) must be a multiple of columnCount).
func Table(columnCount int32, cells ...Widget) *TableWidget {
	return &TableWidget{columnCount: columnCount, cells: cells}
}

// Columns sets the header labels and returns the widget for chaining.
func (t *TableWidget) Columns(headers ...string) *TableWidget {
	t.columns = headers

	return t
}

func (t *TableWidget) isWidget()                {}
func (t *TableWidget) widgetChildren() []Widget { return t.cells }

func (t *TableWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	t.id = *counter

	ids, nodes := walkChildren(t.cells, counter)
	props, err := proto.Marshal(&fugov1.TableProps{
		Columns:     t.columns,
		ColumnCount: t.columnCount,
	})
	if err != nil {
		flog.Errorf("marshal TableProps: %v", err)
	}

	return selfNode(t.id, t.key, fugov1.WidgetType_TABLE, props, ids, nodes)
}
