package render

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/chromedp/chromedp"
)

// Severity classifies an audit Issue. Only SeverityWarning exists today —
// the audit is advisory and never fails a build — but the type leaves room
// for a future SeverityError without an API break.
type Severity string

const SeverityWarning Severity = "warning"

// Issue is one finding from a post-compose/pre-print visual/structural
// audit of the document: something that renders but is likely wrong —
// clipped, overlapping, unreadable, or missing — rather than a hard error.
type Issue struct {
	// Check is a short, stable, machine-readable identifier for the rule
	// that produced this issue (e.g. "overflow-x", "low-contrast",
	// "heading-clip-risk"), suitable for filtering or documentation links.
	Check    string
	Severity Severity
	Message  string
}

// AuditReport collects every Issue found while rendering a single document.
// A nil report or one with no Issues means the audit found nothing to flag
// — it does not guarantee the PDF is perfect, only that none of the checks
// it runs caught a problem.
type AuditReport struct {
	Issues []Issue
}

// HasIssues reports whether the audit found anything worth surfacing to the
// caller. Safe to call on a nil report.
func (r *AuditReport) HasIssues() bool {
	return r != nil && len(r.Issues) > 0
}

// domAuditJS runs inside the already-navigated document, before
// Page.printToPDF, and returns a JSON array of {check, message} objects.
// It checks what's actually observable from the DOM/CSSOM at this stage:
//
//   - overflow-x: content wider than its own box (long code lines, wide
//     tables, oversized images) that print will most likely clip instead
//     of wrapping, since printed pages don't get a horizontal scrollbar.
//   - broken-image: an <img> whose source never resolved to real pixels.
//   - empty-content: the whole document has next to no visible text,
//     usually a sign composition/parsing silently produced nothing.
//   - low-contrast: visible text whose color is too close to its
//     effective background to read comfortably (a common symptom of a
//     custom --color-* override clashing with a theme's own palette).
//   - heading-clip-risk: an element that forces a page break
//     (page-break-before/break-before) but doesn't keep enough top margin
//     to clear the ~0.3in strip chrome-headless-shell silently clips
//     whenever a header or page-number footer is displayed — see
//     TestBaseCSSH1HasTopMarginBuffer and the CHANGELOG entry it guards.
//     This mirrors a real, previously-shipped bug so custom themes/CSS get
//     the same protection builtin themes now have.
//
// It deliberately cannot see the two things that live purely in Chrome's
// print engine rather than the DOM: the fixed ~0.2in header/footer inset,
// and how the browser's print pagination actually slices this content
// into pages. Those are covered by base.css's own layout rules (and their
// regression tests), not by this runtime audit.
const domAuditJS = `(() => {
  const issues = [];
  const needsHeaderFooter = %t;

  function pushIssue(check, message) {
    issues.push({check, message});
  }

  document.querySelectorAll('pre, code, table, img, .component-deep-dive, .component-warning, .component-axiom').forEach(el => {
    if (el.scrollWidth > el.clientWidth + 2) {
      const label = el.id ? ('#' + el.id) : ('<' + el.tagName.toLowerCase() + '>');
      pushIssue('overflow-x', label + ' is wider than its box (' + el.scrollWidth + 'px vs ' + el.clientWidth + 'px) and may be clipped when printed');
    }
  });

  document.querySelectorAll('img').forEach(img => {
    if (img.complete && img.naturalWidth === 0) {
      const src = (img.getAttribute('src') || '(no src)').slice(0, 80);
      pushIssue('broken-image', 'image failed to load: ' + src);
    }
  });

  const textLen = (document.body.innerText || '').trim().length;
  if (textLen < 20) {
    pushIssue('empty-content', 'document has almost no visible text (' + textLen + ' characters) — composition or rendering may have failed');
  }

  function parseColor(str) {
    const m = str && str.match(/rgba?\(([^)]+)\)/);
    if (!m) return null;
    const parts = m[1].split(',').map(s => parseFloat(s.trim()));
    if (parts.length < 3 || parts.some(isNaN)) return null;
    return {r: parts[0], g: parts[1], b: parts[2], a: parts.length > 3 ? parts[3] : 1};
  }
  function luminance(c) {
    const chan = v => {
      v /= 255;
      return v <= 0.03928 ? v / 12.92 : Math.pow((v + 0.055) / 1.055, 2.4);
    };
    return 0.2126 * chan(c.r) + 0.7152 * chan(c.g) + 0.0722 * chan(c.b);
  }
  function effectiveBg(el) {
    let node = el;
    while (node) {
      const bg = parseColor(getComputedStyle(node).backgroundColor);
      if (bg && bg.a > 0.01) return bg;
      node = node.parentElement;
    }
    return {r: 255, g: 255, b: 255, a: 1};
  }

  const seenContrast = new Set();
  let contrastIssues = 0;
  const walker = document.createTreeWalker(document.body, NodeFilter.SHOW_TEXT);
  let node;
  while ((node = walker.nextNode()) && contrastIssues < 5) {
    const text = node.textContent.trim();
    if (text.length < 2) continue;
    const el = node.parentElement;
    if (!el) continue;
    const style = getComputedStyle(el);
    if (style.visibility === 'hidden' || style.display === 'none' || parseFloat(style.opacity) === 0) continue;
    const fg = parseColor(style.color);
    if (!fg || fg.a < 0.5) continue;
    const bg = effectiveBg(el);
    const ratio = (Math.max(luminance(fg), luminance(bg)) + 0.05) / (Math.min(luminance(fg), luminance(bg)) + 0.05);
    const key = style.color + '|' + JSON.stringify(bg);
    if (ratio < 2.2 && !seenContrast.has(key)) {
      seenContrast.add(key);
      contrastIssues++;
      pushIssue('low-contrast', 'text "' + text.slice(0, 40) + '" has a low contrast ratio (' + ratio.toFixed(2) + ':1) against its background and may be hard to read');
    }
  }

  if (needsHeaderFooter) {
    document.querySelectorAll('h1, h2, h3, h4, h5').forEach(h => {
      const style = getComputedStyle(h);
      const breaksPage = style.pageBreakBefore === 'always' || style.breakBefore === 'page';
      if (!breaksPage) return;
      const marginTopIn = parseFloat(style.marginTop) / 96;
      if (marginTopIn < 0.3) {
        const label = (h.textContent || '').trim().slice(0, 40);
        pushIssue('heading-clip-risk', '<' + h.tagName.toLowerCase() + '> "' + label + '" forces a page break but has only ' + marginTopIn.toFixed(2) + 'in of top margin — content flush against a forced page break is clipped by roughly the first 0.3in when a header or page numbers are shown; give it more margin-top');
      }
    });
  }

  return JSON.stringify(issues);
})()`

type domIssue struct {
	Check   string `json:"check"`
	Message string `json:"message"`
}

// runDOMAudit evaluates domAuditJS in the page currently loaded at ctx —
// meant to be called from inside a chromedp.ActionFunc that's already part
// of an in-progress chromedp.Tasks run, after Navigate and before
// PrintToPDF — and converts its findings into Issues. Any failure to run
// or parse the audit script itself is treated as non-fatal — the audit is
// advisory, and a bug in it must never break an otherwise-successful
// render.
func runDOMAudit(ctx context.Context, needsHeaderFooter bool) []Issue {
	var raw string
	script := fmt.Sprintf(domAuditJS, needsHeaderFooter)
	if err := chromedp.Evaluate(script, &raw).Do(ctx); err != nil {
		return nil
	}

	var found []domIssue
	if err := json.Unmarshal([]byte(raw), &found); err != nil {
		return nil
	}

	issues := make([]Issue, 0, len(found))
	for _, f := range found {
		issues = append(issues, Issue{Check: f.Check, Severity: SeverityWarning, Message: f.Message})
	}
	return issues
}

var pdfPageObjectRe = regexp.MustCompile(`/Type\s*/Page\b`)

// countPDFPages counts `/Type /Page` objects directly in the raw PDF bytes
// — a small, dependency-free heuristic (no PDF parsing library needed)
// that matches chrome-headless-shell's uncompressed object output, and
// deliberately excludes `/Type /Pages` (the page-tree node) via the word
// boundary after "Page".
func countPDFPages(pdfBuf []byte) int {
	return len(pdfPageObjectRe.FindAll(pdfBuf, -1))
}

// auditPDFBytes runs structural sanity checks on the finished PDF that can
// only be done once it exists — currently just "did it end up with at
// least one page" — as a last-resort guard against a render that
// technically succeeded (no error) but produced a corrupt or empty file.
func auditPDFBytes(pdfBuf []byte) []Issue {
	if countPDFPages(pdfBuf) == 0 {
		return []Issue{{
			Check:    "page-count",
			Severity: SeverityWarning,
			Message:  "could not find any page objects in the generated PDF — the output file may be empty or corrupt",
		}}
	}
	return nil
}
