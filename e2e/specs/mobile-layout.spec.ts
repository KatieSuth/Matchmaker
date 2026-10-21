// Phone-width layout checks. jsdom (Vitest) has no layout engine, so overlap / stacking
// regressions can only be caught by measuring real boxes in a browser.
import { expect, Locator, Page, test } from "@playwright/test";
import {
  bootstrapPlayerWithValorant,
  createValorantEventApi,
  fetchEventGroupApi,
  fetchValorantCatalog,
  issueTestLogin,
  registerForGroupApi,
  TestUser,
} from "../fixtures/api";
import { loginViaCallback, uniqueUser } from "../fixtures/auth";
import { eventHeading } from "../fixtures/ui";

type Box = { x: number; y: number; width: number; height: number };

/** Returns the element's bounding box, failing the test if it is not rendered. */
async function boxOf(locator: Locator): Promise<Box> {
  await expect(locator).toBeVisible();
  const box = await locator.boundingBox();
  if (!box) {
    throw new Error(`no bounding box for ${locator}`);
  }
  return box;
}

const bottom = (b: Box) => b.y + b.height;

/** Asserts the page cannot be scrolled sideways at the current viewport. */
async function expectNoHorizontalScroll(page: Page): Promise<void> {
  const overflow = await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth);
  expect(overflow).toBeLessThanOrEqual(0);
}

/** Creates a Valorant event with a long title owned by a fresh host; returns the host and the event. */
async function createHostedEvent(): Promise<{ host: TestUser; groupId: string; eventName: string }> {
  const host = uniqueUser("mhost");
  const hostToken = await bootstrapPlayerWithValorant(host);
  const { game } = await fetchValorantCatalog(hostToken);
  // Long enough to wrap at phone width, which is what used to squeeze the title beside the buttons.
  const eventName = `Mobile header ${Date.now()} long title wraps`;
  const groupId = await createValorantEventApi(hostToken, game.id, eventName);
  return { host, groupId, eventName };
}

/** Logs the user into the browser and opens the event page. */
async function openEventAs(page: Page, user: TestUser, groupId: string): Promise<void> {
  const { otc } = await issueTestLogin(user);
  await loginViaCallback(page, otc, false);
  await page.goto(`/event/${groupId}`);
}

/**
 * Asserts the event header card is stacked for phone width: title (full width) → meta →
 * first event → every named action button (inside the card) → status/team size/host,
 * with no horizontal page scroll.
 */
async function expectStackedHeader(page: Page, eventName: string, actionNames: string[]): Promise<void> {
  const title = eventHeading(page, eventName);
  const card = page.locator(".card", { has: title }).first();
  const meta = card.getByText("Valorant · 5v5 · AMER", { exact: true });
  const firstEvent = card.getByText(/^First event:/);
  const status = card.getByText("Status", { exact: true });
  const actions = actionNames.map((name) => card.getByRole("button", { name, exact: true }));

  const [cardBox, titleBox, metaBox, firstBox, statusBox, ...actionBoxes] = await Promise.all(
    [card, title, meta, firstEvent, status, ...actions].map(boxOf),
  );

  // Title spans the card instead of being squeezed into a narrow column beside the actions.
  expect(titleBox.width).toBeGreaterThan(cardBox.width * 0.8);

  expect(bottom(titleBox)).toBeLessThanOrEqual(metaBox.y);
  expect(bottom(metaBox)).toBeLessThanOrEqual(firstBox.y);
  for (const action of actionBoxes) {
    expect(bottom(firstBox)).toBeLessThanOrEqual(action.y);
    expect(bottom(action)).toBeLessThanOrEqual(statusBox.y);
    // Actions stay inside the card even when they wrap.
    expect(action.x).toBeGreaterThanOrEqual(cardBox.x);
    expect(action.x + action.width).toBeLessThanOrEqual(cardBox.x + cardBox.width);
  }

  await expectNoHorizontalScroll(page);
}

test.describe("mobile layout", () => {
  test("host: event header stacks title, meta, first event, actions, then summary", async ({ page }) => {
    const { host, groupId, eventName } = await createHostedEvent();
    await openEventAs(page, host, groupId);
    await expectStackedHeader(page, eventName, ["Lock In & Create Teams", "Copy share link", "Event settings"]);
  });

  test("registered player: event header stacks with calendar action and no host controls", async ({ page }) => {
    const { groupId, eventName } = await createHostedEvent();
    const player = uniqueUser("mplayer");
    const playerToken = await bootstrapPlayerWithValorant(player);
    await registerForGroupApi(playerToken, await fetchEventGroupApi(playerToken, groupId));

    await openEventAs(page, player, groupId);
    await expectStackedHeader(page, eventName, [
      "Copy share link",
      "Download calendar invitation",
      "Event settings",
    ]);
    await expect(page.getByRole("button", { name: "Lock In & Create Teams" })).toHaveCount(0);
  });
});
