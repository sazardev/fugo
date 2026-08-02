package lsp

import (
	"net/url"
	"strings"
)

// FilePath converts a "file://" URI as sent by an LSP client into a local
// filesystem path.
//
// This is a reasonably careful but not exhaustive implementation: primary
// development target is Linux/macOS, where "file:///a/b/c" simply maps to
// "/a/b/c". Windows paths are handled on a best-effort basis: a URI like
// "file:///C:/Users/x" (three slashes, drive letter) maps to "C:/Users/x".
// Percent-encoded characters (e.g. spaces as %20) are decoded. If uri does
// not have a "file" scheme, it is returned unchanged as a fallback.
func FilePath(uri string) string {
	u, err := url.Parse(uri)
	if err != nil {
		return uri
	}
	if u.Scheme != "" && u.Scheme != "file" {
		return uri
	}

	path := u.Path
	if path == "" {
		path = u.Opaque
	}

	// Windows: "file:///C:/Users/x" parses to Path == "/C:/Users/x"; strip
	// the leading slash before the drive letter.
	if len(path) >= 3 && path[0] == '/' && path[2] == ':' {
		path = path[1:]
	}

	return path
}

// FileURI converts a local filesystem path into a "file://" URI suitable
// for sending to an LSP client.
//
// As with FilePath, Windows drive-letter paths ("C:\Users\x" or
// "C:/Users/x") are handled by normalizing backslashes to forward slashes
// and producing "file:///C:/Users/x"; this is not an exhaustive UNC-path
// implementation.
func FileURI(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")

	if len(path) >= 2 && path[1] == ':' {
		// Windows drive-letter absolute path, e.g. "C:/Users/x".
		path = "/" + path
	}

	u := url.URL{Scheme: "file", Path: path}

	return u.String()
}
