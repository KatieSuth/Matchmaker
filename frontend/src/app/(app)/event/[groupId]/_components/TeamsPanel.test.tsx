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

  it("labels absent players on teams and in the substitute pool", () => {
    const teamPlayer = buildLobbyPlayer({ user_id: "team-absent", display_name: "Gone" });
    const presentPlayer = buildLobbyPlayer({ user_id: "team-present", display_name: "Here" });
    const swappedSub = buildLobbyPlayer({ user_id: "sub-absent", display_name: "Swapped" });
    const lobby = buildEventLobby({
      teams: [buildEventTeam({ players: [teamPlayer, presentPlayer] })],
      subs: [swappedSub],
    });
    const event = buildEventGroupEvent({ lobbies: [lobby], absent_user_ids: ["team-absent", "sub-absent"] });
    render(<TeamsPanel {...baseProps({ event })} />);

    expect(screen.getAllByText("Absent")).toHaveLength(2);
  });

  it("hides the attendance toggle until the game has started", async () => {
    const player = buildLobbyPlayer({ user_id: "user-1" });
    const lobby = buildEventLobby({ teams: [buildEventTeam({ players: [player] })] });
    const event = buildEventGroupEvent({ lobbies: [lobby], start_time: "2099-01-01T00:00:00Z" });
    const user = userEvent.setup();
    render(<TeamsPanel {...baseProps({ event, isHostView: true, onToggleAbsent: vi.fn() })} />);

    await user.click(screen.getByRole("button", { name: "Registration actions" }));
    expect(screen.queryByText(/Mark (Absent|Present)/)).not.toBeInTheDocument();
  });

  it("passes the attendance toggle to team players only for the host view", async () => {
    const onToggleAbsent = vi.fn();
    const player = buildLobbyPlayer({ user_id: "user-1" });
    const sub = buildLobbyPlayer({ user_id: "sub-1" });
    const lobby = buildEventLobby({ teams: [buildEventTeam({ players: [player] })], subs: [sub] });
    const event = buildEventGroupEvent({
      lobbies: [lobby],
      absent_user_ids: ["user-1"],
      start_time: "2020-01-01T00:00:00Z",
    });
    const user = userEvent.setup();
    render(<TeamsPanel {...baseProps({ event, isHostView: true, onToggleAbsent })} />);

    const [teamMenu, subMenu] = screen.getAllByRole("button", { name: "Registration actions" });
    await user.click(subMenu);
    expect(screen.queryByText(/Mark (Absent|Present)/)).not.toBeInTheDocument();

    await user.click(teamMenu);
    await user.click(screen.getByText("Mark Present"));
    expect(onToggleAbsent).toHaveBeenCalledWith(
      expect.objectContaining({ eventId: event.id, userId: "user-1", teamNumber: 1 }),
      false,
    );
  });

  it("flags excessive no-shows on team, substitute, and unplaced cards for the host", () => {
    const teamPlayer = buildLobbyPlayer({ user_id: "team-1" });
    const sub = buildLobbyPlayer({ user_id: "sub-1" });
    const clean = buildLobbyPlayer({ user_id: "clean-1" });
    const unplaced = buildEventRegistration({ user_id: "loose-1" });
    const lobby = buildEventLobby({ teams: [buildEventTeam({ players: [teamPlayer, clean] })], subs: [sub] });
    const event = buildEventGroupEvent({ lobbies: [lobby], unplaced: [unplaced] });
    const excessiveNoShows = { "team-1": 3, "sub-1": 4, "loose-1": 5 };

    const { rerender } = render(<TeamsPanel {...baseProps({ event, isHostView: true, excessiveNoShows })} />);
    expect(screen.getAllByRole("button", { name: /^No-show \d+ times/ })).toHaveLength(3);

    rerender(<TeamsPanel {...baseProps({ event, isHostView: false, excessiveNoShows })} />);
    expect(screen.queryByRole("button", { name: /^No-show/ })).not.toBeInTheDocument();
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

  it("has no accessibility violations for a host with absences and no-show warnings", async () => {
    const teamPlayer = buildLobbyPlayer({ user_id: "team-1", display_name: "Gone" });
    const other = buildLobbyPlayer({ user_id: "team-2", display_name: "Here" });
    const sub = buildLobbyPlayer({ user_id: "sub-1", display_name: "Subby" });
    const unplaced = buildEventRegistration({ user_id: "loose-1", display_name: "Loose" });
    const lobby = buildEventLobby({
      fairness_warning: true,
      fairness_warning_at_lock: true,
      fairness_outlier: true,
      teams: [buildEventTeam({ players: [teamPlayer, other] })],
      subs: [sub],
    });
    const event = buildEventGroupEvent({
      lobbies: [lobby],
      unplaced: [unplaced],
      absent_user_ids: ["team-1", "sub-1"],
      start_time: "2020-01-01T00:00:00Z",
    });
    const { container } = render(
      <TeamsPanel
        {...baseProps({
          event,
          isHostView: true,
          onToggleAbsent: vi.fn(),
          excessiveNoShows: { "team-1": 3, "sub-1": 4, "loose-1": 5 },
        })}
      />,
    );
    expect(screen.getAllByText("Absent")).toHaveLength(2);
    expect(await axe(container)).toHaveNoViolations();
  });
});
