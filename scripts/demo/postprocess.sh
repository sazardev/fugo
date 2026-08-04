#!/usr/bin/env bash
# postprocess.sh — turns the raw screen recording from record-demo.sh into a
# short commercial cut: a title card, the live footage with a lower-third
# label at open/close, and an outro card. Real ffmpeg encode, no external
# services.
#
# A sibling <raw>.timeline.tsv (written by record-demo.sh) drives a real,
# precisely-timed digital zoom — a crop that shrinks toward the logged
# (cx, cy) and grows back, then gets scaled back up to full frame. Xvfb has
# no compositor to do this live, so it happens here instead.
#
# Usage: scripts/demo/postprocess.sh <raw.mp4> [output.mp4]
set -euo pipefail

RAW="${1:?usage: postprocess.sh <raw.mp4> [output.mp4]}"
DEMO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT="${2:-$DEMO_DIR/out/fugo-demo-final.mp4}"
TIMELINE="${RAW%.mp4}.timeline.tsv"

command -v ffmpeg >/dev/null 2>&1 || {
	echo "ffmpeg not found" >&2
	exit 1
}
[[ -f "$RAW" ]] || {
	echo "raw recording not found: $RAW" >&2
	exit 1
}

FONT="/usr/share/fonts/TTF/DejaVuSans-Bold.ttf"
[[ -f "$FONT" ]] || FONT=$(fc-match -f '%{file}' sans-bold 2>/dev/null || echo "")

read -r W H FPS <<<"$(ffprobe -v error -select_streams v:0 \
	-show_entries stream=width,height,r_frame_rate \
	-of csv=p=0 "$RAW" | tr ',/' '  ' | awk '{printf "%s %s %s", $1, $2, ($3 && $4 ? $3/$4 : 30)}')"
W="${W:-1920}"
H="${H:-1080}"
FPS="${FPS:-30}"
DUR=$(ffprobe -v error -show_entries format=duration -of csv=p=0 "$RAW")

INTRO="$DEMO_DIR/out/.intro.mp4"
OUTRO="$DEMO_DIR/out/.outro.mp4"
MAIN="$DEMO_DIR/out/.main.mp4"

echo "==> generating intro card"
ffmpeg -y -f lavfi -i "color=c=0x0B0F14:s=${W}x${H}:d=2.5:r=${FPS}" -vf "
drawtext=fontfile=${FONT}:text='FUGO':fontsize=140:fontcolor=white:x=(w-text_w)/2:y=(h-text_h)/2-60:
  alpha='if(lt(t,0.4),t/0.4,1)',
drawtext=fontfile=${FONT}:text='Go escribe. Flutter renderiza.':fontsize=42:fontcolor=0xB9C2CC:x=(w-text_w)/2:y=(h/2)+80:
  alpha='if(lt(t,1.0),0,if(lt(t,1.4),(t-1.0)/0.4,1))'
" -c:v libx264 -pix_fmt yuv420p -an "$INTRO"

echo "==> generating outro card"
ffmpeg -y -f lavfi -i "color=c=0x0B0F14:s=${W}x${H}:d=2.5:r=${FPS}" -vf "
drawtext=fontfile=${FONT}:text='github.com/sazardev/fugo':fontsize=54:fontcolor=white:x=(w-text_w)/2:y=(h-text_h)/2-30:
  alpha='if(lt(t,0.4),t/0.4,1)',
drawtext=fontfile=${FONT}:text='Server-Driven UI. 100% Go.':fontsize=34:fontcolor=0xB9C2CC:x=(w-text_w)/2:y=(h/2)+40:
  alpha='if(lt(t,0.8),0,if(lt(t,1.2),(t-0.8)/0.4,1))'
" -c:v libx264 -pix_fmt yuv420p -an "$OUTRO"

echo "==> generating zoom filter from $TIMELINE"
ZOOM_CMDS="$DEMO_DIR/out/.zoom-cmds.txt"
ZOOM_FILTER=$(python3 "$DEMO_DIR/lib/gen_zoom_filter.py" "$TIMELINE" "$W" "$H" "$ZOOM_CMDS")

echo "==> labeling main footage (zoom beats + lower-thirds at open/close)"
ffmpeg -y -i "$RAW" -vf "
${ZOOM_FILTER},
drawtext=fontfile=${FONT}:text='FUGO — Go escribe. Flutter renderiza.':fontsize=28:fontcolor=white:
  box=1:boxcolor=0x0B0F14@0.55:boxborderw=12:x=40:y=h-th-40:
  enable='between(t,0,4)',
drawtext=fontfile=${FONT}:text='fugo run — hot reload en vivo':fontsize=28:fontcolor=white:
  box=1:boxcolor=0x0B0F14@0.55:boxborderw=12:x=40:y=h-th-40:
  enable='between(t,${DUR%.*}-4,${DUR%.*})'
" -c:v libx264 -pix_fmt yuv420p -an -crf 18 -preset medium "$MAIN"

echo "==> concatenating intro + main + outro"
ffmpeg -y \
	-i "$INTRO" -i "$MAIN" -i "$OUTRO" \
	-filter_complex "[0:v][1:v][2:v]concat=n=3:v=1:a=0[v]" \
	-map "[v]" -c:v libx264 -pix_fmt yuv420p -crf 18 -preset medium "$OUT"

rm -f "$INTRO" "$OUTRO" "$MAIN" "$ZOOM_CMDS"
echo "==> final cut: $OUT"
