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
	xdotool windowsize "$id" "$w" "$h"
	xdotool windowmove "$id" "$x" "$y"
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

x11_click() {
	local x="$1" y="$2"
	DISPLAY="$XVFB_DISPLAY" xdotool mousemove --sync "$x" "$y" click 1
}

x11_move_cursor() {
	DISPLAY="$XVFB_DISPLAY" xdotool mousemove --sync "$1" "$2"
}
