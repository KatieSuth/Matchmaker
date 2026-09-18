// Tests for the authenticated app shell's top nav (logo, nav links, avatar menu).
import { describe, expect, it, vi } from "vitest";
import { NAV_LINKS } from "@/app/_lib/constants";
import { axe } from "@/test/axe";
import { buildUser } from "@/test/fixtures";
import { setMockPathname, resetNextNavigationMock } from "@/test/nextNavigationMock";
import { renderWithProviders, screen } from "@/test/render";
import AppNav from "./AppNav";

vi.mock("next/navigation", () => import("@/test/nextNavigationMock"));

describe("AppNav", () => {
  it("renders a link for every configured nav item", () => {
    resetNextNavigationMock();
    renderWithProviders(<AppNav />);
    for (const { label, href } of NAV_LINKS) {
      expect(screen.getByRole("link", { name: label })).toHaveAttribute("href", href);
    }
  });

  it("marks the nav link matching the current path as active", () => {
    setMockPathname("/my_events");
    renderWithProviders(<AppNav />);
    expect(screen.getByRole("link", { name: "Events" })).toHaveClass("nav-link-active");
  });

  it("does not mark links inactive for other paths", () => {
    setMockPathname("/about");
    renderWithProviders(<AppNav />);
    expect(screen.getByRole("link", { name: "Events" })).not.toHaveClass("nav-link-active");
  });

  it("links the logo to the events dashboard", () => {
    resetNextNavigationMock();
    renderWithProviders(<AppNav />);
    expect(screen.getByRole("link", { name: "Matchmaker" })).toHaveAttribute("href", "/my_events");
  });

  it("shows the account menu when signed in", () => {
    resetNextNavigationMock();
    renderWithProviders(<AppNav />, { isAuthenticated: true, user: buildUser({ display_name: "Ash" }) });
    expect(screen.getByText("Ash")).toBeInTheDocument();
  });

  it("has no accessibility violations", async () => {
    resetNextNavigationMock();
    const { container } = renderWithProviders(<AppNav />, {
      isAuthenticated: true,
      user: buildUser({ display_name: "Ash" }),
    });
    expect(await axe(container)).toHaveNoViolations();
  });
});
