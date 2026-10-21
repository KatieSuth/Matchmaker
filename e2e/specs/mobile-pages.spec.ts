// Phone-width audit of every page and the main sheets/menus: no sideways scroll, nothing cut
// off at the screen edge, and no overlapping text/controls (see fixtures/mobile.ts).
import { expect, Page, test } from "@playwright/test";
import {
  bootstrapPlayerWithValorant,
  createValorantEventApi,
  fetchEventGroupApi,
  fetchValorantCatalog,
  issueTestLogin,
  registerForGroupApi,
  TestUser,
} from "../fixtures/api";
import { loginAsExistingUser, loginAsNewUser, loginViaCallback, uniqueUser } from "../fixtures/auth";
import { expectMobileFriendly } from "../fixtures/mobile";
import { eventHeading, pickSelectOption } from "../fixtures/ui";

const LOCK_IN_PLAYERS = 10;

/** Creates a Valorant event with a long (wrapping) title owned by a fresh host. */
async function createHostedEvent(): Promise<{ host: TestUser; groupId: string; eventName: string }> {
  const host = uniqueUser("mhost");
  const token = await bootstrapPlayerWithValorant(host);
  const { game } = await fetchValorantCatalog(token);
  const eventName = `Mobile audit ${Date.now()} long title wraps`;
  const groupId = await createValorantEventApi(token, game.id, eventName);
  return { host, groupId, eventName };
}

/** Logs an already-onboarded user into the browser (lands on My Events). */
async function loginAs(page: Page, user: TestUser): Promise<void> {
  const { otc } = await issueTestLogin(user);
  await loginViaCallback(page, otc, false);
}

/** Registers `count` fresh API-only players for every game in the group. */
async function seedRegistrations(groupId: string, count: number): Promise<void> {
  await Promise.all(
    Array.from({ length: count }, async (_, i) => {
      const token = await bootstrapPlayerWithValorant(uniqueUser(`mfill${i}`));
      await registerForGroupApi(token, await fetchEventGroupApi(token, groupId));
    }),
  );
}

test.describe("mobile pages", () => {
  test("logged out: landing and about", async ({ page }) => {
    await page.goto("/");
    await expect(page.getByRole("link", { name: "Log in with Discord" })).toBeVisible();
    await expectMobileFriendly(page, "landing");

    await page.goto("/about");
    await expect(page.getByRole("heading", { name: "About Matchmaker" })).toBeVisible();
    await expectMobileFriendly(page, "about");
  });

  test("new user: profile setup with a game row", async ({ page }) => {
    await loginAsNewUser(page, uniqueUser("mprof"));
    await expect(page.getByRole("button", { name: "Save Profile" })).toBeVisible();
    await expectMobileFriendly(page, "profile setup (empty)");

    await page.getByRole("button", { name: "Add game" }).click();
    await pickSelectOption(page, "— Select a game —", "Valorant");
    await expect(page.getByText("— Select rank —").first()).toBeVisible();
    await expectMobileFriendly(page, "profile setup (game row)");
  });

  test("existing user: empty My Events, account menu, and host-an-event sheet", async ({ page }) => {
    await loginAsExistingUser(page, uniqueUser("mempty"), true);
    await expect(page.getByRole("heading", { name: "My Events" })).toBeVisible();
    await expectMobileFriendly(page, "my events (empty)");

    // The avatar button's name label is hidden at phone width, so target it by its menu popup.
    await page.locator('button[aria-haspopup="menu"]').click();
    await expect(page.getByRole("menu")).toBeVisible();
    await expectMobileFriendly(page, "account menu", { root: '[role="menu"]' });
    await page.keyboard.press("Escape");
    await page.getByRole("heading", { name: "My Events" }).click();

    await page.getByRole("button", { name: /Host an event/ }).click();
    const sheet = page.getByRole("dialog", { name: "Host an event" });
    await expect(sheet).toBeVisible();
    await expectMobileFriendly(page, "host-an-event sheet (empty)");

    await pickSelectOption(page, "Select game", "Valorant");
    await expect(page.getByText("Select game mode")).toBeVisible();
    await expectMobileFriendly(page, "host-an-event sheet (game selected)");
  });

  test("host: My Events list, event page, and settings sheet", async ({ page }) => {
    const { host, groupId, eventName } = await createHostedEvent();
    await loginAs(page, host);
    await page.getByRole("button", { name: "Hosted by Me" }).click();
    await expect(page.getByText(eventName).first()).toBeVisible();
    await expectMobileFriendly(page, "my events (hosted list)");

    await page.goto(`/event/${groupId}`);
    await expect(eventHeading(page, eventName)).toBeVisible();
    await expectMobileFriendly(page, "event page (host, open)");

    await page.getByRole("button", { name: "Event settings" }).click();
    await expect(page.getByRole("button", { name: "Save Settings" })).toBeVisible();
    await expectMobileFriendly(page, "event settings sheet");
  });

  test("guest: registration editor sheet, then registered event page", async ({ page }) => {
    const { groupId, eventName } = await createHostedEvent();
    const guest = uniqueUser("mguest");
    await bootstrapPlayerWithValorant(guest);
    await loginAs(page, guest);

    // Unregistered guests get the registration editor auto-opened.
    await page.goto(`/event/${groupId}`);
    await expect(page.getByRole("button", { name: "Save Registration" })).toBeVisible();
    await expectMobileFriendly(page, "registration editor sheet");

    await page.getByRole("button", { name: "Save Registration" }).click();
    const registered = page.getByRole("dialog", { name: "You're registered!" });
    await expect(registered).toBeVisible();
    await expectMobileFriendly(page, "registered confirmation sheet");
    await registered.getByRole("button", { name: "Not now" }).click();
    await expect(registered).toHaveCount(0);

    await expect(eventHeading(page, eventName)).toBeVisible();
    await expect(page.getByRole("button", { name: "Edit My Registration" })).toBeVisible();
    await expectMobileFriendly(page, "event page (registered player)");
  });

  test("host: locked-in event with teams and join-lobby sheet", async ({ page }) => {
    test.setTimeout(180_000);
    const { host, groupId, eventName } = await createHostedEvent();
    await seedRegistrations(groupId, LOCK_IN_PLAYERS);
    await loginAs(page, host);

    await page.goto(`/event/${groupId}`);
    await expect(eventHeading(page, eventName)).toBeVisible();
    await expectMobileFriendly(page, "event page (host, full roster)");

    await page.getByRole("button", { name: "Lock In & Create Teams" }).click();
    const okay = page.getByRole("button", { name: "Okay" });
    if (await okay.isVisible().catch(() => false)) {
      await okay.click();
    }
    await expect(page.getByRole("button", { name: "Delete teams" })).toBeVisible();
    await expect(page.getByRole("button", { name: "Copy Discord Pings" })).toBeVisible();
    await expectMobileFriendly(page, "event page (host, teams created)");

    await page.getByRole("button", { name: "Join Lobby" }).first().click();
    await expect(page.getByRole("heading", { name: /Join Lobby 1/ })).toBeVisible();
    await expectMobileFriendly(page, "join lobby sheet");
  });
});
