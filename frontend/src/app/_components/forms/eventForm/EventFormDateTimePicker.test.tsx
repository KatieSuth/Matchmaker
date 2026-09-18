// Tests for the local-datetime picker wrapper around react-datepicker.
import { describe, expect, it, vi } from "vitest";
import { axe } from "@/test/axe";
import { render, screen, userEvent } from "@/test/render";
import { EventFormDateTimePicker } from "./EventFormDateTimePicker";

describe("EventFormDateTimePicker", () => {
  it("shows the placeholder text when there is no value", () => {
    render(<EventFormDateTimePicker value="" onChange={vi.fn()} placeholderText="Pick a time" />);
    expect(screen.getByPlaceholderText("Pick a time")).toBeInTheDocument();
  });

  it("displays the formatted value when one is set", () => {
    render(<EventFormDateTimePicker value="2026-09-20T14:30" onChange={vi.fn()} />);
    expect(screen.getByDisplayValue(/Sep 20, 2026/)).toBeInTheDocument();
  });

  it("is disabled when the disabled prop is set", () => {
    render(<EventFormDateTimePicker value="" onChange={vi.fn()} disabled />);
    expect(screen.getByRole("textbox")).toBeDisabled();
  });

  it("opens the calendar popup on click", async () => {
    const user = userEvent.setup();
    render(<EventFormDateTimePicker value="" onChange={vi.fn()} />);

    await user.click(screen.getByRole("textbox"));

    expect(document.querySelector(".react-datepicker")).toBeInTheDocument();
  });

  it("has no accessibility violations when labeled", async () => {
    const { container } = render(
      <>
        <label htmlFor="event-start">Start time</label>
        <EventFormDateTimePicker id="event-start" value="2026-09-20T14:30" onChange={vi.fn()} />
      </>,
    );
    expect(await axe(container)).toHaveNoViolations();
  });
});
