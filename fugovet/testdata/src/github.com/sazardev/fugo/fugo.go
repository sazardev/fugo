// Package fugo is a minimal stand-in for github.com/sazardev/fugo, just
// enough to exercise fugovet's type-based detection of *fugo.Context.
package fugo

// Context mirrors the shape of the real fugo.Context enough for testing:
// Update/UpdateNow to mark the retained tree dirty.
type Context struct{}

// Update marks the app dirty for the next scheduler tick.
func (c *Context) Update() {}

// UpdateNow marks the app dirty and wakes the scheduler immediately.
func (c *Context) UpdateNow() {}

// NavigateTo is unused by fugovet but kept for a realistic surface.
func (c *Context) NavigateTo(route string) {}

// Window returns the runtime window controller — out-of-band from the
// widget tree, so its Set* methods must NOT be treated like widget setters.
func (c *Context) Window() *WindowController { return &WindowController{} }

// WindowController is a minimal stand-in for fugo.WindowController.
type WindowController struct{}

// SetTitle sends a window command directly; it has nothing to do with the
// retained widget tree and needs no ctx.Update().
func (w *WindowController) SetTitle(title string) {}
