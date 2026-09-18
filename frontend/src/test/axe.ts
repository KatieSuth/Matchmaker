// Shared axe-core config for component a11y tests. Disables `color-contrast`, which jsdom can't
// meaningfully evaluate (no real layout/paint engine — it also depends on canvas, which jsdom
// doesn't implement and would otherwise print "Not implemented: HTMLCanvasElement's getContext()"
// noise for every check). Real contrast regressions are better caught visually/in Playwright.
import { configureAxe } from "vitest-axe";

export const axe = configureAxe({
  rules: {
    "color-contrast": { enabled: false },
  },
});
