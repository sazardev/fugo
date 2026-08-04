#!/usr/bin/env bash
# xdotool helpers against the headless Xvfb display: window discovery,
# layout, and pointer input. Pure X11 protocol — no root/input-group
# needed (unlike the ydotool/uinput approach used by an earlier,
# desktop-visible version of this pipeline).
set -euo pipefail

x11_wait_for_class() {
	local needle="$1" timeout="${2:-30}"
	local waited=0
	while true; do
		local id
		id=$(xdotool search --class "$needle" 2>/dev/null | tail -1)
		[[ -n "$id" ]] && return 0
		sleep 0.5
		waited=$(python3 -c "print($waited + 0.5)")
		if (($(python3 -c "print(1 if $waited > $timeout else 0)"))); then
			echo "timed out waiting for window class containing '$needle'" >&2
			return 1
		fi
	done
}

x11_window_id() {
	xdotool search --class "$1" 2>/dev/null | tail -1
}

x11_place() {
	local class="$1" x="$2" y="$3" w="$4" h="$5"
	local id
	id=$(x11_window_id "$class")
	[[ -z "$id" ]] && return 1
	xdotool windowsize --sync "$id" "$w" "$h"
	xdotool windowmove --sync "$id" "$x" "$y"
	xdotool windowactivate "$id" >/dev/null 2>&1 || true
}

# Moves a window without resizing it — used for the Fugo app, which
# record-demo.sh gives its final full-screen size up front (fugo.toml +
# FUGO_WIDTH/FUGO_HEIGHT — see the comment in record-demo.sh's DEMO_RC
# about a real gap in cmd/fugo's hot-reload path) instead of resizing the
# live window, so it never needs a runtime resize here.
x11_move_only() {
	local class="$1" x="$2" y="$3"
	local id
	id=$(x11_window_id "$class")
	[[ -z "$id" ]] && return 1
	xdotool windowmove --sync "$id" "$x" "$y"
	xdotool windowactivate "$id" >/dev/null 2>&1 || true
}

# Splits the virtual screen into two side-by-side tiles: terminal on the
# left, the Fugo app window on the right.
x11_layout_split() {
	local term_class="$1" app_class="$2"
	local half=$((XVFB_WIDTH / 2))
	x11_place "$term_class" 0 0 "$half" "$XVFB_HEIGHT"
	x11_place "$app_class" "$half" 0 "$half" "$XVFB_HEIGHT"
}

# Fills the whole virtual screen with the terminal — safe to resize (see
# x11_move_only for why the app window uses move-only instead).
x11_fullscreen() {
	x11_place "$1" 0 0 "$XVFB_WIDTH" "$XVFB_HEIGHT"
}

# Pushes a window fully outside the root window's bounds — X11 simply does
# not composite anything positioned past the root's dimensions, so this
# hides it from the x11grab capture without unmapping it (unmapping a GTK/
# Flutter window can pause its renderer; this keeps it live so hot-reload
# updates still land while it's "offscreen").
x11_park_offscreen() {
	local id
	id=$(x11_window_id "$1")
	[[ -z "$id" ]] && return 0
	xdotool windowmove "$id" "$XVFB_WIDTH" 0
}

x11_click() {
	local x="$1" y="$2"
	DISPLAY="$XVFB_DISPLAY" xdotool mousemove --sync "$x" "$y" click 1
}

x11_move_cursor() {
	DISPLAY="$XVFB_DISPLAY" xdotool mousemove --sync "$1" "$2"
}
