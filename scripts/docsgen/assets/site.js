(function () {
  "use strict";

  var STORAGE_KEY = "pretty-pdf-site-theme";
  var root = document.documentElement;

  function applyTheme(name) {
    root.setAttribute("data-site-theme", name);
    var buttons = document.querySelectorAll(".theme-swatch");
    for (var i = 0; i < buttons.length; i++) {
      var isActive = buttons[i].getAttribute("data-theme") === name;
      buttons[i].setAttribute("aria-pressed", isActive ? "true" : "false");
    }
    updateDownloadLink(name);
  }

  // Keeps the "Download these docs as a PDF" button pointed at the PDF
  // that matches whatever theme is currently on screen — docsgen
  // pre-renders one PDF per builtin theme (go-pretty-pdf-docs-<id>.pdf),
  // so this is just picking the right static file, not generating anything.
  function updateDownloadLink(name) {
    var link = document.getElementById("download-pdf-btn");
    var sub = document.getElementById("download-pdf-sub");
    if (link) link.setAttribute("href", "go-pretty-pdf-docs-" + name + ".pdf");
    if (sub) {
      var swatchLabel = document.querySelector('.theme-swatch[data-theme="' + name + '"] .theme-swatch-label');
      var displayName = swatchLabel ? swatchLabel.textContent : name;
      sub.textContent = "in the " + displayName + " theme — rendered by go-pretty-pdf itself";
    }
  }

  function initThemeSwitcher() {
    var saved = null;
    try {
      saved = localStorage.getItem(STORAGE_KEY);
    } catch (e) {
      /* localStorage unavailable (privacy mode) — fall back to default */
    }
    applyTheme(saved || root.getAttribute("data-site-theme") || "classic");

    var buttons = document.querySelectorAll(".theme-swatch");
    for (var i = 0; i < buttons.length; i++) {
      buttons[i].addEventListener("click", function () {
        var name = this.getAttribute("data-theme");
        applyTheme(name);
        try {
          localStorage.setItem(STORAGE_KEY, name);
        } catch (e) {
          /* ignore */
        }
      });
    }
  }

  function initScrollSpy() {
    var links = document.querySelectorAll(".sidebar-nav a");
    var linkByID = {};
    links.forEach(function (link) {
      linkByID[link.getAttribute("href").slice(1)] = link;
    });

    var sections = document.querySelectorAll(".section[id]");
    if (!("IntersectionObserver" in window) || sections.length === 0) return;

    var observer = new IntersectionObserver(
      function (entries) {
        entries.forEach(function (entry) {
          var link = linkByID[entry.target.id];
          if (!link) return;
          if (entry.isIntersecting) {
            links.forEach(function (l) { l.classList.remove("is-active"); });
            link.classList.add("is-active");
          }
        });
      },
      { rootMargin: "-10% 0px -75% 0px", threshold: 0 }
    );

    sections.forEach(function (section) { observer.observe(section); });
  }

  function isApplePlatform() {
    var platform = (navigator.userAgentData && navigator.userAgentData.platform) ||
      navigator.platform || navigator.userAgent || "";
    return /Mac|iPhone|iPad|iPod/i.test(platform);
  }

  function initShortcutHint() {
    var hint = document.getElementById("palette-shortcut-hint");
    if (!hint) return;
    hint.textContent = isApplePlatform() ? "⌘ K" : "Ctrl K";
  }

  function initNavToggle() {
    var toggle = document.getElementById("nav-toggle");
    var nav = document.getElementById("sidebar-nav");
    if (!toggle || !nav) return;

    function close() {
      nav.classList.remove("is-open");
      toggle.setAttribute("aria-expanded", "false");
    }

    toggle.addEventListener("click", function () {
      var open = nav.classList.toggle("is-open");
      toggle.setAttribute("aria-expanded", open ? "true" : "false");
    });

    nav.querySelectorAll("a").forEach(function (link) {
      link.addEventListener("click", close);
    });
  }

  function initCommandPalette() {
    var palette = document.getElementById("command-palette");
    var input = document.getElementById("command-palette-input");
    var results = document.getElementById("command-palette-results");
    var trigger = document.getElementById("palette-trigger");
    if (!palette || !input || !results || !trigger) return;

    var index = Array.prototype.map.call(
      document.querySelectorAll(".sidebar-nav a"),
      function (link) {
        return { title: link.textContent.trim(), href: link.getAttribute("href") };
      }
    );

    var selectedIndex = 0;
    var visible = [];

    function render(items) {
      visible = items;
      selectedIndex = 0;
      if (items.length === 0) {
        results.innerHTML = '<li class="command-palette-empty">No matching section.</li>';
        return;
      }
      results.innerHTML = items
        .map(function (item, i) {
          return (
            '<li class="' + (i === 0 ? "is-selected" : "") + '" data-index="' + i + '">' +
            '<a href="' + item.href + '">' + item.title + "</a></li>"
          );
        })
        .join("");
    }

    function filter(query) {
      var q = query.trim().toLowerCase();
      if (!q) return render(index);
      var matches = index
        .map(function (item) {
          return { item: item, pos: item.title.toLowerCase().indexOf(q) };
        })
        .filter(function (m) { return m.pos !== -1; })
        .sort(function (a, b) { return a.pos - b.pos; })
        .map(function (m) { return m.item; });
      render(matches);
    }

    function updateSelection(next) {
      var items = results.querySelectorAll("li");
      if (items.length === 0) return;
      items[selectedIndex] && items[selectedIndex].classList.remove("is-selected");
      selectedIndex = (next + items.length) % items.length;
      items[selectedIndex].classList.add("is-selected");
      items[selectedIndex].scrollIntoView({ block: "nearest" });
    }

    function open() {
      palette.hidden = false;
      document.body.style.overflow = "hidden";
      input.value = "";
      render(index);
      setTimeout(function () { input.focus(); }, 0);
    }

    function close() {
      palette.hidden = true;
      document.body.style.overflow = "";
      trigger.focus();
    }

    function commit() {
      var item = visible[selectedIndex];
      if (!item) return;
      close();
      var target = document.querySelector(item.href);
      if (target) target.scrollIntoView({ block: "start" });
      history.replaceState(null, "", item.href);
    }

    trigger.addEventListener("click", open);

    palette.querySelectorAll("[data-palette-close]").forEach(function (el) {
      el.addEventListener("click", close);
    });

    results.addEventListener("click", function (e) {
      var link = e.target.closest("a");
      if (!link) return;
      e.preventDefault();
      var li = link.closest("li");
      selectedIndex = Array.prototype.indexOf.call(results.children, li);
      commit();
    });

    input.addEventListener("input", function () { filter(input.value); });

    input.addEventListener("keydown", function (e) {
      if (e.key === "ArrowDown") { e.preventDefault(); updateSelection(selectedIndex + 1); }
      else if (e.key === "ArrowUp") { e.preventDefault(); updateSelection(selectedIndex - 1); }
      else if (e.key === "Enter") { e.preventDefault(); commit(); }
      else if (e.key === "Escape") { e.preventDefault(); close(); }
    });

    document.addEventListener("keydown", function (e) {
      var isMod = e.metaKey || e.ctrlKey;
      if (isMod && e.key.toLowerCase() === "k") {
        e.preventDefault();
        if (palette.hidden) open(); else close();
      }
    });
  }

  document.addEventListener("DOMContentLoaded", function () {
    initThemeSwitcher();
    initScrollSpy();
    initNavToggle();
    initCommandPalette();
    initShortcutHint();
  });
})();
