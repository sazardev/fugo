#!/usr/bin/env bash
# Manages a virtual X11 display (Xvfb) so the whole demo runs fully
# headless — nothing is ever shown on the real desktop. ffmpeg's x11grab
# records this virtual framebuffer directly.
#
# Gotcha found empirically: alacritty and the Flutter/GTK client both
# prefer Wayland when $WAYLAND_DISPLAY is set, even if $DISPLAY also points
# at Xvfb — they'll silently try to connect to the REAL compositor instead
# and never map a window on the virtual display. WAYLAND_DISPLAY must be
# unset for every child process launched against Xvfb.
set -euo pipefail

XVFB_DISPLAY="${XVFB_DISPLAY:-:97}"
XVFB_WIDTH="${XVFB_WIDTH:-1920}"
XVFB_HEIGHT="${XVFB_HEIGHT:-1080}"
XVFB_PID_FILE="/tmp/fugo-demo-xvfb.pid"

xvfb_start() {
	pkill -f "Xvfb $XVFB_DISPLAY " 2>/dev/null || true
	sleep 0.3
	Xvfb "$XVFB_DISPLAY" -screen 0 "${XVFB_WIDTH}x${XVFB_HEIGHT}x24" -nolisten tcp >/tmp/fugo-demo-xvfb.log 2>&1 &
	echo "$!" >"$XVFB_PID_FILE"
	disown

	local tries=0
	local up=1
	for tries in $(seq 1 50); do
		if DISPLAY="$XVFB_DISPLAY" xdotool getdisplaygeometry >/dev/null 2>&1; then
			up=0
			break
		fi
		sleep 0.2
	done
	if [[ "$up" != "0" ]]; then
		echo "Xvfb never came up on $XVFB_DISPLAY" >&2
		return 1
	fi
}

xvfb_stop() {
	if [[ -f "$XVFB_PID_FILE" ]]; then
		kill "$(<"$XVFB_PID_FILE")" 2>/dev/null || true
		rm -f "$XVFB_PID_FILE"
	fi
	pkill -f "Xvfb $XVFB_DISPLAY " 2>/dev/null || true
}
