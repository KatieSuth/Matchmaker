// Tests for the pill-shaped switch control.
import { describe, expect, it, vi } from "vitest";
import { axe } from "@/test/axe";
import { render, screen, userEvent } from "@/test/render";
import { ToggleSwitch } from "./ToggleSwitch";

describe("ToggleSwitch", () => {
  it("renders as a switch with the correct aria-checked state", () => {
    render(<ToggleSwitch checked={true} onChange={vi.fn()} />);
    expect(screen.getByRole("switch")).toBeChecked();
  });

  it("calls onChange with the toggled value on click", async () => {
    const onChange = vi.fn();
    const user = userEvent.setup();
    render(<ToggleSwitch checked={false} onChange={onChange} />);

    await user.click(screen.getByRole("switch"));

    expect(onChange).toHaveBeenCalledWith(true);
  });

  it("does not call onChange when disabled", async () => {
    const onChange = vi.fn();
    const user = userEvent.setup();
    render(<ToggleSwitch checked={false} onChange={onChange} disabled />);

    await user.click(screen.getByRole("switch"));

    expect(onChange).not.toHaveBeenCalled();
    expect(screen.getByRole("switch")).toBeDisabled();
  });

  it("exposes an accessible name via ariaLabel", () => {
    render(<ToggleSwitch checked={true} onChange={vi.fn()} ariaLabel="Notifications" />);
    expect(screen.getByRole("switch", { name: "Notifications" })).toBeInTheDocument();
  });

  it("has no accessibility violations", async () => {
    const { container } = render(<ToggleSwitch checked={true} onChange={vi.fn()} id="notify" ariaLabel="Notifications" />);
    expect(await axe(container)).toHaveNoViolations();
  });
});
