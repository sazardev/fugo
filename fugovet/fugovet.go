// Package fugovet implements static analysis checks for Fugo apps — Go
// programs built on top of github.com/sazardev/fugo and its declarative
// widget API github.com/sazardev/fugo/fg.
//
// The checks encode conventions from the Fugo mental model (see the
// project's CLAUDE.md, "Key mental model"): the widget tree is built once
// and retained; event handlers mutate widget structs in place and must call
// ctx.Update()/ctx.UpdateNow() for the mutation to be rendered; and button
// widgets are useless without an OnClick handler.
//
// Analyzers is the full set, suitable for passing to
// golang.org/x/tools/go/analysis/multichecker.Main so the resulting binary
// works as a go vet tool (-vettool=...).
package fugovet

import "golang.org/x/tools/go/analysis"

// Analyzers is the complete set of fugovet checks.
var Analyzers = []*analysis.Analyzer{
	UpdateMissing,
	HandlerMissing,
	UIImport,
}
