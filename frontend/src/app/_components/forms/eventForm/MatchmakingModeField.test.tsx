// Tests for the "balanced vs. rank grouping" matchmaking mode radio-card picker.
import { describe, expect, it, vi } from "vitest";
import { axe } from "@/test/axe";
import { render, screen, userEvent } from "@/test/render";
import { MatchmakingModeField } from "./MatchmakingModeField";

describe("MatchmakingModeField", () => {
  it("renders a radiogroup with Balanced and Rank Grouping options", () => {
    render(<MatchmakingModeField value="balanced" onChange={vi.fn()} />);
    expect(screen.getByRole("radiogroup")).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: /Balanced/ })).toBeChecked();
    expect(screen.getByRole("radio", { name: /Rank Grouping/ })).not.toBeChecked();
  });

  it("calls onChange with the newly selected option", async () => {
    const onChange = vi.fn();
    const user = userEvent.setup();
    render(<MatchmakingModeField value="balanced" onChange={onChange} />);

    await user.click(screen.getByRole("radio", { name: /Rank Grouping/ }));

    expect(onChange).toHaveBeenCalledWith("ranked");
  });

  it("disables both options and does not fire onChange when disabled", async () => {
    const onChange = vi.fn();
    const user = userEvent.setup();
    render(<MatchmakingModeField value="balanced" onChange={onChange} disabled />);

    expect(screen.getByRole("radio", { name: /Rank Grouping/ })).toBeDisabled();
    await user.click(screen.getByRole("radio", { name: /Rank Grouping/ }));

    expect(onChange).not.toHaveBeenCalled();
  });

  it("has no accessibility violations", async () => {
    const { container } = render(
      <div>
        <span id="matchmaking-mode-label">Matchmaking Mode</span>
        <MatchmakingModeField value="balanced" onChange={vi.fn()} />
      </div>,
    );
    expect(await axe(container)).toHaveNoViolations();
  });
});
