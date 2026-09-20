// Behavior + a11y coverage for the shared +/- numeric input used across event scheduling forms.
import { describe, expect, it, vi } from "vitest";
import { axe } from "@/test/axe";
import { fireEvent, render, screen, userEvent } from "@/test/render";
import { NumberStepper } from "./NumberStepper";

describe("NumberStepper", () => {
  it("calls onChange with value + 1 when the increment button is clicked", async () => {
    const onChange = vi.fn();
    const user = userEvent.setup();
    render(<NumberStepper label="Sub minimum" value={2} min={0} onChange={onChange} />);

    await user.click(screen.getByRole("button", { name: "Increase Sub minimum" }));

    expect(onChange).toHaveBeenCalledWith(3);
  });

  it("calls onChange with value - 1 when the decrement button is clicked, clamped to min", async () => {
    const onChange = vi.fn();
    const user = userEvent.setup();
    render(<NumberStepper label="Sub minimum" value={0} min={0} onChange={onChange} />);

    const decrementButton = screen.getByRole("button", { name: "Decrease Sub minimum" });
    expect(decrementButton).toBeDisabled();

    await user.click(decrementButton);
    expect(onChange).not.toHaveBeenCalled();
  });

  it("clamps a typed value of \"0\" up to min (the handler's `Number(value) || min` fallback)", () => {
    const onChange = vi.fn();
    // NumberStepper is a fully controlled input (`value` prop never changes here since `onChange`
    // is a mock), so simulating multi-keystroke typing via userEvent isn't meaningful — fire a
    // single change event directly to test the handler's value -> onChange(next) contract.
    render(<NumberStepper label="Sub minimum" value={2} min={1} onChange={onChange} />);
    const input = screen.getByRole("spinbutton");

    fireEvent.change(input, { target: { value: "0" } });

    expect(onChange).toHaveBeenCalledWith(1);
  });

  it("clamps a typed value below min up to min", () => {
    const onChange = vi.fn();
    render(<NumberStepper label="Sub minimum" value={5} min={3} onChange={onChange} />);
    const input = screen.getByRole("spinbutton");

    fireEvent.change(input, { target: { value: "1" } });

    expect(onChange).toHaveBeenCalledWith(3);
  });

  it("disables both buttons and the input when disabled", () => {
    render(<NumberStepper label="Sub minimum" value={2} min={0} onChange={vi.fn()} disabled />);

    expect(screen.getByRole("button", { name: "Increase Sub minimum" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Decrease Sub minimum" })).toBeDisabled();
    expect(screen.getByRole("spinbutton")).toBeDisabled();
  });

  it("renders the optional hint text when provided", () => {
    render(<NumberStepper label="Sub minimum" value={2} min={0} onChange={vi.fn()} hint="At least 1 required" />);
    expect(screen.getByText("At least 1 required")).toBeInTheDocument();
  });

  it("has no detectable accessibility violations", async () => {
    const { container } = render(<NumberStepper label="Sub minimum" value={2} min={0} onChange={vi.fn()} />);
    expect(await axe(container)).toHaveNoViolations();
  });

  it("applies hide-number-spinners so native type=number arrows stay suppressed", () => {
    render(<NumberStepper label="Sub minimum" value={2} min={0} onChange={vi.fn()} />);
    expect(screen.getByRole("spinbutton")).toHaveClass("hide-number-spinners");
  });
});
