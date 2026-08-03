#!/usr/bin/env node
// Zero-dependency validator for site/ — checks route integrity and i18n coverage.
//
// Run: node site/validate.js
// Exits non-zero (and CI fails) on any real problem. Uses only Node built-ins
// (fs, path, vm) — no npm install, consistent with the rest of site/ having no
// external dependencies.
//
// What it checks:
//   1. Routes    — every href resolves: same-page anchors exist, cross-page
//                  files exist, cross-page anchors exist in the target file,
//                  no duplicate ids per page.
//   2. i18n keys — every data-i18n key used in HTML resolves to a non-null
//                  string in BOTH es and en, by actually running i18n.js in a
//                  sandboxed VM (not re-parsing it) so this checks real
//                  runtime behavior, not a guess at the file's structure.
//   3. i18n-attr — data-i18n-attr="a|b" and data-i18n="key1|key2" must have
//                  the same number of pipe-separated parts (a common way to
//                  silently under/over-translate an attribute pair).
//   4. Untranslated text — a heuristic scan for <p>/<h2>/<h3>/<li> elements
//                  with real text and no data-i18n anywhere on them, outside
//                  <pre>/<code> (deliberately-literal blocks). Reported as
//                  warnings, not hard failures — it can't see intent, only
//                  absence of the attribute.
//   5. Accessibility — <html lang>, a skip link, <img alt>, aria-label on
//                  icon-only buttons, and WCAG contrast ratios (computed from
//                  the actual hex tokens in styles.css) for every text/background
//                  pair real content renders on, in both themes.
//   6. Design      — every class="..." used in HTML has a matching selector in
//                  styles.css (an unstyled class is usually a typo or a leftover
//                  from a rename — a real design leak, not just untidy).

const fs = require("fs");
const path = require("path");
const vm = require("vm");

const SITE = __dirname;
const PAGES = ["index.html", "changelog.html", "contributing.html"];

let errors = 0;
let warnings = 0;

function fail(msg) {
  console.log(`  \x1b[31m✗\x1b[0m ${msg}`);
  errors++;
}
function warn(msg) {
  console.log(`  \x1b[33m!\x1b[0m ${msg}`);
  warnings++;
}
function ok(msg) {
  console.log(`  \x1b[32m✓\x1b[0m ${msg}`);
}

// ---------- load raw HTML ----------
const html = {};
for (const p of PAGES) html[p] = fs.readFileSync(path.join(SITE, p), "utf8");

// ---------- tiny attribute/tag extraction (regex-based, no real DOM) ----------
function attrs(tag) {
  // all `name="value"` pairs on a single opening tag string
  const out = {};
  const re = /([a-zA-Z-]+)="([^"]*)"/g;
  let m;
  while ((m = re.exec(tag))) out[m[1]] = m[2];
  return out;
}

function findTags(src, tagName) {
  const re = new RegExp(`<${tagName}\\b[^>]*>`, "g");
  return src.match(re) || [];
}

// ---------- 1. routes ----------
console.log("\n== routes ==");
for (const p of PAGES) {
  const src = html[p];
  const ids = new Set();
  const dupIds = new Set();
  for (const tag of src.match(/\bid="[^"]*"/g) || []) {
    const id = tag.slice(4, -1);
    if (ids.has(id)) dupIds.add(id);
    ids.add(id);
  }
  for (const id of dupIds) fail(`${p}: duplicate id="${id}"`);

  const anchorTags = findTags(src, "a").filter((t) => / href="/.test(t));
  for (const tag of anchorTags) {
    const href = attrs(tag).href;
    if (!href || /^https?:\/\//.test(href) || /^mailto:/.test(href)) continue;

    if (href.startsWith("#")) {
      const id = href.slice(1);
      if (!ids.has(id)) fail(`${p}: href="${href}" has no matching id on this page`);
      continue;
    }

    const [file, hash] = href.split("#");
    if (!PAGES.includes(file)) {
      fail(`${p}: href="${href}" points to a file not in the known page set (${PAGES.join(", ")})`);
      continue;
    }
    if (!fs.existsSync(path.join(SITE, file))) {
      fail(`${p}: href="${href}" — ${file} doesn't exist on disk`);
      continue;
    }
    if (hash) {
      const targetIds = new Set((html[file].match(/\bid="[^"]*"/g) || []).map((t) => t.slice(4, -1)));
      if (!targetIds.has(hash)) fail(`${p}: href="${href}" — no id="${hash}" in ${file}`);
    }
  }
}
if (errors === 0) ok(`${PAGES.length} pages, no broken links or duplicate ids`);

// ---------- 2 & 3. i18n key resolution + attr/key parity ----------
console.log("\n== i18n ==");

// Load the real i18n.js in a sandbox with minimal browser stubs, so key
// resolution reflects actual runtime behavior rather than a re-parse.
function loadI18n() {
  const listeners = {};
  const fakeDocument = {
    documentElement: { lang: "", setAttribute() {} },
    querySelectorAll: () => [],
    getElementById: () => null,
    addEventListener(evt, fn) {
      listeners[evt] = listeners[evt] || [];
      listeners[evt].push(fn);
    },
    dispatchEvent() {},
  };
  const sandbox = {
    document: fakeDocument,
    window: {},
    navigator: { language: "es" },
    localStorage: {
      _s: {},
      getItem(k) { return this._s[k] ?? null; },
      setItem(k, v) { this._s[k] = v; },
    },
    CustomEvent: function (name, opts) { this.name = name; this.detail = opts && opts.detail; },
    console,
  };
  sandbox.window.document = sandbox.document;
  vm.createContext(sandbox);
  const code = fs.readFileSync(path.join(SITE, "i18n.js"), "utf8");
  vm.runInContext(code, sandbox, { filename: "i18n.js" });
  // fire DOMContentLoaded synchronously since we're not a real browser event loop
  (listeners.DOMContentLoaded || []).forEach((fn) => fn());
  return sandbox.window.fugoI18n;
}

let fugoI18n;
try {
  fugoI18n = loadI18n();
  if (!fugoI18n) throw new Error("window.fugoI18n was never set");
} catch (e) {
  fail(`i18n.js failed to load in a sandboxed VM: ${e.message}`);
}

if (fugoI18n) {
  // collect every data-i18n / data-i18n-attr usage across all pages
  const usages = []; // { page, keys: [...], attrs: [...] | null }
  for (const p of PAGES) {
    const src = html[p];
    const re = /<[a-zA-Z][^>]*?\bdata-i18n="([^"]+)"[^>]*>/g;
    let m;
    while ((m = re.exec(src))) {
      const tagStr = m[0];
      const keys = m[1].split("|");
      const attrMatch = /data-i18n-attr="([^"]+)"/.exec(tagStr);
      usages.push({ page: p, tag: tagStr.slice(0, 60), keys, attrsList: attrMatch ? attrMatch[1].split("|") : null });
    }
  }

  let missing = 0;
  const allKeys = new Set();
  for (const u of usages) {
    for (const key of u.keys) {
      allKeys.add(key);
      for (const lang of ["es", "en"]) {
        fugoI18n.set(lang);
        const val = fugoI18n.t(key);
        if (val == null) {
          fail(`${u.page}: data-i18n key "${key}" has no ${lang} translation (${u.tag}…)`);
          missing++;
        }
      }
    }
    if (u.attrsList && u.attrsList.length !== u.keys.length) {
      fail(`${u.page}: data-i18n-attr has ${u.attrsList.length} target(s) but data-i18n lists ${u.keys.length} key(s) (${u.tag}…)`);
      missing++;
    }
  }
  if (missing === 0) ok(`${allKeys.size} distinct data-i18n keys used, all resolve in es + en`);
}

// ---------- 4. heuristic: visible text with no data-i18n ----------
console.log("\n== untranslated text (heuristic — review, not auto-fail) ==");
let heuristicHits = 0;
for (const p of PAGES) {
  let src = html[p];
  // strip <pre>...</pre> blocks — code/ASCII diagrams are deliberately literal
  src = src.replace(/<pre[\s\S]*?<\/pre>/g, "");
  // strip any element (and its whole subtree) explicitly marked data-i18n-exempt,
  // e.g. the changelog's release history, which mirrors CHANGELOG.md verbatim
  // in English by design — see DESIGN.md
  src = src.replace(/<(\w+)\b[^>]*\bdata-i18n-exempt\b[^>]*>[\s\S]*?<\/\1>/g, "");

  for (const tagName of ["p", "h2", "h3", "li"]) {
    const re = new RegExp(`<${tagName}\\b([^>]*)>([\\s\\S]*?)<\\/${tagName}>`, "g");
    let m;
    while ((m = re.exec(src))) {
      const openTagAttrs = m[1];
      const inner = m[2];
      const text = inner.replace(/<[^>]+>/g, "").replace(/&[a-z]+;|&#\d+;/gi, " ").trim();
      if (!text || !/[a-zA-Z]{3,}/.test(text)) continue; // skip empty/symbol-only
      if (/data-i18n=/.test(openTagAttrs)) continue;
      warn(`${p}: <${tagName}> with no data-i18n: "${text.slice(0, 70)}${text.length > 70 ? "…" : ""}"`);
      heuristicHits++;
    }
  }
}
if (heuristicHits === 0) ok("no untagged <p>/<h2>/<h3>/<li> text found");

// ---------- 5. accessibility ----------
console.log("\n== accessibility ==");
let a11yHits = 0;

for (const p of PAGES) {
  const src = html[p];

  if (!/<html\b[^>]*\blang="[a-z-]+"/i.test(src)) fail(`${p}: <html> has no lang attribute`);
  if (!/class="skip-link"/.test(src)) fail(`${p}: no .skip-link found`);

  for (const tag of findTags(src, "img")) {
    const a = attrs(tag);
    if (!("alt" in a)) {
      fail(`${p}: <img> with no alt attribute (${tag.slice(0, 60)}…)`);
      a11yHits++;
    }
  }

  // icon-only <button>: contains an <svg> but no visible text and no aria-label
  const buttonRe = /<button\b([^>]*)>([\s\S]*?)<\/button>/g;
  let bm;
  while ((bm = buttonRe.exec(src))) {
    const openAttrs = bm[1];
    const inner = bm[2];
    const hasSvg = /<svg\b/.test(inner);
    const visibleText = inner.replace(/<[^>]+>/g, "").trim();
    if (hasSvg && !visibleText && !/aria-label="/.test(openAttrs)) {
      fail(`${p}: icon-only <button> with no aria-label (${bm[0].slice(0, 60)}…)`);
      a11yHits++;
    }
  }
}

// contrast: WCAG relative-luminance ratio between the ink/background token pairs
// actually used for text, in both themes.
function hexToRgb(hex) {
  const h = hex.replace("#", "");
  const n = parseInt(h.length === 3 ? h.split("").map((c) => c + c).join("") : h, 16);
  return [(n >> 16) & 255, (n >> 8) & 255, n & 255];
}
function relLuminance([r, g, b]) {
  const f = (c) => { c /= 255; return c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4); };
  const [rl, gl, bl] = [f(r), f(g), f(b)];
  return 0.2126 * rl + 0.7152 * gl + 0.0722 * bl;
}
function contrastRatio(hexA, hexB) {
  const la = relLuminance(hexToRgb(hexA));
  const lb = relLuminance(hexToRgb(hexB));
  const [lighter, darker] = la > lb ? [la, lb] : [lb, la];
  return (lighter + 0.05) / (darker + 0.05);
}

function extractThemeTokens(css, blockSelector, fallback) {
  const re = new RegExp(`${blockSelector.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}\\s*\\{([^}]*)\\}`);
  const m = re.exec(css);
  if (!m) return {};
  const tokens = {};
  const refs = {};
  // a declaration is either a literal hex or a reference to another custom property
  // (e.g. --accent-text:var(--ember-deep)) — collect both, resolve refs below,
  // falling back to a previously-resolved token map (e.g. the base :root block)
  // for names not redeclared in this specific block.
  const tokenRe = /--([\w-]+)\s*:\s*(#[0-9a-fA-F]{3,6}|var\(--([\w-]+)\))/g;
  let tm;
  while ((tm = tokenRe.exec(m[1]))) {
    if (tm[3]) refs[tm[1]] = tm[3];
    else tokens[tm[1]] = tm[2];
  }
  for (let i = 0; i < 5; i++) {
    let progressed = false;
    for (const [name, target] of Object.entries(refs)) {
      const resolved = tokens[target] || (fallback && fallback[target]);
      if (resolved && !tokens[name]) { tokens[name] = resolved; progressed = true; }
    }
    if (!progressed) break;
  }
  return tokens;
}

const css = fs.readFileSync(path.join(SITE, "styles.css"), "utf8");
// tokens shared across themes (e.g. --flame-core) live in the base :root{} block and
// are only overridden per-theme where the value actually differs — merge base + override.
const baseTokens = extractThemeTokens(css, ":root");
const themes = {
  light: { ...baseTokens, ...extractThemeTokens(css, ':root[data-theme="light"]', baseTokens) },
  dark: { ...baseTokens, ...extractThemeTokens(css, ':root[data-theme="dark"]', baseTokens) },
};

// pairs that real text actually renders on, per styles.css (text-color, background-color)
const CONTRAST_PAIRS = [
  ["ink", "bg", 4.5, "body text on page background"],
  ["ink-soft", "bg", 4.5, "paragraph text on page background"],
  ["ink-dim", "bg", 3, "secondary/nav text on page background (UI text, 3:1 floor)"],
  ["ink", "bg-elevated", 4.5, "body text on card background"],
  ["accent-text", "bg", 4.5, "link/accent text on page background"],
  ["accent-text", "bg-elevated", 4.5, "link/accent text on card background"],
];

for (const [themeName, tokens] of Object.entries(themes)) {
  for (const [fgTok, bgTok, min, label] of CONTRAST_PAIRS) {
    const fgHex = tokens[fgTok];
    const bgHex = tokens[bgTok];
    if (!fgHex || !bgHex) { warn(`${themeName} theme: missing --${fgTok} or --${bgTok} to check contrast`); continue; }
    const ratio = contrastRatio(fgHex, bgHex);
    if (ratio < min) {
      fail(`${themeName} theme: ${label} — --${fgTok} (${fgHex}) on --${bgTok} (${bgHex}) is ${ratio.toFixed(2)}:1, below ${min}:1`);
      a11yHits++;
    }
  }
}

if (a11yHits === 0) ok("lang, skip-link, img alt, icon-button labels, and color contrast all pass");

// ---------- 6. design consistency (CSS ↔ HTML class usage) ----------
console.log("\n== design (CSS ↔ HTML consistency) ==");
let designHits = 0;

// classes actually defined by a selector in styles.css (rough but effective: any
// `.name` token appearing in a selector position, i.e. before the first `{`)
const cssSelectorText = css.replace(/\/\*[\s\S]*?\*\//g, "").split("}").map((chunk) => {
  const brace = chunk.lastIndexOf("{");
  return brace === -1 ? "" : chunk.slice(0, brace);
}).join("\n");
const cssClasses = new Set((cssSelectorText.match(/\.[a-zA-Z][\w-]*/g) || []).map((c) => c.slice(1)));

const htmlClasses = new Set();
for (const p of PAGES) {
  for (const m of html[p].matchAll(/\bclass="([^"]*)"/g)) {
    for (const cls of m[1].split(/\s+/)) if (cls) htmlClasses.add(cls);
  }
}

for (const cls of htmlClasses) {
  if (!cssClasses.has(cls)) {
    fail(`class="${cls}" used in HTML but no matching selector in styles.css`);
    designHits++;
  }
}
if (designHits === 0) ok(`${htmlClasses.size} distinct HTML classes, all styled in styles.css`);

// ---------- summary ----------
console.log(`\n${errors === 0 ? "\x1b[32mPASS\x1b[0m" : "\x1b[31mFAIL\x1b[0m"} — ${errors} error(s), ${warnings} warning(s)\n`);
process.exit(errors === 0 ? 0 : 1);
