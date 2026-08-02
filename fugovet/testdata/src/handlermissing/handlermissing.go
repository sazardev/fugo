// Package a exercises the handlermissing analyzer.
package a

import "github.com/sazardev/fugo/fg"

// noHandlerChained: passed straight into Row with no OnClick anywhere —
// should be flagged.
func noHandlerChained() fg.Widget {
	return fg.Row(
		fg.Button("go"), // want `Button has no OnClick handler`
	)
}

// noHandlerReturned: returned directly with no OnClick — should be flagged.
func noHandlerReturned() fg.Widget {
	return fg.IconButton("home") // want `IconButton has no OnClick handler`
}

// chainedOnClick: OnClick is part of the chain — no diagnostic expected.
func chainedOnClick() fg.Widget {
	return fg.Button("go").BgColor("red").OnClick(func(fg.Event) {})
}

// deferredOnClick: the button is assigned to a variable, and OnClick is
// registered later in the same block — no diagnostic expected.
func deferredOnClick() fg.Widget {
	btn := fg.FilledButton("go")
	btn.OnClick(func(fg.Event) {})

	return btn
}

// deferredOnClickViaVar: OnClick called on the variable after other setup —
// still no diagnostic expected.
func deferredOnClickViaVar() fg.Widget {
	btn := fg.OutlinedButton("go")
	btn.BgColor("blue")
	btn.OnClick(func(fg.Event) {})

	return btn
}

// assignedButNeverWired: assigned to a variable but OnClick is never called
// anywhere — should be flagged.
func assignedButNeverWired() fg.Widget {
	btn := fg.TextButton("go") // want `TextButton has no OnClick handler`

	return btn
}
