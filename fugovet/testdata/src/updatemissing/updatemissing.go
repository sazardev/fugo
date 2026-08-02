// Package a exercises the updatemissing analyzer.
package a

import (
	"strconv"

	"github.com/sazardev/fugo"
	"github.com/sazardev/fugo/fg"
)

// missingInHandler: the OnClick closure mutates the widget but never calls
// ctx.Update() anywhere in its own body — should be flagged.
func missingInHandler(ctx *fugo.Context) fg.Widget {
	count := fg.Text("0")
	btn := fg.Button("inc")
	btn.OnClick(func(e fg.Event) {
		count.SetText("1") // want `SetText call is not followed by ctx\.Update\(\)/ctx\.UpdateNow\(\) in this scope`
	})

	return btn
}

// present: the handler calls ctx.Update() right after the mutation — no
// diagnostic expected.
func present(ctx *fugo.Context) fg.Widget {
	count := fg.Text("0")
	btn := fg.Button("inc")
	btn.OnClick(func(e fg.Event) {
		count.SetText("1")
		ctx.Update()
	})

	return btn
}

// presentUpdateNow: UpdateNow also satisfies the check.
func presentUpdateNow(ctx *fugo.Context) fg.Widget {
	count := fg.Text("0")
	btn := fg.Button("inc")
	btn.OnClick(func(e fg.Event) {
		count.SetText("1")
		ctx.UpdateNow()
	})

	return btn
}

// initialSetupNoClosure: Set* calls made directly while building the tree
// (not inside an event handler) with an unrelated closure elsewhere in the
// same function that does call ctx.Update() — must NOT be flagged, since we
// can't tell this apart from legitimate one-time initialization without
// being overly aggressive.
func initialSetupNoClosure(ctx *fugo.Context) fg.Widget {
	count := fg.Text("0")
	count.SetText(strconv.Itoa(42))

	btn := fg.Button("inc")
	btn.OnClick(func(e fg.Event) {
		count.SetText("1")
		ctx.Update()
	})

	return btn
}

// windowSetterNotAWidget: ctx.Window().SetTitle(...) is a Set* call, but the
// receiver is a *fugo.WindowController, not an fg widget — it's an
// out-of-band command, not a widget-tree mutation, and must NOT be flagged
// even though no ctx.Update() follows it.
func windowSetterNotAWidget(ctx *fugo.Context) fg.Widget {
	btn := fg.Button("title")
	btn.OnClick(func(e fg.Event) {
		ctx.Window().SetTitle("new title")
	})

	return btn
}

// noCtxAvailable: no *fugo.Context is reachable at all — must stay silent.
func noCtxAvailable() fg.Widget {
	count := fg.Text("0")
	count.SetText("1")

	return count
}
