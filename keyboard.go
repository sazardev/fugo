package fugo

import (
	"strconv"
	"strings"

	"github.com/sazardev/fugo/fg"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
)

// RegisterShortcuts replaces the app-wide set of keyboard shortcuts the
// client watches for — the missing piece for desktop basics like Ctrl+S,
// Ctrl+Z, or Escape, which the widget tree alone has no way to express.
// Bindings are lowercase, "+"-joined modifier+key strings, e.g. "ctrl+s",
// "ctrl+shift+z", "escape" (modifiers always in the order ctrl, alt, shift,
// meta, matching how the client normalizes a physical key event before
// matching — see flutter_client's shortcut listener).
//
// Calling this again replaces the previous set entirely; it is not additive.
// Handlers run on the event goroutine, so they may mutate widgets and call
// Context.Update, exactly like a widget event handler.
func (c *Context) RegisterShortcuts(bindings map[string]func()) {
	c.app.shortcutsMu.Lock()
	c.app.shortcuts = bindings
	c.app.shortcutsMu.Unlock()

	names := make([]string, 0, len(bindings))
	for k := range bindings {
		names = append(names, k)
	}

	if c.app.reconciler != nil {
		c.app.reconciler.SendShortcutsCommand(&fugov1.ShortcutsCommand{Bindings: names})
	}
}

// OnResize registers a callback invoked whenever the client reports the
// window has been resized, letting the app rebuild its tree conditionally for
// the new size — there is no other channel back for Go to learn the client's
// viewport dimensions (the widget tree is built before anything is measured).
// Calling this again replaces the previous callback; passing nil disables it.
// It runs on the event goroutine, so it may mutate widgets and call
// Context.Update.
func (c *Context) OnResize(fn func(width, height float64)) {
	c.app.resizeMu.Lock()
	c.app.resizeHandler = fn
	c.app.resizeMu.Unlock()
}

// OnFileDrop registers a callback invoked whenever the user drags files from
// the OS onto the client window — the one channel back for Go to learn about
// a drag-and-drop that originates outside the app. Calling this again
// replaces the previous callback; passing nil disables it. It runs on the
// event goroutine, so it may mutate widgets and call Context.Update.
func (c *Context) OnFileDrop(fn func(paths []string)) {
	c.app.fileDropMu.Lock()
	c.app.fileDropHandler = fn
	c.app.fileDropMu.Unlock()
}

// RequestFocus asks the client to move keyboard focus to w (today, only
// TextField supports receiving focus client-side). Useful after a validation
// error to jump back to the offending field. Fire-and-forget — there is no
// reply. w must already be part of the built tree (its node id is 0, and the
// request silently ignored, until the first render pass has run).
func (c *Context) RequestFocus(w fg.Focusable) {
	if c.app.reconciler == nil {
		return
	}

	c.app.reconciler.SendFocusCommand(&fugov1.FocusCommand{NodeId: w.NodeID()})
}

// dispatchShortcut looks up and invokes the handler registered for binding,
// if any.
func (a *App) dispatchShortcut(binding string) {
	a.shortcutsMu.RLock()
	fn, ok := a.shortcuts[binding]
	a.shortcutsMu.RUnlock()

	if ok && fn != nil {
		fn()
	}
}

// dispatchResize parses a "WIDTHxHEIGHT" payload and invokes the registered
// resize callback, if any. Malformed payloads are ignored.
func (a *App) dispatchResize(data string) {
	w, h, ok := parseSize(data)
	if !ok {
		return
	}

	a.resizeMu.Lock()
	fn := a.resizeHandler
	a.resizeMu.Unlock()

	if fn != nil {
		fn(w, h)
	}
}

// dispatchFileDrop splits data (paths joined by "\n") and invokes the
// registered file-drop callback, if any. Empty input yields no call.
func (a *App) dispatchFileDrop(data string) {
	if data == "" {
		return
	}

	a.fileDropMu.Lock()
	fn := a.fileDropHandler
	a.fileDropMu.Unlock()

	if fn != nil {
		fn(strings.Split(data, "\n"))
	}
}

func parseSize(s string) (float64, float64, bool) {
	parts := strings.SplitN(s, "x", 2)
	if len(parts) != 2 {
		return 0, 0, false
	}

	w, errW := strconv.ParseFloat(parts[0], 64)
	h, errH := strconv.ParseFloat(parts[1], 64)
	if errW != nil || errH != nil {
		return 0, 0, false
	}

	return w, h, true
}
