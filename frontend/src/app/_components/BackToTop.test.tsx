// Tests for the "Back to top" anchor link between About / Privacy / FAQ cards.
import { describe, expect, it } from "vitest";
import { axe } from "@/test/axe";
import { render, screen } from "@/test/render";
import BackToTop from "./BackToTop";

describe("BackToTop", () => {
  it("links to the #top anchor", () => {
    render(<BackToTop />);
    expect(screen.getByRole("link", { name: "Back to top" })).toHaveAttribute("href", "#top");
  });

  it("applies a custom className alongside the default layout classes", () => {
    const { container } = render(<BackToTop className="mt-8" />);
    expect(container.firstChild).toHaveClass("mt-8", "flex", "justify-center");
  });

  it("has no accessibility violations", async () => {
    const { container } = render(<BackToTop />);
    expect(await axe(container)).toHaveNoViolations();
  });
});
