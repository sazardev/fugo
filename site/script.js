(function () {
  "use strict";

  // ---- theme: manual toggle overrides prefers-color-scheme, persisted ----
  var root = document.documentElement;
  var stored = null;
  try { stored = localStorage.getItem("fugo-theme"); } catch (e) { /* storage unavailable */ }
  if (stored === "light" || stored === "dark") root.setAttribute("data-theme", stored);

  var themeBtn = document.getElementById("themeToggle");
  if (themeBtn) {
    themeBtn.addEventListener("click", function () {
      var current = root.getAttribute("data-theme");
      if (!current) {
        current = window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
      }
      var next = current === "dark" ? "light" : "dark";
      root.setAttribute("data-theme", next);
      try { localStorage.setItem("fugo-theme", next); } catch (e) { /* storage unavailable */ }
    });
  }

  // ---- mobile sidebar ----
  var navToggle = document.getElementById("navToggle");
  var sidebar = document.getElementById("sidebar");
  if (navToggle && sidebar) {
    navToggle.addEventListener("click", function () {
      var open = sidebar.classList.toggle("open");
      navToggle.setAttribute("aria-expanded", String(open));
    });
    sidebar.querySelectorAll("a").forEach(function (link) {
      link.addEventListener("click", function () {
        sidebar.classList.remove("open");
        navToggle.setAttribute("aria-expanded", "false");
      });
    });
  }

  // ---- active section highlight in sidebar ----
  var links = Array.prototype.slice.call(document.querySelectorAll(".side-nav a"));
  var sections = links
    .map(function (link) { return document.getElementById(link.getAttribute("href").slice(1)); })
    .filter(Boolean);

  if ("IntersectionObserver" in window && sections.length) {
    var observer = new IntersectionObserver(
      function (entries) {
        entries.forEach(function (entry) {
          if (!entry.isIntersecting) return;
          links.forEach(function (link) { link.classList.remove("active"); });
          var active = links.find(function (link) { return link.getAttribute("href") === "#" + entry.target.id; });
          if (active) active.classList.add("active");
        });
      },
      { rootMargin: "-40% 0px -55% 0px", threshold: 0 }
    );
    sections.forEach(function (section) { observer.observe(section); });
  }

  // ---- command palette (Ctrl/Cmd+K) ----
  var overlay = document.getElementById("paletteOverlay");
  var input = document.getElementById("paletteInput");
  var resultsEl = document.getElementById("paletteResults");
  var trigger = document.getElementById("paletteTrigger");

  function jumpTo(hash) {
    var id = (hash || "").replace(/^#/, "");
    var target = id ? document.getElementById(id) : null;
    if (!target) return;
    if (history.pushState) history.pushState(null, "", "#" + id);
    else window.location.hash = "#" + id;
    // requestAnimationFrame: run after the overlay has actually been hidden
    // (closing it first keeps it from covering the scroll animation).
    requestAnimationFrame(function () {
      target.scrollIntoView({ behavior: "smooth", block: "start" });
    });
  }

  function t(key, fallback) {
    var v = window.fugoI18n && window.fugoI18n.t(key);
    return v == null ? fallback : v;
  }

  var commands = [];
  var filtered = [];
  var activeIndex = 0;

  function buildCommands() {
    commands = links.map(function (link) {
      var href = link.getAttribute("href") || "";
      var isHash = href.charAt(0) === "#";
      return {
        label: link.textContent.trim(),
        href: href,
        group: isHash ? t("palette.groupSection", "Sección") : t("palette.groupLink", "Enlace"),
        action: function () {
          if (isHash) jumpTo(href);
          else window.location.href = href;
        }
      };
    }).concat([
      { label: t("palette.openGithub", "Abrir en GitHub"), group: t("palette.groupLink", "Enlace"), action: function () { window.open("https://github.com/sazardev/fugo", "_blank", "noopener"); } },
      { label: t("palette.viewChangelog", "Ver Changelog"), group: t("palette.groupLink", "Enlace"), action: function () { window.open("https://github.com/sazardev/fugo/blob/main/CHANGELOG.md", "_blank", "noopener"); } },
      { label: t("palette.toggleTheme", "Cambiar tema claro/oscuro"), group: t("palette.groupAction", "Acción"), action: function () { themeBtn && themeBtn.click(); } }
    ]);
    filtered = commands.slice();
    activeIndex = 0;
  }
  buildCommands();
  document.addEventListener("fugo:langchange", buildCommands);

  function render() {
    resultsEl.innerHTML = "";
    if (!filtered.length) {
      var empty = document.createElement("li");
      empty.className = "empty";
      empty.textContent = t("palette.empty", "Sin resultados");
      resultsEl.appendChild(empty);
      return;
    }
    filtered.forEach(function (cmd, i) {
      var li = document.createElement("li");
      li.textContent = cmd.label;
      if (i === activeIndex) li.classList.add("active");
      var group = document.createElement("span");
      group.className = "p-group";
      group.textContent = cmd.group;
      li.appendChild(group);
      li.addEventListener("mouseenter", function () { activeIndex = i; render(); });
      li.addEventListener("click", function () { run(cmd); });
      resultsEl.appendChild(li);
    });
  }

  function run(cmd) {
    if (!cmd) return;
    close();
    cmd.action();
  }

  function open() {
    overlay.hidden = false;
    input.value = "";
    filtered = commands.slice();
    activeIndex = 0;
    render();
    setTimeout(function () { input.focus(); }, 0);
  }

  function close() {
    overlay.hidden = true;
    if (trigger) trigger.focus();
  }

  if (overlay && input && resultsEl) {
    if (trigger) trigger.addEventListener("click", open);

    document.addEventListener("keydown", function (e) {
      var isK = e.key === "k" || e.key === "K";
      if ((e.metaKey || e.ctrlKey) && isK) {
        e.preventDefault();
        overlay.hidden ? open() : close();
      } else if (e.key === "Escape" && !overlay.hidden) {
        close();
      }
    });

    overlay.addEventListener("click", function (e) {
      if (e.target === overlay) close();
    });

    input.addEventListener("input", function () {
      var q = input.value.trim().toLowerCase();
      filtered = q ? commands.filter(function (c) { return c.label.toLowerCase().indexOf(q) !== -1; }) : commands.slice();
      activeIndex = 0;
      render();
    });

    input.addEventListener("keydown", function (e) {
      if (e.key === "ArrowDown") {
        e.preventDefault();
        activeIndex = Math.min(activeIndex + 1, filtered.length - 1);
        render();
      } else if (e.key === "ArrowUp") {
        e.preventDefault();
        activeIndex = Math.max(activeIndex - 1, 0);
        render();
      } else if (e.key === "Enter") {
        e.preventDefault();
        run(filtered[activeIndex]);
      }
    });
  }
})();
