// Package engine is Fugo's render core. [Diff] turns two widget trees into a
// minimal list of patches; [Reconciler] streams those patches to the client and
// buffers them until one connects; [Scheduler] coalesces updates into a single
// flush per frame (~16ms / 60fps).
package engine
