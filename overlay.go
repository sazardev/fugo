package fugo

import fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"

// ShowSnackBar shows a brief Material snackbar with text at the bottom of the
// window. It is fire-and-forget — call it from an event handler.
func (c *Context) ShowSnackBar(text string) {
	c.app.reconciler.SendOverlayCommand(&fugov1.OverlayCommand{
		Op:      fugov1.OverlayOp_OVERLAY_SNACKBAR,
		Message: text,
	})
}

// ShowDialog shows a modal Material alert dialog with a title and message,
// dismissed by an OK button. It is fire-and-forget.
func (c *Context) ShowDialog(title, message string) {
	c.app.reconciler.SendOverlayCommand(&fugov1.OverlayCommand{
		Op:      fugov1.OverlayOp_OVERLAY_DIALOG,
		Title:   title,
		Message: message,
	})
}
