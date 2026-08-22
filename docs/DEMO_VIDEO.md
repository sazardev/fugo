# Demo video headless — `scripts/demo/`

Grabador del video demo oficial de Fugo ("de cero a una app corriendo, con hot
reload en vivo"), 100 % headless y reproducible: nada toca el escritorio real.
Toda la rama `feat/headless-demo-recorder` construye este pipeline; este doc es
el registro de lo que se hizo, cómo funciona por dentro y cómo usarlo.

## Qué produce

Un MP4 tipo comercial corto:

1. **Intro card** — "FUGO / Go escribe. Flutter renderiza." con la paleta ember
   del sitio (`site/styles.css`).
2. **Footage real** — un terminal con la config genuina del machine (alacritty +
   nvim propios, sin overrides) donde se ejecutan, de verdad, `fugo init` → tour
   del scaffold → `fugo doctor` → `fugo run` (la ventana Flutter real aparece) →
   tres ediciones en vivo en nvim que disparan hot reload → `fugo build`. Con
   zooms digitales sincronizados y lower-thirds al abrir/cerrar.
3. **Outro card** — repo + tagline.

La app demo no es el contador del template `app`: es un scaffold mobile-style
escrito a mano (app bar, lista de tareas, FAB funcional, bottom nav de 3 tabs)
en una **ventana portrait 420×900** como de teléfono — ese es el punto del demo.

## Cronología de la rama

| Commit | Aporte |
|---|---|
| `f5ec405` | Pipeline base headless: `record-demo.sh` + libs (Xvfb, x11, typewriter, generador de zoom) + `postprocess.sh` (cards intro/outro + lower-thirds). |
| `49df4be` | Branding del demo: `assets/branded_main.go.tmpl` aplica los tokens ember del sitio al tema de fg (`#FF6A1A`, superficies near-black, radius 0). |
| `37729b5` | Branding del terminal y las cards: `alacritty-ember.toml` + `tmux-ember.conf` (luego retirados) y paleta ember en las title cards de postproducción. |
| `3f21f1c` | Editor real: se abandona el terminal tematizado y se usa nvim con la config propia de la máquina; alternancia full-screen terminal ↔ app vía raise; las 3 ediciones en vivo definidas en `hot_reload_edit.env`. |
| `717a732` | Scaffold mobile-style real (`mobile_home.go.tmpl`) y ventana portrait fijada desde `fugo.toml` antes de `fugo run` (ya no un contador landscape). |

## Piezas

| Archivo | Rol |
|---|---|
| `record-demo.sh` | Orquestador: levanta Xvfb, lanza tmux+alacritty+nvim, ejecuta los 6 beats del guion, escribe `<raw>.timeline.tsv` con cada zoom beat. |
| `postprocess.sh` | Convierte el raw en el corte final: genera intro/outro con ffmpeg lavfi, aplica el zoom por `sendcmd` + lower-thirds, concatena. |
| `lib/xvfb.sh` | Ciclo de vida del display virtual `:97` (1920×1080x24). |
| `lib/x11.sh` | xdotool puro: esperar por clase de ventana, mover sin redimensionar, fullscreen, raise, click. Sin WM ni root. |
| `lib/typewriter.sh` | Tipeo humano en panes tmux reales (`send-keys -l` char a char con jitter); todo lo tipeado se ejecuta de verdad. |
| `lib/gen_zoom_filter.py` | Timeline → archivo `sendcmd` de ffmpeg: crop que se encoge hacia (cx,cy), aguanta y vuelve; re-escalado a full frame. |
| `assets/branded_main.go.tmpl` | `main.go` del proyecto demo con el tema ember de Fugo inyectado. |
| `assets/mobile_home.go.tmpl` | El scaffold mobile (tareas/FAB/bottom nav) que reemplaza a `ui/home.go` tras `fugo init`. |
| `assets/hot_reload_edit.env` | Las 3 ediciones en vivo (búsqueda → línea nueva) que dispara el hot reload. |

## Cómo funciona (lo no obvio)

- **Todo corre contra Xvfb**, nunca en el desktop; ffmpeg `x11grab` captura el
  framebuffer virtual directo. Crítico: `unset WAYLAND_DISPLAY` para TODO hijo
  de Xvfb — si está seteada, alacritty y el cliente GTK/Flutter prefieren
  Wayland y conectan silenciosamente al compositor real sin mapear jamás una
  ventana en el display virtual (hallazgo empírico, ver `lib/xvfb.sh`).
- **Sin WM bajo Xvfb** no hay minimizar/ocultar fiable: la alternancia
  "editor ↔ app" es simplemente *raise* del window que va arriba
  (`x11_raise`). El terminal queda full-screen abajo haciendo de fondo oscuro.
- **La app nunca se redimensiona en caliente**: su tamaño portrait final se fija
  ANTES de `fugo run` editando `fugo.toml` (+ env `FUGO_WIDTH/HEIGHT/TITLE`);
  el resize runtime de ventanas GTK/Flutter bajo Xvfb no es fiable, por eso
  `x11_move_only` solo la mueve al centro del canvas.
- **Zoom digital en post**: Xvfb no tiene compositor que haga zoom en vivo.
  `record-demo.sh` loguea cada pulso (centro, duraciones, peak) en un TSV y
  duerme lo que dura el beat para que el contenido quede quieto en cámara;
  `gen_zoom_filter.py` lo convierte en comandos `crop` por timestamp. Se usa
  `sendcmd` y no expresiones `crop w/h` porque este ffmpeg evalúa w/h UNA sola
  vez en init (empírico); `sendcmd` sí aplica comandos runtime, así que muchos
  pasos pequeños = movimiento suave.
- **Los clicks son best-effort**: FAB y tabs del bottom nav se clickean por
  coordenadas derivadas de la geometría conocida de la ventana (APPX/APPY…),
  no consultadas en vivo; si Material cambia de layout, falla inofensivamente.

## Gotchas descubiertos (documentados en comentarios del código)

1. **Gap real de producto**: el path hot-reload de `fugo run`
   (`cmd/fugo` → `startFlutterClient`) spawnea el binario Flutter sin forwardear
   `FUGO_WIDTH/HEIGHT/TITLE` ni `FUGO_THEME_SEED/BRIGHTNESS`, así que la
   ventana cae al default 800×600 de main.dart y esquema claro aunque
   `fugo.toml`/`fg.UseTheme` digan otra cosa. El script lo sortea exportando
   esas vars en el shell que lanza `fugo run`; el fix de verdad pertenece a
   cmd/fugo (candidato a issue).
2. **`--no-git` en `fugo init`**: el git init + commit inicial por defecto
   hereda la config global del dueño de la máquina — con firma GPG activa
   aparecería un diálogo pinentry REAL que bloquea la grabación.
3. **Shell plano para el terminal grabado** (`bash --rcfile` propio): evita que
   un wizard de primera ejecución (Powerlevel10k/oh-my-zsh) robe los keystrokes;
   pero el *look* (colores nvim/alacritty) sí es el real, a propósito.
4. **Ediciones nvim con `0C` y no `cc`/`S`**: `C` (change-to-EOL) no reaplica
   autoindent, así que los tabs propios del replacement quedan exactos sin
   importar la config de indent del nvim local.
5. **Split de tmux no hereda el cwd vivo** del pane origen: hay que `cd`
   explícito antes de abrir el editor.

## Uso

```sh
make demo-record    # graba el raw (construye cli + flutter release si faltan)
make demo-post      # corta el raw más nuevo -> out/fugo-demo-final.mp4
make demo           # ambos en un tiro
FAST=1 make demo-record   # smoke test: todos los sleeps a 1/4
```

Prerrequisitos: Linux con Xvfb, xdotool, tmux, alacritty, nvim, python3, ffmpeg
(y Go + Flutter SDK para los builds). Los outputs van a `scripts/demo/out/`,
gitignoreado — regenerables siempre.

## Estado

Pipeline completo y probado de punta a punta: ya existe un raw + un corte final
generados localmente. Pendiente externo: publicar el video (README/site) cuando
se apruebe el corte.
