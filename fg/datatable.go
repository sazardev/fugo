package fg

import (
	"strconv"
	"strings"

	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// DataTableWidget is a Material data table of text cells. Build one with
// DataTable, set the header with Columns, and add data with Row.
type DataTableWidget struct {
	onSort     func(columnIndex int, ascending bool)
	onSelect   func(rowIndex int, selected bool)
	columns    []string
	rows       [][]string
	selected   []bool
	sortColumn int
	sortAsc    bool
	sortable   bool
	selectable bool
	baseWidget
}

// DataTable creates an empty data table.
func DataTable() *DataTableWidget {
	return &DataTableWidget{sortColumn: -1}
}

// Columns sets the header labels and returns the widget for chaining.
func (d *DataTableWidget) Columns(cols ...string) *DataTableWidget {
	d.columns = cols

	return d
}

// Row appends a data row (one cell per column) and returns the widget for chaining.
func (d *DataTableWidget) Row(cells ...string) *DataTableWidget {
	d.rows = append(d.rows, cells)
	d.selected = append(d.selected, false)

	return d
}

// Sortable enables tappable column headers and registers the handler invoked
// with the tapped column index and the requested sort direction; Go owns
// reordering d's rows (call Columns/Row again, or mutate the underlying data
// and rebuild) and should also call SetSort to reflect the new state in the
// header arrows. Returns the widget for chaining.
func (d *DataTableWidget) Sortable(handler func(columnIndex int, ascending bool)) *DataTableWidget {
	d.sortable = true
	d.onSort = handler

	return d
}

// SetSort sets which column shows the sort arrow and its direction, purely
// cosmetic (Go must reorder the rows itself). Returns the widget for chaining.
func (d *DataTableWidget) SetSort(columnIndex int, ascending bool) *DataTableWidget {
	d.sortColumn = columnIndex
	d.sortAsc = ascending

	return d
}

// Selectable shows a leading checkbox column and registers the handler
// invoked with the toggled row index and its new selected state. Returns the
// widget for chaining.
func (d *DataTableWidget) Selectable(handler func(rowIndex int, selected bool)) *DataTableWidget {
	d.selectable = true
	d.onSelect = handler

	return d
}

// SetSelected sets whether row i is checked; out-of-range indices are
// ignored. Returns the widget for chaining.
func (d *DataTableWidget) SetSelected(rowIndex int, selected bool) *DataTableWidget {
	if rowIndex >= 0 && rowIndex < len(d.selected) {
		d.selected[rowIndex] = selected
	}

	return d
}

func (d *DataTableWidget) isWidget()                {}
func (d *DataTableWidget) widgetChildren() []Widget { return nil }

// HasHandler reports whether a Sortable or Selectable handler has been registered.
func (d *DataTableWidget) HasHandler() bool { return d.onSort != nil || d.onSelect != nil }

// Handle dispatches a "sort" or "select" event to the matching registered handler.
func (d *DataTableWidget) Handle(event Event) {
	parts := strings.SplitN(string(event.Data), ",", 2)
	if len(parts) != 2 {
		return
	}

	idx, err := strconv.Atoi(parts[0])
	if err != nil {
		return
	}

	flag := parts[1] == "1"

	switch event.EventType {
	case "sort":
		if d.onSort != nil {
			d.onSort(idx, flag)
		}
	case "select":
		if d.onSelect != nil {
			d.onSelect(idx, flag)
		}
	}
}

func (d *DataTableWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	d.id = *counter

	rows := make([]*fugov1.DataRow, 0, len(d.rows))
	for _, r := range d.rows {
		rows = append(rows, &fugov1.DataRow{Cells: r})
	}

	selected := d.selected
	if selected == nil && len(d.rows) > 0 {
		selected = make([]bool, len(d.rows))
	}

	props, _ := proto.Marshal(&fugov1.DataTableProps{
		Columns:         d.columns,
		Rows:            rows,
		Sortable:        d.sortable,
		SortColumnIndex: int32(d.sortColumn), //nolint:gosec // small column count
		SortAscending:   d.sortAsc,
		Selectable:      d.selectable,
		SelectedRows:    selected,
	})

	return []*fugov1.WidgetNode{{
		Id:    d.id,
		Key:   d.key,
		Type:  fugov1.WidgetType_DATATABLE,
		Props: props,
	}}
}
