// Tests for the signed-in avatar dropdown menu (My Account / Logout).
import { beforeEach, describe, expect, it, vi } from "vitest";
import { axe } from "@/test/axe";
import { buildUser } from "@/test/fixtures";
import { mockRouter, resetNextNavigationMock } from "@/test/nextNavigationMock";
import { renderWithProviders, screen, userEvent } from "@/test/render";
import UserAccountMenu from "./UserAccountMenu";

vi.mock("next/navigation", () => import("@/test/nextNavigationMock"));

describe("UserAccountMenu", () => {
  beforeEach(() => resetNextNavigationMock());

  it("renders nothing when signed out", () => {
    const { container } = renderWithProviders(<UserAccountMenu />, { isAuthenticated: false, user: null });
    expect(container).toBeEmptyDOMElement();
  });

  it("shows the user's display name when signed in", () => {
    const user = buildUser({ display_name: "Shallan Davar" });
    renderWithProviders(<UserAccountMenu />, { isAuthenticated: true, user });
    expect(screen.getByText("Shallan Davar")).toBeInTheDocument();
  });

  it("falls back to the Discord handle when there's no display name", () => {
    const user = buildUser({ display_name: "", discord_name: "ashketchum#0001" });
    renderWithProviders(<UserAccountMenu />, { isAuthenticated: true, user });
    expect(screen.getByText("@ashketchum#0001")).toBeInTheDocument();
  });

  it("opens the dropdown menu on click, with My Account and Logout items", async () => {
    const user = buildUser();
    const uEvent = userEvent.setup();
    renderWithProviders(<UserAccountMenu />, { isAuthenticated: true, user });

    await uEvent.click(screen.getByRole("button", { expanded: false }));

    expect(screen.getByRole("menuitem", { name: "My Account" })).toHaveAttribute("href", "/my_account");
    expect(screen.getByRole("menuitem", { name: "Logout" })).toBeInTheDocument();
  });

  it("closes the dropdown when clicking outside", async () => {
    const user = buildUser();
    const uEvent = userEvent.setup();
    renderWithProviders(
      <div>
        <UserAccountMenu />
        <button>Outside</button>
      </div>,
      { isAuthenticated: true, user },
    );
    await uEvent.click(screen.getByRole("button", { expanded: false }));
    expect(screen.getByRole("menu")).toBeInTheDocument();

    await uEvent.click(screen.getByRole("button", { name: "Outside" }));

    expect(screen.queryByRole("menu")).not.toBeInTheDocument();
  });

  it("logs out and redirects home when Logout is clicked", async () => {
    const user = buildUser();
    const logout = vi.fn().mockResolvedValue(undefined);
    const uEvent = userEvent.setup();
    renderWithProviders(<UserAccountMenu />, { isAuthenticated: true, user, logout });

    await uEvent.click(screen.getByRole("button", { expanded: false }));
    await uEvent.click(screen.getByRole("menuitem", { name: "Logout" }));

    expect(logout).toHaveBeenCalledTimes(1);
    expect(mockRouter.push).toHaveBeenCalledWith("/");
    expect(screen.queryByRole("menu")).not.toBeInTheDocument();
  });

  it("has no accessibility violations when closed", async () => {
    const { container } = renderWithProviders(<UserAccountMenu />, {
      isAuthenticated: true,
      user: buildUser({ display_name: "Shallan Davar" }),
    });
    expect(await axe(container)).toHaveNoViolations();
  });

  it("has no accessibility violations when the menu is open", async () => {
    const user = buildUser();
    const uEvent = userEvent.setup();
    const { container } = renderWithProviders(<UserAccountMenu />, { isAuthenticated: true, user });

    await uEvent.click(screen.getByRole("button", { expanded: false }));

    expect(await axe(container)).toHaveNoViolations();
  });
});
