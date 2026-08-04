#!/usr/bin/env bash
# Human-like typing into a live tmux pane via `tmux send-keys -l` (literal
# text). This drives a REAL shell — every command actually executes — the
# only thing simulated is the pacing of keystrokes appearing on screen.
set -euo pipefail

# Base per-character delay in seconds; overridable per-call for pacing.
TYPE_BASE_DELAY="${TYPE_BASE_DELAY:-0.028}"
TYPE_JITTER="${TYPE_JITTER:-0.035}"

_type_sleep() {
	python3 -c "import random; print($TYPE_BASE_DELAY + random.random()*$TYPE_JITTER)"
}

# tmux_type <session:window.pane> <text>
# Sends text one character at a time with jittered delays, no trailing Enter.
tmux_type() {
	local target="$1" text="$2"
	local i char
	for ((i = 0; i < ${#text}; i++)); do
		char="${text:i:1}"
		tmux send-keys -t "$target" -l -- "$char"
		sleep "$(_type_sleep)"
	done
}

# tmux_type_line <target> <text> [post_sleep]
# Types the line, presses Enter, then waits post_sleep (default 0.6s) so the
# command's own output has time to appear before the next beat starts.
tmux_type_line() {
	local target="$1" text="$2" post="${3:-0.6}"
	tmux_type "$target" "$text"
	tmux send-keys -t "$target" Enter
	sleep "$post"
}

# tmux_pause <seconds> — a beat with nothing happening, for pacing.
tmux_pause() {
	sleep "$1"
}

# tmux_new_session <name> <cwd> <command>
# Starts a detached tmux session running $command, sized for a 1280x1000-ish
# terminal tile (adjust with -x/-y if the layout changes).
tmux_new_session() {
	local name="$1" cwd="$2" cmd="$3"
	tmux new-session -d -s "$name" -x 220 -y 55 -c "$cwd" "$cmd"
}

tmux_kill_session() {
	tmux kill-session -t "$1" 2>/dev/null || true
}

# tmux_key <target> <key...> — send named keys (Enter, C-o, C-x, Home, ...).
tmux_key() {
	local target="$1"
	shift
	tmux send-keys -t "$target" "$@"
}

# tmux_split_below <target> — splits the pane vertically (new pane below),
# returns the new pane's index on stdout so callers can target it.
tmux_split_below() {
	local target="$1"
	tmux split-window -v -t "$target"
	tmux display-message -t "$target" -p '#{window_id}'
}
