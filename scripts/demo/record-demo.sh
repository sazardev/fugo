#!/usr/bin/env bash
# record-demo.sh — records a real, live "0 to running app" Fugo demo, fully
# headless: `fugo init` -> `fugo doctor` -> `fugo run` (a real Flutter
# window spawns, on a virtual X11 display nobody sees) -> a live code edit
# that hot-reloads the running app -> `fugo build`.
#
# Nothing touches the real desktop: everything runs against a headless
# Xvfb display (see lib/xvfb.sh) and is captured straight off that virtual
# framebuffer via ffmpeg's x11grab. The only "trick" is pacing (typing
# speed) and a scripted digital zoom applied in postprocess.sh, timed from
# a beat log this script writes while it runs (see lib/gen_zoom_filter.py)
# — Xvfb has no compositor to do a live zoom the way a real desktop would.
#
# Requires: Xvfb, xdotool, tmux, alacritty, python3, ffmpeg.
# (ffmpeg is only needed by postprocess.sh, not this script.)
#
# Usage: scripts/demo/record-demo.sh [output.mp4]
set -euo pipefail

DEMO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$DEMO_DIR/../.." && pwd)"

# shellcheck source=lib/xvfb.sh
source "$DEMO_DIR/lib/xvfb.sh"
# shellcheck source=lib/x11.sh
source "$DEMO_DIR/lib/x11.sh"
# shellcheck source=lib/typewriter.sh
source "$DEMO_DIR/lib/typewriter.sh"
# shellcheck source=assets/hot_reload_edit.env
source "$DEMO_DIR/assets/hot_reload_edit.env"

APP_NAME="${APP_NAME:-nimbus}"
SCRATCH_ROOT="${SCRATCH_ROOT:-/tmp/fugo-demo-workspace}"
PROJECT_DIR="$SCRATCH_ROOT/$APP_NAME"
TMUX_SESSION="fugodemo"
TERM_CLASS="fugodemo-term"
APP_CLASS="fugo_flutter_client" # substring of the GTK app-id class,
# e.g. "com.sazardev.fugo.fugo_flutter_client" — the window TITLE is the
# project's own fugo.toml title (e.g. "nimbus"), not the binary name.

FUGO_BIN="$REPO_DIR/bin/fugo.exe"
FLUTTER_BUNDLE="$REPO_DIR/flutter_client/build/linux/x64/release/bundle/fugo_flutter_client"

OUT_DIR="$DEMO_DIR/out"
mkdir -p "$OUT_DIR"
OUTPUT="${1:-$OUT_DIR/fugo-demo-raw-$(date +%Y%m%d-%H%M%S).mp4}"
TIMELINE="${OUTPUT%.mp4}.timeline.tsv"
: >"$TIMELINE"

# Overridable pacing knobs (smoke tests pass FAST=1 to shrink every sleep).
FAST="${FAST:-0}"
p() { # p <seconds> — a pacing sleep, scaled down under FAST=1
	if [[ "$FAST" == "1" ]]; then
		sleep "$(python3 -c "print(max(0.2, $1/4))")"
	else
		sleep "$1"
	fi
}

REQUIRED_CMDS=(Xvfb xdotool tmux alacritty python3)
for c in "${REQUIRED_CMDS[@]}"; do
	command -v "$c" >/dev/null 2>&1 || {
		echo "missing required command: $c" >&2
		exit 1
	}
done
[[ -x "$FUGO_BIN" ]] || {
	echo "fugo binary not found at $FUGO_BIN — run 'make cli' first" >&2
	exit 1
}
[[ -x "$FLUTTER_BUNDLE" ]] || {
	echo "Flutter client not built at $FLUTTER_BUNDLE — run 'make flutter-build' first" >&2
	exit 1
}

REC_START=""
FFMPEG_PID=""
# zoom_beat cx cy peak [hold] [in] [out] — logs a zoom-pulse for
# postprocess.sh, then sleeps through it so the on-screen content it will
# zoom into is actually held still for that long in the recording.
zoom_beat() {
	local cx="$1" cy="$2" peak="$3" hold="${4:-1.0}" in="${5:-0.7}" out="${6:-0.7}"
	local elapsed
	elapsed=$(python3 -c "import time; print(time.time() - $REC_START)")
	echo "$elapsed $in $hold $out $cx $cy $peak" >>"$TIMELINE"
	p "$(python3 -c "print($in + $hold + $out)")"
}

cleanup() {
	local ec=$?
	[[ -n "$FFMPEG_PID" ]] && kill -INT "$FFMPEG_PID" 2>/dev/null || true
	sleep 1
	tmux_kill_session "$TMUX_SESSION"
	pkill -f "alacritty --class $TERM_CLASS" 2>/dev/null || true
	xvfb_stop
	exit "$ec"
}
trap cleanup EXIT INT TERM

echo "==> preparing scratch project dir"
rm -rf "$SCRATCH_ROOT"
mkdir -p "$SCRATCH_ROOT"

echo "==> starting headless Xvfb on $XVFB_DISPLAY"
xvfb_start
export DISPLAY="$XVFB_DISPLAY"
unset WAYLAND_DISPLAY # see lib/xvfb.sh — critical, or apps connect to the real session

echo "==> launching terminal (headless, on $XVFB_DISPLAY)"
# A plain, self-contained bash — not the user's login shell — so a
# first-run Powerlevel10k/oh-my-zsh wizard can never steal the recorded
# keystrokes. All env the demo needs is set right here, explicitly.
DEMO_RC="$SCRATCH_ROOT/.demo_bashrc"
cat >"$DEMO_RC" <<RC
export DISPLAY='$XVFB_DISPLAY'
unset WAYLAND_DISPLAY
export FUGO_FLUTTER_BINARY="$FLUTTER_BUNDLE"
export PATH="$REPO_DIR/bin:\$PATH"
alias fugo="$FUGO_BIN"
export PS1='demo ~$ '
RC

# tmux's own default-shell (whatever the account's login shell is) is what
# actually runs inside the pane, so the clean bash must be its explicit
# shell-command argument — not just how we invoke tmux itself.
alacritty --config-file "$DEMO_DIR/assets/alacritty-ember.toml" --class "$TERM_CLASS" \
	-o 'window.position.x=0' -o 'window.position.y=0' \
	-e tmux -f "$DEMO_DIR/assets/tmux-ember.conf" new-session -s "$TMUX_SESSION" -x 100 -y 45 -c "$SCRATCH_ROOT" \
	"bash --rcfile '$DEMO_RC' -i" &

tries=0
until tmux has-session -t "$TMUX_SESSION" 2>/dev/null; do
	sleep 0.2
	tries=$((tries + 1))
	((tries > 50)) && {
		echo "tmux session never came up" >&2
		exit 1
	}
done
x11_wait_for_class "$TERM_CLASS" 15
x11_place "$TERM_CLASS" 0 0 $((XVFB_WIDTH / 2)) "$XVFB_HEIGHT"
# Any later split-window (the nano-editing beat) must also get the clean
# bash, not tmux's configured default-shell.
tmux set-option -t "$TMUX_SESSION" default-command "bash --rcfile '$DEMO_RC' -i"
sleep 0.5

echo "==> recording $XVFB_DISPLAY -> $OUTPUT"
ffmpeg -y -f x11grab -video_size "${XVFB_WIDTH}x${XVFB_HEIGHT}" -framerate 30 -i "$DISPLAY" \
	-c:v libx264 -preset superfast -crf 20 -pix_fmt yuv420p "$OUTPUT" >/tmp/fugo-demo-ffmpeg.log 2>&1 &
FFMPEG_PID=$!
REC_START=$(python3 -c "import time; print(time.time())")
sleep 1

T="$TMUX_SESSION"

# --- Beat 1: fugo init -------------------------------------------------
tmux_type_line "$T" "fugo init $APP_NAME -t app --theme dark -y" 1.2
p 1.0

# Swap in the Fugo brand theme (matches site/styles.css's ember palette)
# before anything else touches the project — silent, instant, so the app
# opens already on-brand instead of the template's default blue/purple.
sed "s/__MODULE__/$APP_NAME/g" "$DEMO_DIR/assets/branded_main.go.tmpl" >"$PROJECT_DIR/main.go"

tmux_type_line "$T" "cd $APP_NAME" 0.4

# --- Beat 2: fugo doctor -------------------------------------------------
tmux_type_line "$T" "fugo doctor" 1.5
p 1.2
zoom_beat 480 300 1.6

# --- Beat 3: fugo run — spawns the real Flutter window (headless) -------
tmux_type_line "$T" "fugo run" 0.3
x11_wait_for_class "$APP_CLASS" 90 || true
p 1.0
x11_layout_split "$TERM_CLASS" "$APP_CLASS"
p 0.5

APPX=$((XVFB_WIDTH * 3 / 4))
APPY=$((XVFB_HEIGHT / 2))
zoom_beat "$APPX" "$APPY" 1.7 1.4

# --- Beat 4: live edit -> hot reload -------------------------------------
# tmux's split-window does not reliably inherit the split-from pane's LIVE
# cwd (it's still sitting on the session's original start dir), so `cd`
# explicitly before opening the editor.
tmux_split_below "$T:0.0" >/dev/null
tmux_type_line "$T:0.1" "cd $PROJECT_DIR && nano ui/home.go" 1.0
tmux_key "$T:0.1" "C-w"
sleep 0.3
tmux_type "$T:0.1" "$SEARCH_TERM"
tmux_key "$T:0.1" Enter
p 0.6
tmux_key "$T:0.1" Home
tmux_key "$T:0.1" "C-k"
tmux_type "$T:0.1" "$NEW_LINE"
p 0.3
tmux_key "$T:0.1" Enter
tmux_key "$T:0.1" "C-o"
tmux_key "$T:0.1" Enter
p 0.3
tmux_key "$T:0.1" "C-x"

p 2.0 # let the hot reload rebuild + push the patch
zoom_beat "$APPX" "$APPY" 1.8 1.6

# --- Beat 5: navigate the router (About page) — best-effort click -------
x11_click "$APPX" "$((APPY + 180))" || true
p 1.0
zoom_beat "$APPX" "$((APPY + 80))" 1.5 1.0

# --- Beat 6: fugo build (ship it) ----------------------------------------
tmux select-pane -t "$T:0.0"
tmux_key "$T:0.0" C-c
p 1.0
tmux_type_line "$T:0.0" "fugo build" 1.5
p 1.5

# --- Outro: hold a beat before cutting ------------------------------------
p 1.5

echo "==> stopping recorder"
kill -INT "$FFMPEG_PID"
wait "$FFMPEG_PID" 2>/dev/null || true
FFMPEG_PID=""

echo "==> raw recording saved to $OUTPUT"
echo "==> zoom timeline saved to $TIMELINE"
