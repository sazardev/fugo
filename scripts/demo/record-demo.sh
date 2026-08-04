#!/usr/bin/env bash
# record-demo.sh — records a real, live "0 to running app" Fugo demo, fully
# headless: `fugo init` -> a quick scaffold tour -> `fugo doctor` ->
# `fugo run` (a real Flutter window spawns, on a virtual X11 display
# nobody sees) -> three separate live nvim edits, each hot-reloading the
# running app -> `fugo build`.
#
# Nothing touches the real desktop: everything runs against a headless
# Xvfb display (see lib/xvfb.sh) and is captured straight off that virtual
# framebuffer via ffmpeg's x11grab. The terminal is YOUR real alacritty +
# nvim config (no theming override) — only the app itself carries Fugo's
# own brand theme (assets/branded_main.go.tmpl) and a hand-written
# mobile-style scaffold (assets/mobile_home.go.tmpl: an app bar, a task
# list, a FAB, a bottom nav with three tabs) instead of the `app`
# template's bare counter — a portrait window sized like a phone, not a
# landscape desktop window, is the whole point of the demo. The terminal
# stays mapped full-screen underneath the whole time and doubles as a dark
# backdrop; "switching" between "editing code" and "the app, live" is just
# raising whichever window should be on top (x11_raise in lib/x11.sh) —
# more dramatic than a permanently shared split screen. The other "trick"
# is a scripted digital zoom applied in postprocess.sh, timed from a beat
# log this script writes while it runs (see lib/gen_zoom_filter.py) — Xvfb
# has no compositor to do a live zoom the way a real desktop would.
#
# Requires: Xvfb, xdotool, tmux, alacritty, nvim, python3, ffmpeg.
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

REQUIRED_CMDS=(Xvfb xdotool tmux alacritty nvim python3)
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

# show_app / show_term — the full-screen-terminal-as-backdrop alternation
# described above. The app is only ever MOVED, never resized (see
# x11_move_only in lib/x11.sh) — it's already given its final portrait
# size up front, in fugo.toml, before `fugo run` ever launches it.
show_app() {
	x11_move_only "$APP_CLASS" "$APPX" "$APPY"
	x11_raise "$APP_CLASS"
}
show_term() {
	x11_raise "$TERM_CLASS"
}

# nvim_replace_line <pane> <search> <new_line_with_its_own_leading_tabs>
# Searches for the line, jumps to column 0, then `C` (change-to-end-of-line
# — NOT `cc`/`S`, which would re-apply autoindent and double up the
# replacement's own leading tabs) so the result is correct regardless of
# the machine's nvim indent settings.
nvim_replace_line() {
	local target="$1" search="$2" new_line="$3"
	tmux_key "$target" "/"
	tmux_type "$target" "$search"
	tmux_key "$target" Enter
	p 0.5
	tmux_key "$target" "0"
	tmux_key "$target" "C"
	tmux_type "$target" "$new_line"
	tmux_key "$target" Escape
}
nvim_save() { tmux_type_line "$1" ":w" 0.8; }
nvim_quit() { tmux_type_line "$1" ":wq" 0.8; }

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

# The app's portrait, phone-like window — centered on the canvas, over the
# full-screen terminal underneath. Coordinates for the FAB and the bottom
# nav's Stats/Profile destinations are derived from this same geometry
# (verified empirically against a real render at this exact size).
APPW=420
APPH=900
APPX=$(((XVFB_WIDTH - APPW) / 2))
APPY=$(((XVFB_HEIGHT - APPH) / 2))
NAVY=$((APPY + APPH - 37))
TASKSX=$((APPX + APPW / 6))
STATSX=$((APPX + APPW / 2))
PROFILEX=$((APPX + APPW * 5 / 6))
FABX=$((APPX + APPW - 45))
FABY=$((APPY + APPH - 124))

echo "==> launching terminal (headless, on $XVFB_DISPLAY)"
# A plain, self-contained bash — not the user's login shell — so a
# first-run Powerlevel10k/oh-my-zsh wizard can never steal the recorded
# keystrokes. All env the demo needs is set right here, explicitly. The
# terminal/editor look, though, is deliberately the real thing: no
# --config-file override on alacritty, no -f override on tmux, plain nvim
# with whatever colorscheme/LSP/treesitter is already configured.
DEMO_RC="$SCRATCH_ROOT/.demo_bashrc"
cat >"$DEMO_RC" <<RC
export DISPLAY='$XVFB_DISPLAY'
unset WAYLAND_DISPLAY
export FUGO_FLUTTER_BINARY="$FLUTTER_BUNDLE"
export PATH="$REPO_DIR/bin:\$PATH"
alias fugo="$FUGO_BIN"
export PS1='demo ~$ '
# Workaround for a real Fugo gap, not a demo-only hack: hot-reload mode
# (cmd/fugo's startFlutterClient — the default \`fugo run\` path) spawns the
# Flutter client itself and never forwards the window/theme env vars that
# app.go's exportWindowEnv sets (FUGO_WIDTH/HEIGHT/TITLE, and — same bug —
# FUGO_THEME_SEED/BRIGHTNESS), so the window falls back to main.dart's
# 800x600 default and Material's light scheme regardless of fugo.toml /
# fg.UseTheme. Exporting them here works because that spawn does
# \`cmd.Env = append(os.Environ(), ...)\` — it inherits whatever's already
# in the shell that ran \`fugo run\`.
export FUGO_WIDTH="$APPW"
export FUGO_HEIGHT="$APPH"
export FUGO_TITLE="$APP_NAME"
export FUGO_THEME_SEED="#FF6A1A"
export FUGO_THEME_BRIGHTNESS="dark"
RC

alacritty --class "$TERM_CLASS" \
	-e tmux new-session -s "$TMUX_SESSION" -x 220 -y 55 -c "$SCRATCH_ROOT" \
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
x11_fullscreen "$TERM_CLASS"
# Any later split-window (the nvim-editing beat) must also get the clean
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
MIDX=$((XVFB_WIDTH / 2))

# --- Beat 1: fugo init ----------------------------------------------------
# --no-git: `fugo init`'s default git init + initial commit picks up
# whatever global git/GPG config is on this machine — if commit signing is
# on, that pops a real pinentry dialog (with the machine owner's real
# key/email) that blocks the recording and would leak into the video.
tmux_type_line "$T" "fugo init $APP_NAME --theme dark --no-git -y" 1.2
p 1.0

# Swap in the Fugo brand theme (matches site/styles.css's ember palette)
# and a hand-written mobile-style scaffold (assets/mobile_home.go.tmpl)
# before anything else touches the project — silent, instant, so the app
# opens already looking like a real built app instead of the default
# template's bare counter-and-two-buttons starter.
sed "s/__MODULE__/$APP_NAME/g" "$DEMO_DIR/assets/branded_main.go.tmpl" >"$PROJECT_DIR/main.go"
cp "$DEMO_DIR/assets/mobile_home.go.tmpl" "$PROJECT_DIR/ui/home.go"

# Give the app its final portrait, phone-like size up front, in
# fugo.toml, instead of resizing the live GTK/Flutter window later — see
# x11_move_only in lib/x11.sh for why a runtime resize of the APP (unlike
# the terminal) is unreliable under Xvfb.
sed -i \
	-e "s/^width  = .*/width  = $APPW/" \
	-e "s/^height = .*/height = $APPH/" \
	"$PROJECT_DIR/fugo.toml"

tmux_type_line "$T" "cd $APP_NAME" 0.4

# --- Beat 2: a quick scaffold tour -----------------------------------------
tmux_type_line "$T" "ls" 0.8
p 0.6
tmux_type_line "$T" "cat fugo.toml" 1.0
p 0.8
zoom_beat "$MIDX" 350 1.5 0.9

# --- Beat 3: fugo doctor ---------------------------------------------------
tmux_type_line "$T" "fugo doctor" 1.5
p 1.2
zoom_beat "$MIDX" 400 1.6

# --- Beat 4: fugo run — spawns the real Flutter window (headless) --------
tmux_type_line "$T" "fugo run" 0.3
x11_wait_for_class "$APP_CLASS" 90 || true
p 1.5 # let Flutter finish its first real frame before revealing it
show_app
zoom_beat "$((APPX + APPW / 2))" "$((APPY + 200))" 1.5 1.3

# --- Beat 5: open nvim alongside the running server -----------------------
show_term
tmux_split_below "$T:0.0" >/dev/null
# tmux's split-window does not reliably inherit the split-from pane's LIVE
# cwd (it's still sitting on the session's original start dir), so `cd`
# explicitly before opening the editor.
tmux_type_line "$T:0.1" "cd $PROJECT_DIR && nvim ui/home.go" 2.2
zoom_beat "$MIDX" 500 1.4 1.0

# --- Edit 1: a 4th task appears in the list --------------------------------
nvim_replace_line "$T:0.1" "$EDIT1_SEARCH" "$EDIT1_NEW"
nvim_save "$T:0.1"
p 1.6 # let the hot reload rebuild + push the patch

show_app
zoom_beat "$((APPX + APPW / 2))" "$((APPY + 260))" 1.6 1.5

# --- Edit 2: a real logic change — the FAB's new-task label --------------
show_term
nvim_replace_line "$T:0.1" "$EDIT2_SEARCH" "$EDIT2_NEW"
nvim_save "$T:0.1"
p 1.6

show_app
# Best-effort click: coordinates are derived from the portrait window's
# own known geometry (APPX/APPY/APPW/APPH), not queried live — if the
# exact Material layout ever shifts, this just misses the FAB harmlessly.
x11_click "$FABX" "$FABY" || true
p 0.8
zoom_beat "$((APPX + APPW / 2))" "$((APPY + 340))" 1.7 1.4

# --- Quick tour: the Stats tab (no edit, just showing off the scaffold) --
x11_click "$STATSX" "$NAVY" || true
p 0.8
zoom_beat "$((APPX + APPW / 2))" "$((APPY + APPH / 2))" 1.6 1.2

# --- Edit 3: restyle the Profile tab's tagline -----------------------------
show_term
nvim_replace_line "$T:0.1" "$EDIT3_SEARCH" "$EDIT3_NEW"
nvim_save "$T:0.1"
p 1.0
nvim_quit "$T:0.1"
p 1.5

show_app
x11_click "$PROFILEX" "$NAVY" || true
p 1.0
zoom_beat "$((APPX + APPW / 2))" "$((APPY + APPH / 2))" 1.7 1.4

# --- Beat 6: fugo build (ship it) ------------------------------------------
show_term
tmux select-pane -t "$T:0.0"
tmux_key "$T:0.0" C-c
p 1.0
tmux_type_line "$T:0.0" "fugo build" 1.5
p 1.5

# --- Outro: hold a beat before cutting --------------------------------------
p 1.5

echo "==> stopping recorder"
kill -INT "$FFMPEG_PID"
wait "$FFMPEG_PID" 2>/dev/null || true
FFMPEG_PID=""

echo "==> raw recording saved to $OUTPUT"
echo "==> zoom timeline saved to $TIMELINE"
