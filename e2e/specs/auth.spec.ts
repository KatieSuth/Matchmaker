import { expect, test } from "@playwright/test";
import { loginAsExistingUser, loginAsNewUser, uniqueUser } from "../fixtures/auth";

test.describe("auth bypass", () => {
  test("new user lands on profile setup", async ({ page }) => {
    await loginAsNewUser(page, uniqueUser("new"));
    await expect(page.getByRole("button", { name: "Save Profile" })).toBeVisible();
  });

  test("existing user lands on My Events", async ({ page }) => {
    await loginAsExistingUser(page, uniqueUser("host"));
    await expect(page.getByRole("heading", { name: "My Events" })).toBeVisible();
    await expect(page.getByRole("button", { name: /Host an event/ })).toBeVisible();
  });
});
