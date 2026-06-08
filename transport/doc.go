// Package transport hosts Fugo's gRPC server. It serves the bidirectional
// render stream over a Unix domain socket (or TCP on Windows), with a standard
// health check and keepalive.
package transport
