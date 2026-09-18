// Tests for the post-lock-in teams view: per-lobby rosters, substitutes, and host-only unplaced.
import { describe, expect, it, vi } from "vitest";
import { axe } from "@/test/axe";
import {
  buildEventGroupEvent,
  buildEventLobby,
  buildEventRegistration,
  buildEventTeam,
  buildGameRank,
  buildLobbyPlayer,
} from "@/test/fixtures";
import { render, screen, userEvent } from "@/test/render";
import { TeamsPanel } from "./TeamsPanel";

function baseProps(overrides: Partial<Parameters<typeof TeamsPanel>[0]> = {}) {
  return {
    event: buildEventGroupEvent(),
    gameNumber: 1,
    eventRegion: "AMER",
    isHostView: false,
    gameRanks: [],
    onShowDetails: vi.fn(),
    onDeleteRegistrationForGame: vi.fn(),
    onDeleteAllFromUser: vi.fn(),
    onJoinLobby: vi.fn(),
    showJoinLobby: false,
    ...overrides,
  };
}

describe("TeamsPanel", () => {
  it("renders nothing when the game has no lobbies yet", () => {
    const event = buildEventGroupEvent({ lobbies: [] });
    const { container } = render(<TeamsPanel {...baseProps({ event })} />);
    expect(container).toBeEmptyDOMElement();
  });

  it("renders a lobby heading and its teams with players", () => {
    const player = buildLobbyPlayer({ display_name: "Alice" });
    const lobby = buildEventLobby({ teams: [buildEventTeam({ team_number: 1, players: [player] })] });
    const event = buildEventGroupEvent({ lobbies: [lobby] });
    render(<TeamsPanel {...baseProps({ event })} />);

    expect(screen.getByText(/Lobby 1/)).toBeInTheDocument();
    expect(screen.getByText("Team 1")).toBeInTheDocument();
    expect(screen.getByText(/Alice/)).toBeInTheDocument();
  });

  it("shows the lobby host name when set", () => {
    const host = buildLobbyPlayer({ user_id: "host-1", display_name: "Hosty", discord_name: "hosty#0001" });
    const lobby = buildEventLobby({ host_id: "host-1", teams: [buildEventTeam({ players: [host] })] });
    const event = buildEventGroupEvent({ lobbies: [lobby] });
    render(<TeamsPanel {...baseProps({ event })} />);

    expect(screen.getByText(/Host: Hosty @hosty#0001/)).toBeInTheDocument();
  });

  it("shows the fairness warning banner and icon for an unfair lobby", () => {
    const lobby = buildEventLobby({ fairness_warning: true, fairness_warning_at_lock: true });
    const event = buildEventGroupEvent({ lobbies: [lobby] });
    render(<TeamsPanel {...baseProps({ event })} />);

    expect(screen.getByLabelText("Unfair lobby")).toBeInTheDocument();
    expect(screen.getByText(/rank spread was too wide/)).toBeInTheDocument();
  });

  it("shows the Join Lobby link only when showJoinLobby is true, and wires it up", async () => {
    const onJoinLobby = vi.fn();
    const lobby = buildEventLobby({ id: "lobby-1" });
    const event = buildEventGroupEvent({ id: "event-1", start_time: "2026-09-20T18:00:00Z", lobbies: [lobby] });
    const { rerender } = render(<TeamsPanel {...baseProps({ event, showJoinLobby: false })} />);
    expect(screen.queryByText("Join Lobby")).not.toBeInTheDocument();

    rerender(<TeamsPanel {...baseProps({ event, showJoinLobby: true, onJoinLobby })} />);
    const user = userEvent.setup();
    await user.click(screen.getByText("Join Lobby"));

    expect(onJoinLobby).toHaveBeenCalledWith(lobby, 0, 1, "2026-09-20T18:00:00Z");
  });

  it("shows the average rank per team only for the host view", () => {
    const player = buildLobbyPlayer({ avg_rank_order: 1 });
    const lobby = buildEventLobby({ teams: [buildEventTeam({ players: [player] })] });
    const event = buildEventGroupEvent({ lobbies: [lobby] });
    const gameRanks = [buildGameRank({ order: 1, name: "Gold" })];
    const { rerender } = render(<TeamsPanel {...baseProps({ event, isHostView: false, gameRanks })} />);
    expect(screen.queryByText(/Average:/)).not.toBeInTheDocument();

    rerender(<TeamsPanel {...baseProps({ event, isHostView: true, gameRanks })} />);
    expect(screen.getByText(/Average:/)).toBeInTheDocument();
  });

  it("shows a no-substitutes message when the game has none", () => {
    const lobby = buildEventLobby({ subs: [] });
    const event = buildEventGroupEvent({ lobbies: [lobby] });
    render(<TeamsPanel {...baseProps({ event })} />);
    expect(screen.getByText("Substitutes")).toBeInTheDocument();
  });

  it("lists substitutes with a player count", () => {
    const sub = buildLobbyPlayer({ display_name: "Subby" });
    const lobby = buildEventLobby({ subs: [sub] });
    const event = buildEventGroupEvent({ lobbies: [lobby] });
    render(<TeamsPanel {...baseProps({ event })} />);

    expect(screen.getByText(/Substitutes · 1/)).toBeInTheDocument();
    expect(screen.getByText(/Subby/)).toBeInTheDocument();
  });

  it("shows unplaced players only for the host view", () => {
    const unplaced = buildEventRegistration({ display_name: "Loose Player" });
    const lobby = buildEventLobby();
    const event = buildEventGroupEvent({ lobbies: [lobby], unplaced: [unplaced] });
    const { rerender } = render(<TeamsPanel {...baseProps({ event, isHostView: false })} />);
    expect(screen.queryByText(/Loose Player/)).not.toBeInTheDocument();
    expect(screen.queryByText(/Unplaced/)).not.toBeInTheDocument();

    rerender(<TeamsPanel {...baseProps({ event, isHostView: true })} />);
    expect(screen.getByText(/Unplaced · 1/)).toBeInTheDocument();
    expect(screen.getByText(/Loose Player/)).toBeInTheDocument();
  });

  it("passes Swap/Make Lobby Host actions through to team players for the host view", async () => {
    const onSwapPlayer = vi.fn();
    const player = buildLobbyPlayer({ user_id: "user-1" });
    const lobby = buildEventLobby({ host_id: "someone-else", teams: [buildEventTeam({ players: [player] })] });
    const event = buildEventGroupEvent({ lobbies: [lobby] });
    const user = userEvent.setup();
    render(<TeamsPanel {...baseProps({ event, isHostView: true, onSwapPlayer })} />);

    await user.click(screen.getByRole("button", { name: "Registration actions" }));
    await user.click(screen.getByText("Swap"));

    expect(onSwapPlayer).toHaveBeenCalledWith(
      expect.objectContaining({ eventId: event.id, userId: "user-1", lobbyId: lobby.id, teamNumber: 1 }),
    );
  });

  it("has no accessibility violations for a guest viewer", async () => {
    const player = buildLobbyPlayer({ display_name: "Alice" });
    const sub = buildLobbyPlayer({ display_name: "Subby" });
    const lobby = buildEventLobby({
      fairness_warning: true,
      fairness_warning_at_lock: true,
      teams: [buildEventTeam({ players: [player] })],
      subs: [sub],
    });
    const event = buildEventGroupEvent({ lobbies: [lobby] });
    const { container } = render(<TeamsPanel {...baseProps({ event, showJoinLobby: true })} />);
    expect(await axe(container)).toHaveNoViolations();
  });

  it("has no accessibility violations for a host with unplaced players", async () => {
    const player = buildLobbyPlayer({ avg_rank_order: 1, display_name: "Alice" });
    const unplaced = buildEventRegistration({ display_name: "Loose Player" });
    const lobby = buildEventLobby({ teams: [buildEventTeam({ players: [player] })] });
    const event = buildEventGroupEvent({ lobbies: [lobby], unplaced: [unplaced] });
    const gameRanks = [buildGameRank({ order: 1, name: "Gold" })];
    const { container } = render(
      <TeamsPanel {...baseProps({ event, isHostView: true, gameRanks, showJoinLobby: true })} />,
    );
    expect(await axe(container)).toHaveNoViolations();
  });
});
