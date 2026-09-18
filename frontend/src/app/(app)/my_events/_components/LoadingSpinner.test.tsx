// Tests for the small inline "Load more" spinner.
import { describe, expect, it } from "vitest";
import { axe } from "@/test/axe";
import { render } from "@/test/render";
import { LoadingSpinner } from "./LoadingSpinner";

describe("LoadingSpinner", () => {
  it("renders an animated svg", () => {
    render(<LoadingSpinner />);
    // Purely decorative SVG with no accessible role, so query the DOM directly.
    expect(document.querySelector("svg")).toHaveClass("animate-spin");
  });

  it("has no accessibility violations", async () => {
    const { container } = render(<LoadingSpinner />);
    expect(await axe(container)).toHaveNoViolations();
  });
});
