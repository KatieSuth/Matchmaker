import { expect, test } from "@playwright/test";

test.describe("public pages", () => {
  test("landing shows Discord login", async ({ page }) => {
    await page.goto("/");
    await expect(page.getByText(/Welcome to Matchmaker/)).toBeVisible();
    await expect(page.getByRole("link", { name: "Log in with Discord" })).toBeVisible();
  });

  test("about page is reachable logged out", async ({ page }) => {
    await page.goto("/");
    await page.getByRole("link", { name: "About & Privacy" }).click();
    await expect(page).toHaveURL(/\/about$/);
    await expect(page.getByRole("heading", { name: "About Matchmaker" })).toBeVisible();
  });
});
