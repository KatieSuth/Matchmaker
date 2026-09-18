// Tests for the labeled on/off row (label + optional description + ToggleSwitch).
import { describe, expect, it, vi } from "vitest";
import { axe } from "@/test/axe";
import { render, screen, userEvent } from "@/test/render";
import { ToggleRow } from "./ToggleRow";

describe("ToggleRow", () => {
  it("renders the label and optional description", () => {
    render(<ToggleRow label="Show pronouns" description="Visible on your profile" checked={false} onChange={vi.fn()} />);
    expect(screen.getByText("Show pronouns")).toBeInTheDocument();
    expect(screen.getByText("Visible on your profile")).toBeInTheDocument();
  });

  it("omits the description paragraph when none is given", () => {
    render(<ToggleRow label="Show pronouns" checked={false} onChange={vi.fn()} />);
    expect(screen.queryByText("Visible on your profile")).not.toBeInTheDocument();
  });

  it("renders a labelAccessory inline with the label", () => {
    render(<ToggleRow label="Show pronouns" labelAccessory={<span>info</span>} checked={false} onChange={vi.fn()} />);
    expect(screen.getByText("info")).toBeInTheDocument();
  });

  it("toggles via the embedded switch", async () => {
    const onChange = vi.fn();
    const user = userEvent.setup();
    render(<ToggleRow label="Show pronouns" checked={false} onChange={onChange} />);

    await user.click(screen.getByRole("switch"));

    expect(onChange).toHaveBeenCalledWith(true);
  });

  it("has no accessibility violations", async () => {
    const { container } = render(<ToggleRow label="Show pronouns" description="Visible on your profile" checked={true} onChange={vi.fn()} />);
    expect(await axe(container)).toHaveNoViolations();
  });
});
