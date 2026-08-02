# Fugo site — design system

This documents the design system behind `site/` (the project presentation page deployed to
GitHub Pages via `.github/workflows/pages.yml`). It's a **single static page** — no build
step, no framework, no external requests — so the system below is enforced by convention
(CSS custom properties + a few structural rules), not by tooling.

## Inspiration

- **VitePress docs layout** — sticky top bar, a left sidebar of section anchors, a single
  readable content column. Familiar to anyone who's used Go/Rust/JS project docs; no
  onboarding needed to navigate it.
- **Fugo's own mark** — the flame-F: a rigid stem (Go, the stable engine) with flame-tongue
  arms (Flutter, the render surface). The site's palette is lifted directly from that mark
  rather than invented separately, so the logo and the page read as the same object.
- **Flat, borderless, rectangular** — a deliberate reaction against soft-UI defaults
  (rounded cards, hairline borders, drop shadows). Every visual boundary in this page is a
  **background-color step**, never a `border` or `border-radius`. This was a direct design
  request, not an accident — see "Borders: none, ever" below.
- **A spec sheet, not a marketing page** — the content is status/scope/security/CI, so the
  layout borrows more from technical documentation than from a product landing page: dense,
  scannable, low-ornament.

## Color

All colors are CSS custom properties on `:root`, redefined under
`@media (prefers-color-scheme: dark)` and again under `:root[data-theme="dark"]` /
`:root[data-theme="light"]` so a manual toggle (see `i18n.js`/`script.js` — theme state
lives in `localStorage["fugo-theme"]`) always wins over the OS preference. Components
consume the tokens, never raw hex — this is what makes the whole page re-theme from one
place.

### Brand (fixed — same value in both themes)

| Token | Value | Use |
|---|---|---|
| `--ember-deep` | `#c7350c` | Hover state for the primary button, deepest point of the flame gradient in the logo |
| `--flame-core` | `#ff6a1a` | The one accent color: links, active nav indicator, primary button, badge highlight |
| `--flame-bright` | `#ffb238` | Reserved for the logo/lockups; rarely used directly in page CSS |
| `--flame-gold` | `#ffdd6b` | Logo fold accent only — not used in page chrome |

`--flame-core` is the **only** accent in the entire page. It never competes with a second
accent color — semantic meaning (pros vs. cons, in-scope vs. out-of-scope) is carried by
`--ink` vs. `--ink-dim` and by the card's position in a `.card-in`/`.card-out` pair, not by
introducing a second hue.

### Neutrals (theme-dependent)

| Token | Light | Dark | Use |
|---|---|---|---|
| `--bg` | `#ede7dc` | `#17130f` | Page background |
| `--bg-soft` | `#e2dbcd` | `#1f1911` | One step up: topbar, sidebar, hero, footer — "banded" sections |
| `--bg-card` | `#e6e0d3` | `#201911` | Cards, code blocks, callouts, the command palette panel |
| `--bg-strong` | `#d9d1bf` | `#2a2116` | Two steps up: badges, buttons' resting background, palette input row |
| `--ink` | `#221a14` | `#f2eae0` | Primary text |
| `--ink-dim` | `#5c5044` | `#b8ab98` | Secondary text, nav links, captions |
| `--code-bg` | `#e2dbcd` | `#1f1911` | `<pre>`/`<code>` background |

Both light and dark are **warm neutrals** — a slight orange/brown bias, never a pure grey —
so they read as chosen to sit next to the flame accent rather than a generic UI grey with
orange bolted on top.

### Borders: none, ever

There is no `border` (beyond `border: 0` resets) and no `border-radius` anywhere in
`styles.css`. Every visual seam — topbar/content, sidebar/main, card/card in a grid,
active-nav indicator, callout accent, palette input row — is drawn with a **background
color step** between adjacent `--bg*` tokens, or with a small solid `::before` block
(e.g. `.side-nav a.active::before`, `.callout::before`) instead of a `border-left`. This
keeps the whole page rectangular and flat by construction: there's no rule to "remember,"
because the primitive (a border) simply isn't in the toolbox.

## Typography

No webfonts — this is a zero-external-request page, so type comes from the OS via careful
stacks, chosen to feel considered rather than left to the browser default:

```css
/* body + headings */
font-family: ui-sans-serif, "Segoe UI", "Avenir Next", system-ui, sans-serif;

/* code, badges, nav "brand-version" pill, section labels, the command palette's kbd hints */
font-family: ui-monospace, "Cascadia Code", "SF Mono", monospace;
```

The monospace face isn't decorative — it marks anything literal (a version string, a shell
command, an env var, a keyboard shortcut), which matters on a page that's mostly *describing
a CLI and a wire protocol*.

### Scale

Fully fluid via `clamp()`, not fixed breakpoint jumps — a heading is a different size at
900px than at 1400px, not just at the two or three points a media query happens to check:

| Element | `clamp()` |
|---|---|
| `<h1>` (hero) | `clamp(32px, 6vw, 52px)` |
| Hero tagline | `clamp(15px, 2.4vw, 19px)` |
| `<h2>` (section) | `clamp(21px, 3vw, 26px)` |
| Body | `clamp(15px, 0.5vw + 14px, 16px)` |
| ASCII diagram (`pre.diagram`) | `clamp(9.5px, 1.6vw, 12.5px)` — the widest single element on the page, so it needs the most aggressive shrink to survive a phone width without horizontal scroll |

## Spacing & layout tokens

```css
--sidebar-w: clamp(200px, 18vw, 260px);
--topbar-h:  56px;
--content-w: min(72ch, 100%);

--space-xs: clamp(6px,  0.6vw, 10px);
--space-sm: clamp(10px, 1.2vw, 16px);
--space-md: clamp(16px, 2.4vw, 28px);
--space-lg: clamp(28px, 4.5vw, 56px);
--space-xl: clamp(40px, 7vw, 88px);
```

A five-step fluid scale (`xs`→`xl`) replaces fixed pixel margins/padding/gaps throughout —
`section` spacing, card padding, the hero's vertical rhythm, the sidebar's inset. Content
width caps at `72ch` (not a pixel value) so line length stays readable regardless of how
wide the content column is allowed to get.

### Grids

`.grid-2`/`.grid-3` are actually the **same rule**:

```css
grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
```

They reflow continuously — 3 columns, 2, or 1 depending on available width — instead of
snapping at a fixed set of breakpoints. This is the "fluid on every side" requirement: sizing
adapts constantly, and the handful of `@media` queries that remain exist only for
**structural** mode changes a fluid rule can't express (the sidebar becoming a slide-in
drawer below 900px, the topbar's search label/kbd-hint hiding to save space).

## Components

| Component | Notes |
|---|---|
| **Topbar** | Sticky, `--bg-soft`. Brand mark + version pill, search/command-palette trigger, GitHub/Changelog links, language toggle, theme toggle. Collapses to a hamburger (`.nav-toggle`) below 900px. |
| **Sidebar** | Sticky under the topbar, `--bg-soft`, grouped anchor links with an active-section indicator (`::before` bar in `--flame-core`, driven by an `IntersectionObserver` in `script.js`). Becomes a fixed slide-in drawer on narrow viewports. |
| **Hero** | The one section allowed a "banded" full-bleed background (`--bg-soft`, bleeding past `main`'s padding) — it's the thesis statement, so it gets a visually distinct band the way a doc page's title block would. |
| **Cards / grids** | Flat `--bg-card` tiles separated by a 2px `--bg` gap (not a border) — see `.grid-2`/`.grid-3`. Paired grids (`.card-in`/`.card-out`) encode pros/cons and in-scope/out-of-scope by position + `--flame-core` vs `--ink-dim` heading color, not by a second accent hue. |
| **Badges** | Flat `--bg-strong` pill (no radius — "pill" here just means a small rectangle), monospace. `.badge-flame` is the one place a badge takes the accent fill. |
| **Callouts** | `--bg-card` block with a solid 4px left accent block (`::before`), used sparingly for the two or three claims on the page worth visually isolating ("Go is the source of truth", "this page deploys itself"). |
| **Command palette** | `Ctrl/Cmd+K` or the topbar trigger opens a centered, borderless `--bg-card` panel (`site/script.js`). Filters by label as you type, arrow keys + Enter to navigate, click-outside or Escape to close. Jumping to a section uses `scrollIntoView({behavior:'smooth'})` + `history.pushState` (not a bare `location.hash` assignment, which is a no-op when the hash doesn't change — e.g. re-selecting the section you're already on). |

## Internationalization

`i18n.js` holds two flat-nested dictionaries (`es`/`en`) and an `applyLang(lang)` that walks
every `[data-i18n]` element: elements with a matching `data-i18n-attr` get an attribute set
(`aria-label`, `title`, `placeholder`, …); everything else gets `innerHTML` replaced (safe
here since both dictionaries are authored content, not user input). Language is detected from
`localStorage["fugo-lang"]`, then `navigator.language`, defaulting to Spanish. `script.js`'s
command palette re-reads its labels via `window.fugoI18n.t(key)` and rebuilds on a
`fugo:langchange` event, so dynamically-generated UI stays in sync with a language switch,
not just the static markup.

## Usability & accessibility

- **Skip link** (`.skip-link`) — first focusable element, jumps to `#main`.
- **Visible focus** — a single `:focus-visible` rule (`outline: 2px solid var(--flame-core)`)
  applies uniformly instead of relying on (or suppressing) the browser default.
- **`prefers-reduced-motion`** — smooth scrolling and all transitions/animations are disabled
  wholesale when the OS requests it.
- **Keyboard-first command palette** — open (`Ctrl/Cmd+K`), filter (type), navigate
  (`↑`/`↓`), commit (`Enter`), dismiss (`Escape` or click-outside). The trigger button also
  carries a translated `aria-label` and `title`.
- **`aria-label`s throughout** — sidebar nav, mobile nav toggle, theme toggle, the palette
  dialog itself (`role="dialog" aria-modal="true"`) — and all of them are localized, not just
  the visible copy.
- **No dependency on JS for content** — the page's information (status, scope, security,
  CI/CD) is present in the initial HTML in Spanish (the `data-i18n` default); JS only
  swaps language and adds the palette/theme conveniences. A crawler or a no-JS browser still
  gets the full page.

## Multi-page structure

The site is three static pages sharing the same shell (topbar, sidebar, palette, theme/lang
toggles) and the same `styles.css`/`i18n.js`/`script.js`:

| Page | Purpose |
|---|---|
| `index.html` | The pitch: what/why/objective/audience/how-it-works, status, scope, pros & cons, security, CI/CD, license. |
| `changelog.html` | The versioning scheme + the full release history, translated from `CHANGELOG.md` into both languages — every release entry is fully bilingual, same as the rest of the site. |
| `contributing.html` | Dev setup, code style, tests, the commit/PR flow (mirrors `CONTRIBUTING.md`), releases, reporting bugs/security, code of conduct — fully bilingual, since it's original page copy rather than a verbatim technical log. |

Each page's sidebar is **page-local** (its own anchors under "On this page"), not a single
global nav repeated everywhere — a shared sidebar would need `index.html#que-es`-style hrefs
on every other page and silently break if a page reordered its sections. A small "Resources"
group cross-links the three pages. `script.js`'s command palette reads whatever sidebar
exists in the current page's DOM, so it needs no per-page configuration — and it now
distinguishes in-page jumps (`#id`, via `scrollIntoView`) from cross-page links (a bare
`location.href` navigation), which the initial single-page version never had to handle.

## Files

```
site/
├── index.html         # landing/pitch page
├── changelog.html      # versioning scheme + full release history (mirrors CHANGELOG.md)
├── contributing.html   # dev setup, code style, PR flow, code of conduct (mirrors CONTRIBUTING.md)
├── styles.css          # tokens (:root) + components + the handful of structural breakpoints
├── i18n.js              # es/en dictionaries, applyLang(), language toggle + persistence
├── script.js            # theme toggle, mobile sidebar, active-section highlight, command palette
├── validate.js          # zero-dependency route + i18n coverage validator, see below
├── DESIGN.md            # this file
└── assets/
    └── logo.svg          # flat single-path flame-F, no gradient — see ../assets/logo.svg (same file)
```

## Validation

`node site/validate.js` (Node built-ins only — no install) gates every deploy in
`.github/workflows/pages.yml`, run before the Pages artifact is built:

- **Routes** — every `href` resolves: same-page anchors exist, cross-page files exist,
  cross-page anchors exist in the target file, no duplicate `id`s per page.
- **i18n keys** — every `data-i18n` key actually resolves to a string in *both* `es` and
  `en`, checked by running the real `i18n.js` in a sandboxed VM (`node:vm`) rather than
  re-parsing the dictionaries — this reflects real runtime behavior, not a guess at the
  file's shape. A `data-i18n-attr="a|b"` with a mismatched number of `data-i18n="k1|k2"`
  keys is also caught (a common way to silently under-translate an attribute pair).
- **Untranslated text** — a heuristic, non-fatal scan for `<p>`/`<h2>`/`<h3>`/`<li>` text
  with no `data-i18n` anywhere on the tag, outside `<pre>` (deliberately-literal code) and
  outside anything explicitly marked `data-i18n-exempt` (for a future block that's meant
  to stay untranslated by design — no page currently needs the exemption).
