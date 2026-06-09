package fugo

import fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"

// Clipboard returns a controller for the OS clipboard, backed by the client.
func (c *Context) Clipboard() *Clipboard {
	return &Clipboard{app: c.app}
}

// Clipboard reads from and writes to the OS clipboard. Obtain one with
// Context.Clipboard. Reads are asynchronous: the result arrives in a callback
// that runs on the event goroutine, so it may mutate widgets and call Update
// just like a widget event handler.
type Clipboard struct{ app *App }

// Write places text on the OS clipboard. It is fire-and-forget.
func (cb *Clipboard) Write(text string) {
	cb.app.sendHost(&fugov1.HostCommand{
		Op:   fugov1.HostOp_HOST_CLIPBOARD_WRITE,
		Text: text,
	}, nil)
}

// Read fetches the clipboard's plain-text contents and delivers them to fn
// (empty if the clipboard holds no text).
func (cb *Clipboard) Read(fn func(text string)) {
	cb.app.sendHost(&fugov1.HostCommand{
		Op: fugov1.HostOp_HOST_CLIPBOARD_READ,
	}, func(data []byte) {
		fn(string(data))
	})
}

// FileDialog configures a native file open/save dialog.
type FileDialog struct {
	// Title is the dialog window title.
	Title string
	// DefaultName is the suggested file name, used by Save dialogs.
	DefaultName string
	// Extensions restricts selectable files to these extensions, written
	// without the leading dot (e.g. []string{"png", "jpg"}). Empty allows any
	// file.
	Extensions []string
}

// Files returns a controller for native file dialogs, backed by the client.
func (c *Context) Files() *FilePicker {
	return &FilePicker{app: c.app}
}

// FilePicker opens native open/save dialogs. Obtain one with Context.Files. The
// chosen path is delivered to a callback that runs on the event goroutine, so
// it may mutate widgets and call Update just like a widget event handler.
type FilePicker struct{ app *App }

// Open shows a native "open file" dialog and delivers the selected absolute
// path to fn, or "" if the user cancelled.
func (f *FilePicker) Open(dlg FileDialog, fn func(path string)) {
	f.app.sendHost(&fugov1.HostCommand{
		Op:         fugov1.HostOp_HOST_FILE_OPEN,
		Text:       dlg.Title,
		Extensions: dlg.Extensions,
	}, func(data []byte) { fn(string(data)) })
}

// Save shows a native "save file" dialog and delivers the chosen absolute path
// to fn, or "" if the user cancelled.
func (f *FilePicker) Save(dlg FileDialog, fn func(path string)) {
	f.app.sendHost(&fugov1.HostCommand{
		Op:          fugov1.HostOp_HOST_FILE_SAVE,
		Text:        dlg.Title,
		DefaultName: dlg.DefaultName,
		Extensions:  dlg.Extensions,
	}, func(data []byte) { fn(string(data)) })
}
