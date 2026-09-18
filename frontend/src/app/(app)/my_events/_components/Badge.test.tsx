// Tests for the open/closed registration-status pill.
import { describe, expect, it } from "vitest";
import { axe } from "@/test/axe";
import { render, screen } from "@/test/render";
import { Badge } from "./Badge";

describe("Badge", () => {
  it("renders 'Open' when open", () => {
    render(<Badge open={true} />);
    expect(screen.getByText("Open")).toBeInTheDocument();
  });

  it("renders 'Closed' when not open", () => {
    render(<Badge open={false} />);
    expect(screen.getByText("Closed")).toBeInTheDocument();
  });

  it("has no accessibility violations", async () => {
    const { container } = render(<Badge open={true} />);
    expect(await axe(container)).toHaveNoViolations();
  });
});
