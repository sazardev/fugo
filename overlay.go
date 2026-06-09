package fugo

import fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"

// ShowSnackBar shows a brief Material snackbar with text at the bottom of the
// window. It is fire-and-forget — call it from an event handler.
func (c *Context) ShowSnackBar(text string) {
	c.app.sendOverlay(&fugov1.OverlayCommand{
		Op:      fugov1.OverlayOp_OVERLAY_SNACKBAR,
		Message: text,
	}, nil)
}

// ShowDialog shows a modal Material alert dialog with a title and message,
// dismissed by an OK button. It is fire-and-forget.
func (c *Context) ShowDialog(title, message string) {
	c.app.sendOverlay(&fugov1.OverlayCommand{
		Op:      fugov1.OverlayOp_OVERLAY_DIALOG,
		Title:   title,
		Message: message,
	}, nil)
}

// ShowBottomSheet shows a modal Material bottom sheet with a title and message.
// It is fire-and-forget.
func (c *Context) ShowBottomSheet(title, message string) {
	c.app.sendOverlay(&fugov1.OverlayCommand{
		Op:      fugov1.OverlayOp_OVERLAY_BOTTOMSHEET,
		Title:   title,
		Message: message,
	}, nil)
}

// PickDate opens the native Material date picker and delivers the chosen date as
// an ISO "YYYY-MM-DD" string to fn (empty if the user cancelled). The callback
// runs on the event goroutine, so it may mutate widgets and call Update.
func (c *Context) PickDate(fn func(date string)) {
	c.app.sendOverlay(&fugov1.OverlayCommand{
		Op: fugov1.OverlayOp_OVERLAY_DATE_PICKER,
	}, func(b []byte) { fn(string(b)) })
}

// PickTime opens the native Material time picker and delivers the chosen time as
// a 24-hour "HH:MM" string to fn (empty if the user cancelled). The callback
// runs on the event goroutine, so it may mutate widgets and call Update.
func (c *Context) PickTime(fn func(t string)) {
	c.app.sendOverlay(&fugov1.OverlayCommand{
		Op: fugov1.OverlayOp_OVERLAY_TIME_PICKER,
	}, func(b []byte) { fn(string(b)) })
}
