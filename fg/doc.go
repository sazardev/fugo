// Package fg is Fugo's declarative widget API: the building blocks you compose
// in Go to describe a UI that a Flutter client renders.
//
// Constructors are prefix-free — [Text], [Button], [Container], [Column],
// [Row], [Stack], [Router], and so on — and each returns a concrete *XxxWidget
// with chainable setters:
//
//	fg.Button("Save").BgColor(fg.Hex("#10B981")).OnClick(handler)
//
// Widgets are mutable and retained: handlers mutate them in place (for example
// [TextWidget.SetText]) and call Context.Update to schedule a re-render.
//
// Visual defaults (colors, sizes, radii) come from the active [Theme]; set it
// once with [UseTheme] and override per widget with the chainable setters.
package fg
