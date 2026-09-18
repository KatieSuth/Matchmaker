// Tests for the host-only "swap player" sheet.
import { describe, expect, it, vi } from "vitest";
import { axe } from "@/test/axe";
import { render, screen, userEvent } from "@/test/render";
import { PlayerPlacement } from "../../_types";
import { SwapPlayerSheet } from "./SwapPlayerSheet";

const placement: PlayerPlacement = {
  eventId: "event-1",
  userId: "user-1",
  discordName: "Alice#0001",
  lobbyId: "lobby-1",
  sourceLobbyIndex: 0,
  teamNumber: 1,
};

function baseProps(overrides: Partial<Parameters<typeof SwapPlayerSheet>[0]> = {}) {
  return {
    isOpen: true,
    onClose: vi.fn(),
    pendingSwap: placement,
    swapTargetUserId: "",
    onChangeSwapTarget: vi.fn(),
    swapCandidateOptions: [{ value: "user-2", label: "Bob#0001" }],
    swapError: null,
    working: false,
    onSubmit: vi.fn(),
    ...overrides,
  };
}

describe("SwapPlayerSheet", () => {
  it("titles the sheet with the target player's name", () => {
    render(<SwapPlayerSheet {...baseProps()} />);
    expect(screen.getByRole("dialog", { name: "Swap Alice#0001" })).toBeInTheDocument();
  });

  it("shows a swap error message when present", () => {
    render(<SwapPlayerSheet {...baseProps({ swapError: "That player cannot be swapped." })} />);
    expect(screen.getByText("That player cannot be swapped.")).toBeInTheDocument();
  });

  it("disables Submit until a swap target is chosen", () => {
    const { rerender } = render(<SwapPlayerSheet {...baseProps({ swapTargetUserId: "" })} />);
    expect(screen.getByRole("button", { name: "Submit" })).toBeDisabled();

    rerender(<SwapPlayerSheet {...baseProps({ swapTargetUserId: "user-2" })} />);
    expect(screen.getByRole("button", { name: "Submit" })).toBeEnabled();
  });

  it("calls onSubmit and onClose from their respective buttons", async () => {
    const onSubmit = vi.fn();
    const onClose = vi.fn();
    const user = userEvent.setup();
    render(<SwapPlayerSheet {...baseProps({ swapTargetUserId: "user-2", onSubmit, onClose })} />);

    await user.click(screen.getByRole("button", { name: "Submit" }));
    expect(onSubmit).toHaveBeenCalledTimes(1);

    await user.click(screen.getByRole("button", { name: "Cancel" }));
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("has no accessibility violations", async () => {
    render(<SwapPlayerSheet {...baseProps({ swapTargetUserId: "user-2" })} />);
    // The sheet portals to document.body, so axe needs to scan the whole document.
    expect(await axe(document.body)).toHaveNoViolations();
  });
});
