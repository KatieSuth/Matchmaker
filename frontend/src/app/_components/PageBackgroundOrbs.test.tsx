// Tests for the decorative background-orb glow effect (login + app shell background).
import { describe, expect, it } from "vitest";
import { axe } from "@/test/axe";
import { render, screen } from "@/test/render";
import PageBackgroundOrbs from "./PageBackgroundOrbs";

describe("PageBackgroundOrbs", () => {
  it("renders two decorative orbs hidden from assistive technology", () => {
    render(<PageBackgroundOrbs />);
    // aria-hidden elements are excluded from the accessibility tree, so query the DOM directly.
    expect(document.querySelectorAll('[aria-hidden="true"]')).toHaveLength(2);
    expect(screen.queryByRole("img")).not.toBeInTheDocument();
  });

  it("has no accessibility violations (purely decorative)", async () => {
    const { container } = render(<PageBackgroundOrbs />);
    expect(await axe(container)).toHaveNoViolations();
  });
});
