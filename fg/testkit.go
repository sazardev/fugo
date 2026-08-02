package fg

import "strconv"

// This file is Fugo's supported testing surface: synthetic Event constructors
// that make the wire format of Event.Data explicit, so a Go test can drive a
// widget's Handle method without guessing the ad-hoc string convention each
// widget's Flutter counterpart uses. See BuildTree's doc comment for the full
// pattern (build the tree, grab a widget reference, call Handle with one of
// these, assert on exported fields).

// ClickEvent returns a synthetic Event suitable for simulating a tap or click.
// Its Data is empty, matching handlers that ignore Event.Data entirely —
// Button/FilledButton/.../IconButton's OnClick, GestureDetector's OnTap, FAB's
// OnClick, and Radio's OnChange (Radio carries no value in Data; the handler
// itself sets GroupValue, as in the showcase template).
func ClickEvent() Event {
	return Event{EventType: "click"}
}

// BoolEvent returns a synthetic Event carrying a boolean toggle, matching the
// "1"/"0" wire format that Checkbox, Switch, CheckboxListTile, and
// SwitchListTile use (their OnChange handlers compare string(e.Data) == "1").
func BoolEvent(v bool) Event {
	data := "0"
	if v {
		data = "1"
	}

	return Event{EventType: "change", Data: []byte(data)}
}

// TextEvent returns a synthetic Event carrying a plain string value, matching
// the wire format TextField, Dropdown, and Autocomplete use — their OnChange
// handlers read the new value directly as string(e.Data).
func TextEvent(s string) Event {
	return Event{EventType: "change", Data: []byte(s)}
}

// FloatEvent returns a synthetic Event carrying a decimal string, matching the
// wire format Slider uses — its OnChange handler parses the new value with
// strconv.ParseFloat(string(e.Data), 64).
func FloatEvent(v float64) Event {
	return Event{EventType: "change", Data: []byte(strconv.FormatFloat(v, 'f', -1, 64))}
}

// RangeEvent returns a synthetic Event carrying a "start,end" pair, matching
// the wire format RangeSlider uses — its OnChange doc comment specifies both
// values formatted to 2 decimal places (e.g. "20.50,80.00"), split on "," and
// parsed with strconv.ParseFloat.
func RangeEvent(start, end float64) Event {
	data := strconv.FormatFloat(start, 'f', 2, 64) + "," + strconv.FormatFloat(end, 'f', 2, 64)

	return Event{EventType: "change", Data: []byte(data)}
}
