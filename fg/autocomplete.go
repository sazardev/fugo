package fg

import (
	"github.com/sazardev/fugo/flog"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// AutocompleteWidget is a text input that suggests matches from a fixed list
// of options as the user types. Build one with Autocomplete.
type AutocompleteWidget struct {
	handler     func(Event)
	options     []string
	value       string
	placeholder string
	baseWidget
}

// Autocomplete creates a text input that suggests matches from options.
func Autocomplete(options []string) *AutocompleteWidget {
	return &AutocompleteWidget{options: options}
}

// SetValue sets the current text and returns the widget for chaining.
func (a *AutocompleteWidget) SetValue(v string) *AutocompleteWidget {
	a.value = v

	return a
}

// SetPlaceholder sets the placeholder shown when the input is empty and
// returns the widget for chaining.
func (a *AutocompleteWidget) SetPlaceholder(v string) *AutocompleteWidget {
	a.placeholder = v

	return a
}

// OnChange registers the handler invoked when an option is chosen/committed.
// e.Data carries the final text as string(e.Data), same as TextField/Dropdown.
func (a *AutocompleteWidget) OnChange(handler func(Event)) *AutocompleteWidget {
	a.handler = handler

	return a
}

func (a *AutocompleteWidget) isWidget()                {}
func (a *AutocompleteWidget) widgetChildren() []Widget { return nil }

// HasHandler reports whether an OnChange handler has been registered.
func (a *AutocompleteWidget) HasHandler() bool { return a.handler != nil }

// Handle dispatches event to the registered OnChange handler, if any.
func (a *AutocompleteWidget) Handle(event Event) {
	if a.handler != nil {
		a.handler(event)
	}
}

func (a *AutocompleteWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	a.id = *counter

	props, err := proto.Marshal(&fugov1.AutocompleteProps{
		Options:     a.options,
		Value:       a.value,
		Placeholder: a.placeholder,
	})
	if err != nil {
		flog.Errorf("marshal AutocompleteProps: %v", err)
	}

	return []*fugov1.WidgetNode{{
		Id:    a.id,
		Key:   a.key,
		Type:  fugov1.WidgetType_AUTOCOMPLETE,
		Props: props,
	}}
}
