import { Page } from "@playwright/test";
import { bootstrapExistingUser, issueTestLogin, TestUser } from "./api";

/** Completes browser auth by exchanging the OTC on /auth/callback (sets cookies + JWT). */
export async function loginViaCallback(page: Page, otc: string, newUser: boolean): Promise<void> {
  await page.goto(`/auth/callback?otc=${encodeURIComponent(otc)}&new_user=${newUser}`);
  if (newUser) {
    await page.waitForURL("**/my_account");
    return;
  }
  await page.waitForURL("**/my_events");
}

/** First-time Discord bypass: lands on profile setup. */
export async function loginAsNewUser(page: Page, user: TestUser): Promise<void> {
  const { otc, new_user } = await issueTestLogin(user);
  if (!new_user) {
    throw new Error(`expected a new user for ${user.discordId}`);
  }
  await loginViaCallback(page, otc, true);
}

/** Onboards via API, then logs in through the callback so the app skips /my_account. */
export async function loginAsExistingUser(page: Page, user: TestUser, withValorant = false): Promise<void> {
  await bootstrapExistingUser(user, withValorant);
  const { otc, new_user } = await issueTestLogin(user);
  if (new_user) {
    throw new Error(`expected an existing user for ${user.discordId}`);
  }
  await loginViaCallback(page, otc, false);
}

export function uniqueUser(prefix: string): TestUser {
  const stamp = `${Date.now()}-${Math.floor(Math.random() * 1_000_000)}`;
  return {
    discordId: `${prefix}-${stamp}`,
    username: `${prefix}${stamp.slice(-6)}`,
    globalName: `${prefix} ${stamp.slice(-4)}`,
  };
}
