import { expect, test } from "@playwright/test";
import { loginAsNewUser, uniqueUser } from "../fixtures/auth";
import { pickSelectOption } from "../fixtures/ui";

test.describe("profile", () => {
  test("new user can save a profile and reach My Events", async ({ page }) => {
    await loginAsNewUser(page, uniqueUser("prof"));

    await page.getByPlaceholder("Optional public name").fill("E2E Player");
    await pickSelectOption(page, "— No preference —", "AMER");
    await page.getByRole("button", { name: "Add game" }).click();
    await pickSelectOption(page, "— Select a game —", "Valorant");
    await expect(page.getByText("— Select rank —").first()).toBeVisible();
    await page.getByPlaceholder("YourTag#1234").fill("E2E#0001");
    await pickSelectOption(page, "— Select rank —", "Iron 1");
    await pickSelectOption(page, "— Select rank —", "Iron 1");

    await page.getByRole("button", { name: "Save Profile" }).click();
    await expect(page.getByRole("heading", { name: "My Events" })).toBeVisible();
  });
});
