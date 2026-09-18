// Tests for the "make lobby host" confirmation sheet and its volunteer list.
import { describe, expect, it, vi } from "vitest";
import { axe } from "@/test/axe";
import { render, screen, userEvent } from "@/test/render";
import { PendingLobbyHostChange } from "../../_types";
import { LobbyHostConfirmSheet } from "./LobbyHostConfirmSheet";

const pendingLobbyHostChange: PendingLobbyHostChange = {
  placement: {
    eventId: "event-1",
    userId: "user-1",
    discordName: "Alice#0001",
    lobbyId: "lobby-1",
    sourceLobbyIndex: 0,
    teamNumber: 1,
  },
  volunteerOptions: [],
};

function baseProps(overrides: Partial<Parameters<typeof LobbyHostConfirmSheet>[0]> = {}) {
  return {
    isOpen: true,
    onClose: vi.fn(),
    pendingLobbyHostChange,
    working: false,
    onConfirm: vi.fn(),
    ...overrides,
  };
}

describe("LobbyHostConfirmSheet", () => {
  it("shows a fallback message when there is nothing pending", () => {
    render(<LobbyHostConfirmSheet {...baseProps({ pendingLobbyHostChange: null })} />);
    expect(screen.getByText("No player selected.")).toBeInTheDocument();
  });

  it("titles the sheet with the target player's name", () => {
    render(<LobbyHostConfirmSheet {...baseProps()} />);
    expect(screen.getByRole("dialog", { name: "Make Alice#0001 Lobby Host" })).toBeInTheDocument();
  });

  it("shows a no-volunteers message when there are none", () => {
    render(<LobbyHostConfirmSheet {...baseProps()} />);
    expect(screen.getByText(/no other players on a team in this lobby who want to host/)).toBeInTheDocument();
  });

  it("lists volunteer options with team number and current-host tag", () => {
    render(
      <LobbyHostConfirmSheet
        {...baseProps({
          pendingLobbyHostChange: {
            ...pendingLobbyHostChange,
            volunteerOptions: [
              { userId: "user-2", discordName: "Bob#0001", teamNumber: 2, isCurrentHost: false },
              { userId: "user-3", discordName: "Cara#0001", teamNumber: 1, isCurrentHost: true },
            ],
          },
        })}
      />,
    );
    expect(screen.getByText(/Bob#0001 · Team 2/)).toBeInTheDocument();
    expect(screen.getByText(/Cara#0001 · Team 1 · \(current host\)/)).toBeInTheDocument();
  });

  it("calls onConfirm / onClose from their respective buttons", async () => {
    const onConfirm = vi.fn();
    const onClose = vi.fn();
    const user = userEvent.setup();
    render(<LobbyHostConfirmSheet {...baseProps({ onConfirm, onClose })} />);

    await user.click(screen.getByRole("button", { name: "Yes, make lobby host" }));
    expect(onConfirm).toHaveBeenCalledTimes(1);

    await user.click(screen.getByRole("button", { name: "No" }));
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("has no accessibility violations", async () => {
    render(<LobbyHostConfirmSheet {...baseProps()} />);
    // The sheet portals to document.body, so axe needs to scan the whole document.
    expect(await axe(document.body)).toHaveNoViolations();
  });
});
