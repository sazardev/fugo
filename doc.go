// Package fugo is the root of the Fugo framework: a local Server-Driven UI
// (SDUI) toolkit for desktop apps where all logic, state, and routing live in
// Go and a precompiled Flutter binary renders the UI.
//
// The two processes talk over gRPC bidirectional streaming — Unix domain
// sockets, or TCP on Windows — using standard Protocol Buffers. Go is the single
// source of truth; Flutter holds no business logic or state.
//
// # How it works
//
// You build a retained widget tree once in buildUI. Event handlers are Go
// closures that mutate widget fields in place and call [Context.Update]. A
// 60fps scheduler re-walks the retained tree, diffs it against the previous
// snapshot, and streams only the resulting patches to the client.
//
// # Packages
//
//   - fugo (this package): [App], [Context], and the application lifecycle.
//   - fg: the declarative widget API (fg.Text, fg.Button, ...) and the Theme system.
//   - style: styling primitives (Color, EdgeInsets, TextStyle, Border).
//   - engine: the diffing engine, reconciler, and frame scheduler.
//   - transport: the gRPC server (UDS/TCP) with health check and keepalive.
//   - supervisor: spawns and monitors the Flutter render subprocess.
//
// # Getting started
//
//	func main() {
//		fugo.RunStandalone(fugo.AppOptions{Title: "Hello", Width: 800, Height: 600}, buildUI)
//	}
//
//	func buildUI(ctx *fugo.Context) fg.Widget {
//		return fg.Center(fg.Text("Hello, Fugo!").FontSize(24))
//	}
package fugo
