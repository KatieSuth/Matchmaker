// Tests for the small titled section divider used in long forms.
import { describe, expect, it } from "vitest";
import { axe } from "@/test/axe";
import { render, screen } from "@/test/render";
import { SectionDivider } from "./SectionDivider";

describe("SectionDivider", () => {
  it("renders the given title as a heading", () => {
    render(<SectionDivider title="Account Details" />);
    expect(screen.getByRole("heading", { name: "Account Details" })).toBeInTheDocument();
  });

  it("has no obvious accessibility violations", async () => {
    const { container } = render(<SectionDivider title="Account Details" />);
    expect(await axe(container)).toHaveNoViolations();
  });
});
