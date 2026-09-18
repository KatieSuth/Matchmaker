// Tests for the EventCard-shaped loading placeholder.
import { describe, expect, it } from "vitest";
import { axe } from "@/test/axe";
import { render } from "@/test/render";
import { SkeletonCard } from "./SkeletonCard";

describe("SkeletonCard", () => {
  it("renders a pulsing placeholder card", () => {
    render(<SkeletonCard />);
    expect(document.querySelector(".animate-pulse")).toBeInTheDocument();
  });

  it("has no accessibility violations", async () => {
    const { container } = render(<SkeletonCard />);
    expect(await axe(container)).toHaveNoViolations();
  });
});
