// Tests for the host-only "move to subs" lobby-picker sheet.
import { describe, expect, it, vi } from "vitest";
import { axe } from "@/test/axe";
import { render, screen, userEvent } from "@/test/render";
import { PlayerPlacement } from "../../_types";
import { MoveToSubsSheet } from "./MoveToSubsSheet";

const placement: PlayerPlacement = {
  eventId: "event-1",
  userId: "user-1",
  discordName: "Alice#0001",
  lobbyId: null,
  sourceLobbyIndex: null,
  teamNumber: undefined,
};

function baseProps(overrides: Partial<Parameters<typeof MoveToSubsSheet>[0]> = {}) {
  return {
    isOpen: true,
    onClose: vi.fn(),
    pendingMoveToSubs: placement,
    moveToSubsLobbyId: "",
    onChangeLobby: vi.fn(),
    moveToSubsLobbyOptions: [{ value: "lobby-1", label: "Lobby 1" }],
    moveToSubsError: null,
    working: false,
    onSubmit: vi.fn(),
    ...overrides,
  };
}

describe("MoveToSubsSheet", () => {
  it("titles the sheet with the target player's name", () => {
    render(<MoveToSubsSheet {...baseProps()} />);
    expect(screen.getByRole("dialog", { name: "Move Alice#0001 to subs" })).toBeInTheDocument();
  });

  it("shows a lobby-selection error when present", () => {
    render(<MoveToSubsSheet {...baseProps({ moveToSubsError: "No room in any sub pool." })} />);
    expect(screen.getByText("No room in any sub pool.")).toBeInTheDocument();
  });

  it("disables Submit until a lobby is chosen", () => {
    const { rerender } = render(<MoveToSubsSheet {...baseProps({ moveToSubsLobbyId: "" })} />);
    expect(screen.getByRole("button", { name: "Submit" })).toBeDisabled();

    rerender(<MoveToSubsSheet {...baseProps({ moveToSubsLobbyId: "lobby-1" })} />);
    expect(screen.getByRole("button", { name: "Submit" })).toBeEnabled();
  });

  it("calls onSubmit / onClose from their buttons", async () => {
    const onSubmit = vi.fn();
    const onClose = vi.fn();
    const user = userEvent.setup();
    render(<MoveToSubsSheet {...baseProps({ moveToSubsLobbyId: "lobby-1", onSubmit, onClose })} />);

    await user.click(screen.getByRole("button", { name: "Submit" }));
    expect(onSubmit).toHaveBeenCalledTimes(1);

    await user.click(screen.getByRole("button", { name: "Cancel" }));
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("has no accessibility violations", async () => {
    render(<MoveToSubsSheet {...baseProps({ moveToSubsLobbyId: "lobby-1" })} />);
    // The sheet portals to document.body, so axe needs to scan the whole document.
    expect(await axe(document.body)).toHaveNoViolations();
  });
});
