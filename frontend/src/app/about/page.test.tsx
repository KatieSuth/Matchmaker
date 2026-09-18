import { describe, expect, it, vi } from "vitest";
import { renderWithProviders, screen } from "@/test/render";
import { axe } from "@/test/axe";
import AboutPage from "./page";

vi.mock("next/navigation", () => import("@/test/nextNavigationMock"));

describe("About page", () => {
  it("renders the about heading and in-page nav", () => {
    renderWithProviders(<AboutPage />);
    expect(screen.getByRole("link", { name: "Matchmaker" })).toHaveAttribute("href", "/");
    expect(screen.getByRole("heading", { name: "About Matchmaker" })).toBeInTheDocument();
  });

  it("has no accessibility violations", async () => {
    const { container } = renderWithProviders(<AboutPage />);
    expect(await axe(container)).toHaveNoViolations();
  });
});
