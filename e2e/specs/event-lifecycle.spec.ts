import { Browser, expect, Page, test } from "@playwright/test";
import { bootstrapPlayerWithValorant, fetchEventGroupApi, registerForGroupApi } from "../fixtures/api";
import { loginAsExistingUser, uniqueUser } from "../fixtures/auth";
import { eventHeading, pickSelectOption, uniqueEventName } from "../fixtures/ui";

const LOCK_IN_PLAYERS = 10;

async function newAuthedPage(browser: Browser, withValorant: boolean) {
  const context = await browser.newContext();
  const page = await context.newPage();
  const user = uniqueUser(withValorant ? "guest" : "host");
  await loginAsExistingUser(page, user, withValorant);
  return { context, page, user };
}

async function createValorantEvent(page: Page, name: string): Promise<string> {
  await page.getByRole("button", { name: /Host an event/ }).click();
  await expect(page.getByRole("heading", { name: "Host an event" })).toBeVisible();
  await expect(page.getByText("Select game", { exact: true })).toBeVisible();

  await page.getByPlaceholder("Optional custom title").fill(name);
  await pickSelectOption(page, "Select game", "Valorant");
  await expect(page.getByText("Select game mode")).toBeVisible();
  await pickSelectOption(page, "Select game mode", "5v5");
  await pickSelectOption(page, "Select region", "AMER");
  await page.getByRole("button", { name: "Create Event" }).click();
  await expect(eventHeading(page, name)).toBeVisible();
  const url = page.url();
  const match = url.match(/\/event\/([^/?#]+)/);
  if (!match) {
    throw new Error(`expected event URL, got ${url}`);
  }
  return match[1];
}

async function seedFillerRegistrations(groupId: string, alreadyRegistered: number): Promise<void> {
  const needed = LOCK_IN_PLAYERS - alreadyRegistered;
  await Promise.all(
    Array.from({ length: needed }, async (_, i) => {
      const token = await bootstrapPlayerWithValorant(uniqueUser(`fill${i}`));
      const group = await fetchEventGroupApi(token, groupId);
      await registerForGroupApi(token, group);
    }),
  );
}

test.describe("event lifecycle", () => {
  test("host creates, guest registers, lock-in hides host controls, join lobby, edit, delete", async ({
    browser,
  }) => {
    test.setTimeout(180_000);
    const [host, guest] = await Promise.all([newAuthedPage(browser, false), newAuthedPage(browser, true)]);
    const eventName = uniqueEventName();

    const groupId = await createValorantEvent(host.page, eventName);

    await guest.page.goto(`/event/${groupId}`);
    await expect(eventHeading(guest.page, eventName)).toBeVisible();
    // Guests who are not yet registered get the editor auto-opened (same as Register Now).
    await guest.page.getByRole("button", { name: "Save Registration" }).click();
    await expect(guest.page.getByRole("button", { name: "Edit My Registration" })).toBeVisible();

    await seedFillerRegistrations(groupId, 1);

    await host.page.reload();
    await expect(host.page.getByRole("button", { name: "Lock In & Create Teams" })).toBeEnabled();
    await host.page.getByRole("button", { name: "Lock In & Create Teams" }).click();
    const okay = host.page.getByRole("button", { name: "Okay" });
    if (await okay.isVisible().catch(() => false)) {
      await okay.click();
    }
    await expect(host.page.getByRole("button", { name: "Delete teams" })).toBeVisible();
    await expect(host.page.getByRole("button", { name: "Join Lobby" })).toBeVisible();

    await guest.page.reload();
    await expect(guest.page.getByRole("heading", { name: eventName, exact: true })).toBeVisible();
    await expect(guest.page.getByRole("button", { name: "Lock In & Create Teams" })).toHaveCount(0);
    await expect(guest.page.getByRole("button", { name: "Delete teams" })).toHaveCount(0);
    await expect(guest.page.getByRole("button", { name: "Copy Discord Pings" })).toHaveCount(0);

    await host.page.getByRole("button", { name: "Join Lobby" }).click();
    await host.page.getByPlaceholder("Code or https://gg.riotgames.com/…").fill("ABC-123");
    await host.page.getByRole("button", { name: "Save" }).click();
    await expect(host.page.getByPlaceholder("Code or https://gg.riotgames.com/…")).toHaveCount(0);

    await guest.page.reload();
    await guest.page.getByRole("button", { name: "Join Lobby" }).click();
    await expect(guest.page.getByRole("heading", { name: /Join Lobby 1/ })).toBeVisible();
    const joinCodeInput = guest.page.getByPlaceholder("Code or https://gg.riotgames.com/…");
    // Lobby hosts see an editable field; other players see the code as static text.
    if (await joinCodeInput.isVisible()) {
      await expect(joinCodeInput).toHaveValue("ABC-123");
    } else {
      await expect(guest.page.getByText("ABC-123", { exact: true })).toBeVisible();
    }
    await guest.page.getByRole("button", { name: "Close", exact: true }).filter({ hasText: "Close" }).click();

    const renamed = `${eventName} edited`;
    await host.page.getByRole("button", { name: "Event settings" }).click();
    await host.page.getByPlaceholder("Optional custom title").fill(renamed);
    await host.page.getByRole("button", { name: "Save Settings" }).click();
    await expect(eventHeading(host.page, renamed)).toBeVisible();

    await host.page.getByRole("button", { name: "Event settings" }).click();
    await host.page.getByRole("button", { name: "Delete Event" }).click();
    await host.page.getByRole("button", { name: "Delete Permanently" }).click();
    await expect(host.page.getByRole("heading", { name: "My Events" })).toBeVisible();

    await host.context.close();
    await guest.context.close();
  });
});
