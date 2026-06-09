package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// DataTableWidget is a Material data table of text cells. Build one with
// DataTable, set the header with Columns, and add data with Row.
type DataTableWidget struct {
	columns []string
	rows    [][]string
	baseWidget
}

// DataTable creates an empty data table.
func DataTable() *DataTableWidget {
	return &DataTableWidget{}
}

// Columns sets the header labels and returns the widget for chaining.
func (d *DataTableWidget) Columns(cols ...string) *DataTableWidget {
	d.columns = cols

	return d
}

// Row appends a data row (one cell per column) and returns the widget for chaining.
func (d *DataTableWidget) Row(cells ...string) *DataTableWidget {
	d.rows = append(d.rows, cells)

	return d
}

func (d *DataTableWidget) isWidget()                {}
func (d *DataTableWidget) widgetChildren() []Widget { return nil }

func (d *DataTableWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	d.id = *counter

	rows := make([]*fugov1.DataRow, 0, len(d.rows))
	for _, r := range d.rows {
		rows = append(rows, &fugov1.DataRow{Cells: r})
	}

	props, _ := proto.Marshal(&fugov1.DataTableProps{
		Columns: d.columns,
		Rows:    rows,
	})

	return []*fugov1.WidgetNode{{
		Id:    d.id,
		Key:   d.key,
		Type:  fugov1.WidgetType_DATATABLE,
		Props: props,
	}}
}
