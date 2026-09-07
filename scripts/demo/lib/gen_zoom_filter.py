#!/usr/bin/env python3
"""Turns a recorded zoom-pulse timeline into an ffmpeg `sendcmd` script that
drives a real, precisely-timed digital zoom: a crop window that shrinks
toward (cx, cy) and grows back, scaled up to the output resolution.

Why sendcmd and not a time-expression crop: this ffmpeg build's `crop`
filter only evaluates w/h expressions ONCE at init (x/y alone would update
per-frame, but with w/h frozen at their t=0 value the crop never actually
changes size) — confirmed empirically, not every ffmpeg build behaves this
way, but this is the portable fix: sendcmd emits explicit `crop w/h/x/y`
commands at closely-spaced timestamps, which crop's runtime command
handler DOES apply every time, so many small steps look like smooth
motion.

Xvfb has no compositor to do this live (unlike the Hyprland
`cursor:zoom_factor` this pipeline used before going headless), so it
happens here, once, at encode time.

Timeline file format (one pulse per line, whitespace-separated):
    start_elapsed  in_dur  hold_dur  out_dur  cx  cy  peak_zoom

Usage: gen_zoom_filter.py <timeline.tsv> <out_w> <out_h> <cmds_path> > filter.txt
Writes the sendcmd command file to <cmds_path> and prints the -vf filter
chain (referencing it) to stdout.
"""
import sys

STEP = 0.05  # seconds between keyframes — smooth enough, not too many lines


def envelope(t, t0, din, dhold, dout):
    t1, t2, t3 = t0 + din, t0 + din + dhold, t0 + din + dhold + dout
    if t < t0 or t >= t3:
        return 0.0
    if t < t1:
        return (t - t0) / din
    if t < t2:
        return 1.0
    return 1.0 - (t - t2) / dout


def main():
    path, ow, oh, cmds_path = sys.argv[1], int(sys.argv[2]), int(sys.argv[3]), sys.argv[4]

    pulses = []
    try:
        with open(path) as f:
            for line in f:
                line = line.strip()
                if not line or line.startswith("#"):
                    continue
                t0, din, dhold, dout, cx, cy, peak = line.split()
                pulses.append(
                    (float(t0), float(din), float(dhold), float(dout), float(cx), float(cy), float(peak))
                )
    except FileNotFoundError:
        pulses = []

    lines = []
    for t0, din, dhold, dout, cx, cy, peak in pulses:
        t3 = t0 + din + dhold + dout
        n_steps = max(2, int((t3 - t0) / STEP) + 1)
        for i in range(n_steps + 1):
            t = t0 + (t3 - t0) * i / n_steps
            env = envelope(t, t0, din, dhold, dout)
            zoom = 1 + (peak - 1) * env
            w = ow / zoom
            h = oh / zoom
            x = min(max(cx - w / 2, 0), ow - w)
            y = min(max(cy - h / 2, 0), oh - h)
            lines.append(
                f"{t:.3f} crop w {w:.1f}, crop h {h:.1f}, crop x {x:.1f}, crop y {y:.1f};"
            )
        # Snap fully back to the untouched full frame at pulse end.
        lines.append(f"{t3:.3f} crop w {ow}, crop h {oh}, crop x 0, crop y 0;")

    with open(cmds_path, "w") as f:
        f.write("\n".join(lines) + ("\n" if lines else ""))

    print(f"sendcmd=f='{cmds_path}',crop=w={ow}:h={oh}:x=0:y=0,scale={ow}:{oh}")


if __name__ == "__main__":
    main()
