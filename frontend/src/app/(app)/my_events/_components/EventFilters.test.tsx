// Tests for the My Events filters card: time toggle, apply/clear gating, and active-filter summary.
import { describe, expect, it, vi } from "vitest";
import { SelectOption } from "@/app/_components/Select";
import { axe } from "@/test/axe";
import { render, screen, userEvent } from "@/test/render";
import { EventFilters } from "./EventFilters";

const gameOption: SelectOption = { value: "game-1", label: "Valorant" };

function baseProps(overrides: Partial<Parameters<typeof EventFilters>[0]> = {}) {
  return {
    timeFilter: "upcoming" as const,
    setTimeFilter: vi.fn(),
    dateRangeActive: false,
    pendingGame: null,
    setPendingGame: vi.fn(),
    gameSelectOptions: [{ label: "", options: [gameOption] }],
    gamesSelectLoading: false,
    gamesError: false,
    pendingFrom: null,
    setPendingFrom: vi.fn(),
    pendingTo: null,
    setPendingTo: vi.fn(),
    appliedGame: null,
    appliedFrom: null,
    appliedTo: null,
    onApply: vi.fn(),
    onClear: vi.fn(),
    ...overrides,
  };
}

describe("EventFilters", () => {
  it("switches the time filter when a pill is clicked", async () => {
    const setTimeFilter = vi.fn();
    const user = userEvent.setup();
    render(<EventFilters {...baseProps({ setTimeFilter })} />);

    await user.click(screen.getByRole("button", { name: "Past" }));

    expect(setTimeFilter).toHaveBeenCalledWith("past");
  });

  it("disables the time toggle while a date range filter is active", async () => {
    const setTimeFilter = vi.fn();
    const user = userEvent.setup();
    render(<EventFilters {...baseProps({ setTimeFilter, dateRangeActive: true })} />);

    await user.click(screen.getByRole("button", { name: "Past" }));

    expect(setTimeFilter).not.toHaveBeenCalled();
  });

  it("disables Apply when the pending filters match the applied ones", () => {
    render(<EventFilters {...baseProps({ appliedGame: gameOption, pendingGame: gameOption })} />);
    expect(screen.getByRole("button", { name: "Apply" })).toBeDisabled();
  });

  it("enables Apply once the pending game differs from the applied one", () => {
    render(<EventFilters {...baseProps({ appliedGame: null, pendingGame: gameOption })} />);
    expect(screen.getByRole("button", { name: "Apply" })).toBeEnabled();
  });

  it("disables Apply and shows a validation message when 'From' is after 'To'", () => {
    const from = new Date("2026-09-20T00:00:00Z");
    const to = new Date("2026-09-10T00:00:00Z");
    render(<EventFilters {...baseProps({ pendingFrom: from, pendingTo: to })} />);

    expect(screen.getByRole("button", { name: "Apply" })).toBeDisabled();
    expect(screen.getByText(/"From" date must be on or before "To" date/)).toBeInTheDocument();
  });

  it("calls onApply when Apply is clicked while enabled", async () => {
    const onApply = vi.fn();
    const user = userEvent.setup();
    render(<EventFilters {...baseProps({ onApply, pendingGame: gameOption })} />);

    await user.click(screen.getByRole("button", { name: "Apply" }));

    expect(onApply).toHaveBeenCalledTimes(1);
  });

  it("shows no active-filter summary when nothing is applied", () => {
    render(<EventFilters {...baseProps()} />);
    expect(screen.queryByText("Showing:")).not.toBeInTheDocument();
  });

  it("shows the applied game as an active-filter chip with a Clear action", async () => {
    const onClear = vi.fn();
    const user = userEvent.setup();
    render(<EventFilters {...baseProps({ appliedGame: gameOption, onClear })} />);

    expect(screen.getByText("Showing:")).toBeInTheDocument();
    expect(screen.getByText("Valorant")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Clear" }));

    expect(onClear).toHaveBeenCalledTimes(1);
  });

  it("shows a combined date-range chip when both From and To are applied", () => {
    const appliedFrom = new Date("2026-09-01T00:00:00Z");
    const appliedTo = new Date("2026-09-30T00:00:00Z");
    render(<EventFilters {...baseProps({ appliedFrom, appliedTo })} />);

    expect(screen.getByText(/–/)).toBeInTheDocument();
  });

  it("shows a 'From ...' chip when only a start date is applied", () => {
    const appliedFrom = new Date("2026-09-01T00:00:00Z");
    render(<EventFilters {...baseProps({ appliedFrom })} />);

    expect(screen.getByText(/^From /)).toBeInTheDocument();
  });

  it("has no accessibility violations", async () => {
    const { container } = render(
      <EventFilters {...baseProps({ appliedGame: gameOption, pendingGame: gameOption })} />,
    );
    expect(await axe(container)).toHaveNoViolations();
  });
});
