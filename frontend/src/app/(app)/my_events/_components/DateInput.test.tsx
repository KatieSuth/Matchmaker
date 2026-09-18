// Tests for the custom react-datepicker trigger button.
import { createRef } from "react";
import { describe, expect, it, vi } from "vitest";
import { axe } from "@/test/axe";
import { render, screen, userEvent } from "@/test/render";
import { DateInput } from "./DateInput";

describe("DateInput", () => {
  it("shows the placeholder when there is no value", () => {
    render(<DateInput placeholder="Select a date" />);
    expect(screen.getByRole("button", { name: "Select a date" })).toBeInTheDocument();
  });

  it("shows the value instead of the placeholder when given", () => {
    render(<DateInput placeholder="Select a date" value="Sep 14, 2026" />);
    expect(screen.getByRole("button", { name: "Sep 14, 2026" })).toBeInTheDocument();
  });

  it("calls onClick when the trigger button is clicked", async () => {
    const onClick = vi.fn();
    const user = userEvent.setup();
    render(<DateInput placeholder="Select a date" onClick={onClick} />);

    await user.click(screen.getByRole("button", { name: "Select a date" }));

    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it("does not show a clear button without a value, even if isClearable", () => {
    render(<DateInput placeholder="Select a date" isClearable />);
    expect(screen.getAllByRole("button")).toHaveLength(1);
  });

  it("shows a clear button when isClearable and a value is set, and calls onClear without triggering onClick", async () => {
    const onClick = vi.fn();
    const onClear = vi.fn();
    const user = userEvent.setup();
    render(<DateInput placeholder="Select a date" value="Sep 14, 2026" isClearable onClick={onClick} onClear={onClear} />);

    const buttons = screen.getAllByRole("button");
    expect(buttons).toHaveLength(2);
    await user.click(buttons[1]);

    expect(onClear).toHaveBeenCalledTimes(1);
    expect(onClick).not.toHaveBeenCalled();
  });

  it("forwards a ref to the trigger button for react-datepicker", () => {
    const ref = createRef<HTMLButtonElement>();
    render(<DateInput ref={ref} placeholder="Select a date" />);
    expect(ref.current).toBeInstanceOf(HTMLButtonElement);
    expect(ref.current).toHaveAttribute("aria-label", "Select a date");
  });

  it("has no accessibility violations", async () => {
    const { container } = render(
      <DateInput placeholder="Select a date" value="Sep 14, 2026" isClearable onClear={vi.fn()} />,
    );
    expect(await axe(container)).toHaveNoViolations();
  });
});
