package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// FormWidget lays out children vertically (typically fields followed by a
// submit button) and aggregates per-field validation, entirely on the Go
// side: Fugo has no client-owned FormField state to orchestrate, so a Form
// is just a Column plus a validator registry and an optional error banner.
type FormWidget struct {
	items     []Widget
	validate  []func() string
	errorText string
	baseWidget
}

// Form creates a form containing the given items (fields, buttons, etc.).
func Form(items ...Widget) *FormWidget {
	return &FormWidget{items: items}
}

// AddField registers a validator run by Validate; it should set the field's
// own error state (e.g. via TextFieldWidget.SetError) and return a non-empty
// message when invalid. Returns the widget for chaining.
func (f *FormWidget) AddField(validate func() string) *FormWidget {
	f.validate = append(f.validate, validate)

	return f
}

// Validate runs every registered field validator and reports whether all of
// them passed. Call it from a submit button's OnClick before acting on the
// form's data.
func (f *FormWidget) Validate() bool {
	ok := true

	for _, v := range f.validate {
		if v() != "" {
			ok = false
		}
	}

	return ok
}

// SetError sets a form-level error banner shown above the fields (e.g. after
// Validate fails, or a submit request comes back rejected); empty clears it.
// Returns the widget for chaining.
func (f *FormWidget) SetError(msg string) *FormWidget {
	f.errorText = msg

	return f
}

func (f *FormWidget) isWidget() {}

func (f *FormWidget) widgetChildren() []Widget { return f.items }

func (f *FormWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	f.id = *counter

	childIDs, allNodes := walkChildren(f.widgetChildren(), counter)

	props, _ := proto.Marshal(&fugov1.FormProps{
		ErrorText: f.errorText,
	})

	self := &fugov1.WidgetNode{
		Id:       f.id,
		Key:      f.key,
		Type:     fugov1.WidgetType_FORM,
		Props:    props,
		Children: childIDs,
	}

	return append([]*fugov1.WidgetNode{self}, allNodes...)
}
