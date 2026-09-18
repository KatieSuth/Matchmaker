import { Locator, Page } from "@playwright/test";

/** Clicks a react-select placeholder, then the visible option label. */
export async function pickSelectOption(page: Page, placeholder: string, option: string): Promise<void> {
  // Current/peak rank (and similar paired fields) share a placeholder; take the first unmatched control.
  await page.getByText(placeholder, { exact: true }).first().click();
  await page.getByRole("option", { name: option, exact: true }).click();
}

/** Unique event title so parallel-ish local reruns do not collide in assertions. */
export function uniqueEventName(): string {
  return `E2E Event ${Date.now()}`;
}

export function eventHeading(page: Page, name: string): Locator {
  return page.getByRole("heading", { name, exact: true });
}
