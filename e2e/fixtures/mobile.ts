// Generic phone-width layout audit. jsdom (Vitest) has no layout engine, so these checks
// measure real boxes in the browser instead of asserting on class names.
import { expect, Page } from "@playwright/test";

export interface MobileAuditOptions {
  /** CSS selector to audit instead of the page; defaults to the open modal sheet, else <body>. */
  root?: string;
  /** CSS selector for elements whose overlap is intentional (e.g. a badge pinned over an avatar). */
  ignore?: string;
}

/** Waits for finite CSS animations (sheet rise, menu drop-in) to settle so boxes are final. */
async function waitForAnimations(page: Page): Promise<void> {
  await page.waitForFunction(() =>
    document
      .getAnimations()
      .every((a) => a.playState !== "running" || a.effect?.getTiming().iterations === Infinity),
  );
}

/**
 * Asserts the current view is usable at phone width:
 * - no horizontal page scroll;
 * - no visible text or control extends past the viewport's left/right edge
 *   (content inside an intentional horizontal scroller is exempt);
 * - no two controls overlap, and no text overlaps a control it isn't part of.
 * Decorative content (aria-hidden, invisible, ≤1px like sr-only) is skipped. When a modal sheet
 * is open only the sheet is audited, since it intentionally covers the page behind it.
 */
export async function expectMobileFriendly(page: Page, label: string, options: MobileAuditOptions = {}): Promise<void> {
  await waitForAnimations(page);
  const issues = await page.evaluate(({ rootSelector, ignoreSelector }) => {
    // Range rects use font ascent/descent, which can exceed tight line-heights by a few px.
    const TOLERANCE = 4;
    const vw = document.documentElement.clientWidth;
    const found: string[] = [];

    const describe = (el: Element): string => {
      const name = el.getAttribute("aria-label") ?? (el.textContent ?? "").trim().replace(/\s+/g, " ").slice(0, 40);
      return `<${el.tagName.toLowerCase()}${name ? ` "${name}"` : ""}>`;
    };

    const isHidden = (el: Element): boolean => {
      for (let n: Element | null = el; n; n = n.parentElement) {
        const aria = n.getAttribute("aria-hidden");
        if (aria !== null && aria !== "false") return true;
        const cs = getComputedStyle(n);
        if (cs.display === "none" || cs.visibility === "hidden" || cs.opacity === "0") return true;
      }
      return false;
    };

    const inHorizontalScroller = (el: Element): boolean => {
      for (let n = el.parentElement; n && n !== document.body; n = n.parentElement) {
        const ox = getComputedStyle(n).overflowX;
        if ((ox === "auto" || ox === "scroll") && n.scrollWidth > n.clientWidth) return true;
      }
      return false;
    };

    const overlaps = (a: DOMRect, b: DOMRect): boolean =>
      Math.min(a.right, b.right) - Math.max(a.left, b.left) > TOLERANCE &&
      Math.min(a.bottom, b.bottom) - Math.max(a.top, b.top) > TOLERANCE;

    const contains = (outer: DOMRect, inner: DOMRect): boolean =>
      inner.left >= outer.left && inner.right <= outer.right && inner.top >= outer.top && inner.bottom <= outer.bottom;

    const usable = (r: DOMRect): boolean => r.width > 1 && r.height > 1;

    if (document.documentElement.scrollWidth > vw || document.body.scrollWidth > vw) {
      found.push(`page scrolls horizontally (scrollWidth ${document.documentElement.scrollWidth} > ${vw})`);
    }

    const modals = document.querySelectorAll('[aria-modal="true"]');
    const root = rootSelector
      ? document.querySelector(rootSelector)
      : (modals[modals.length - 1] ?? document.body);
    if (!root) return [`audit root ${rootSelector} not found`];

    const ignored = (el: Element): boolean => !!ignoreSelector && !!el.closest(ignoreSelector);
    const viewportArea = window.innerWidth * window.innerHeight;

    const controlSelector =
      'button, a[href], input:not([type="hidden"]), select, textarea, [role="button"], [role="switch"], [role="tab"], [role="menuitem"], [role="combobox"]';
    // Per-fragment rects: an inline link wrapping across lines has one rect per line, whereas its
    // bounding box would cover the surrounding paragraph text.
    const controls = Array.from(root.querySelectorAll(controlSelector))
      .filter((el) => !isHidden(el) && !ignored(el))
      .flatMap((el) => Array.from(el.getClientRects()).map((rect) => ({ el, rect })))
      // Skip tiny elements (sr-only, react-select's auto-sized input) and full-screen backdrops.
      .filter(({ rect }) => usable(rect) && rect.width * rect.height < viewportArea * 0.8);

    const isTextField = (el: Element) => el.matches("input, textarea, select, [role='combobox']");

    // One rect per rendered line of each text node that isn't part of a control.
    const texts: { el: Element; rect: DOMRect }[] = [];
    const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT);
    for (let node = walker.nextNode(); node; node = walker.nextNode()) {
      const parent = node.parentElement;
      if (!parent || !node.textContent?.trim()) continue;
      if (parent.closest("script, style, noscript") || isHidden(parent) || ignored(parent)) continue;
      const range = document.createRange();
      range.selectNodeContents(node);
      for (const rect of Array.from(range.getClientRects())) {
        if (usable(rect)) texts.push({ el: parent, rect });
      }
    }

    for (const { el, rect } of [...texts, ...controls]) {
      if (rect.left >= -1 && rect.right <= vw + 1) continue;
      if (inHorizontalScroller(el)) continue;
      found.push(`${describe(el)} extends past the viewport (x ${Math.round(rect.left)}–${Math.round(rect.right)}, viewport ${vw})`);
    }

    for (let i = 0; i < controls.length; i++) {
      for (let j = i + 1; j < controls.length; j++) {
        const a = controls[i];
        const b = controls[j];
        if (a.el.contains(b.el) || b.el.contains(a.el)) continue;
        if (!overlaps(a.rect, b.rect)) continue;
        // In-field adornments (clear buttons, toggles inside a text field) sit inside the field on purpose.
        if ((isTextField(a.el) && contains(a.rect, b.rect)) || (isTextField(b.el) && contains(b.rect, a.rect))) continue;
        found.push(`${describe(a.el)} overlaps ${describe(b.el)}`);
      }
    }

    for (const text of texts) {
      for (const control of controls) {
        // Custom selects render placeholder/value text over their input on purpose.
        if (isTextField(control.el) || control.el.contains(text.el)) continue;
        if (text.el.closest(controlSelector)) continue;
        if (overlaps(text.rect, control.rect)) {
          found.push(`text in ${describe(text.el)} overlaps ${describe(control.el)}`);
        }
      }
    }

    return Array.from(new Set(found));
  }, { rootSelector: options.root ?? null, ignoreSelector: options.ignore ?? null });

  expect(issues, `${label}: mobile layout issues`).toEqual([]);
}
