// Tests for roster-placement predicates and the swap/host-volunteer candidate builders. These
// drive the host's swap-player and set-lobby-host dropdowns, so incorrect eligibility here would
// let a host attempt swaps the backend will reject (or silently omit valid candidates).
import { describe, expect, it } from "vitest";
import { buildEventGroupEvent, buildEventLobby, buildEventRegistration, buildEventTeam, buildLobbyPlayer } from "@/test/fixtures";
import { PlayerPlacement } from "../_types";
import {
  buildLobbyHostVolunteers,
  buildSwapCandidates,
  findUserLobbyHostAssignments,
  isSubPlacement,
  isTeamAssignedPlacement,
  isUnplacedPlacement,
  lobbyHostName,
  lobbyPlayerAsRegistration,
} from "./placements";

function placement(overrides: Partial<PlayerPlacement>): PlayerPlacement {
  return {
    eventId: "event-1",
    userId: "user-1",
    discordName: "User#0001",
    lobbyId: null,
    sourceLobbyIndex: null,
    teamNumber: undefined,
    ...overrides,
  };
}

describe("placement predicates", () => {
  it("isTeamAssignedPlacement is true only for a lobbyId + numeric teamNumber", () => {
    expect(isTeamAssignedPlacement(placement({ lobbyId: "lobby-1", teamNumber: 1 }))).toBe(true);
    expect(isTeamAssignedPlacement(placement({ lobbyId: "lobby-1", teamNumber: null }))).toBe(false);
    expect(isTeamAssignedPlacement(placement({ lobbyId: null, teamNumber: 1 }))).toBe(false);
    expect(isTeamAssignedPlacement(undefined)).toBe(false);
  });

  it("isSubPlacement is true only for a lobbyId + null teamNumber", () => {
    expect(isSubPlacement(placement({ lobbyId: "lobby-1", teamNumber: null }))).toBe(true);
    expect(isSubPlacement(placement({ lobbyId: "lobby-1", teamNumber: 1 }))).toBe(false);
    expect(isSubPlacement(placement({ lobbyId: null, teamNumber: null }))).toBe(false);
  });

  it("isUnplacedPlacement is true only when teamNumber is undefined", () => {
    expect(isUnplacedPlacement(placement({ lobbyId: null, teamNumber: undefined }))).toBe(true);
    expect(isUnplacedPlacement(placement({ lobbyId: null, teamNumber: null }))).toBe(false);
    expect(isUnplacedPlacement(placement({ lobbyId: "lobby-1", teamNumber: 1 }))).toBe(false);
  });
});

describe("buildLobbyHostVolunteers", () => {
  it("lists team players who volunteered, excluding the given player", () => {
    const volunteer = buildLobbyPlayer({ user_id: "v1", can_lobby_host: true, display_name: "Vee", discord_name: "vee#0001" });
    const excluded = buildLobbyPlayer({ user_id: "current", can_lobby_host: true });
    const notVolunteering = buildLobbyPlayer({ user_id: "v2", can_lobby_host: false });
    const lobby = buildEventLobby({ teams: [buildEventTeam({ team_number: 1, players: [volunteer, excluded, notVolunteering] })] });

    const result = buildLobbyHostVolunteers(lobby, 0, "current");

    expect(result).toEqual([
      { userId: "v1", discordName: "Vee @vee#0001", teamNumber: 1, isCurrentHost: false },
    ]);
  });

  it("marks the current lobby host", () => {
    const volunteer = buildLobbyPlayer({ user_id: "v1", can_lobby_host: true });
    const lobby = buildEventLobby({ host_id: "v1", teams: [buildEventTeam({ players: [volunteer] })] });

    const result = buildLobbyHostVolunteers(lobby, 0, "someone-else", "v1");

    expect(result[0].isCurrentHost).toBe(true);
  });
});

describe("findUserLobbyHostAssignments", () => {
  it("finds every lobby a user hosts across games, 1-indexed", () => {
    const events = [
      buildEventGroupEvent({ lobbies: [buildEventLobby({ host_id: "other" }), buildEventLobby({ host_id: "user-1" })] }),
      buildEventGroupEvent({ lobbies: [buildEventLobby({ host_id: "user-1" })] }),
    ];

    expect(findUserLobbyHostAssignments(events, "user-1")).toEqual([
      { gameNumber: 1, lobbyNumber: 2 },
      { gameNumber: 2, lobbyNumber: 1 },
    ]);
  });

  it("returns an empty array when the user hosts nothing", () => {
    const events = [buildEventGroupEvent({ lobbies: [buildEventLobby({ host_id: "other" })] })];
    expect(findUserLobbyHostAssignments(events, "user-1")).toEqual([]);
  });
});

describe("lobbyPlayerAsRegistration", () => {
  it("maps every shared field from LobbyPlayer to EventRegistration", () => {
    const player = buildLobbyPlayer({ user_id: "p1", discord_name: "P#0001" });
    const registration = lobbyPlayerAsRegistration(player, "event-1");

    expect(registration).toMatchObject({
      event_id: "event-1",
      user_id: "p1",
      discord_name: "P#0001",
      display_name: player.display_name,
      in_game_name: player.in_game_name,
      can_substitute: player.can_substitute,
      can_lobby_host: player.can_lobby_host,
    });
  });
});

describe("lobbyHostName", () => {
  it("returns null when there is no host", () => {
    const lobby = buildEventLobby({ host_id: null });
    const event = buildEventGroupEvent({ unplaced: [] });
    expect(lobbyHostName(lobby, event)).toBeNull();
  });

  it("returns the formatted display label of the host", () => {
    const host = buildLobbyPlayer({ user_id: "h1", display_name: "Host", discord_name: "host#0001" });
    const lobby = buildEventLobby({ host_id: "h1", teams: [buildEventTeam({ players: [host] })] });
    const event = buildEventGroupEvent({ unplaced: [] });
    expect(lobbyHostName(lobby, event)).toBe("Host @host#0001");
  });

  it("returns null when host_id doesn't match any known player", () => {
    const lobby = buildEventLobby({ host_id: "missing", teams: [] });
    const event = buildEventGroupEvent({ unplaced: [] });
    expect(lobbyHostName(lobby, event)).toBeNull();
  });
});

describe("buildSwapCandidates", () => {
  it("excludes the source player and same-team roster mates", () => {
    const source = placement({ userId: "p1", lobbyId: "lobby-1", sourceLobbyIndex: 0, teamNumber: 1 });
    const sourcePlayer = buildLobbyPlayer({ user_id: "p1" });
    const teammate = buildLobbyPlayer({ user_id: "p1b" });
    const otherTeamPlayer = buildLobbyPlayer({ user_id: "p3", display_name: "P3", discord_name: "p3#0001" });
    const lobby = buildEventLobby({
      id: "lobby-1",
      teams: [
        buildEventTeam({ team_number: 1, players: [sourcePlayer, teammate] }),
        buildEventTeam({ team_number: 2, players: [otherTeamPlayer] }),
      ],
      subs: [],
    });
    const event = buildEventGroupEvent({ lobbies: [lobby], unplaced: [] });

    const candidates = buildSwapCandidates(event, source);

    // Same-team players (p1's own team, team 1 — including the teammate, not just the source
    // player itself) are excluded; only the opposing team's player remains.
    expect(candidates.map((c) => c.value)).toEqual(["p3"]);
  });

  it("excludes sub<->sub and unplaced<->sub swaps but allows team<->sub", () => {
    const sub = buildLobbyPlayer({ user_id: "sub-1" });
    const teamPlayer = buildLobbyPlayer({ user_id: "team-1" });
    const lobby = buildEventLobby({
      id: "lobby-1",
      teams: [buildEventTeam({ team_number: 1, players: [teamPlayer] })],
      subs: [sub],
    });
    const event = buildEventGroupEvent({ lobbies: [lobby], unplaced: [] });

    const fromSub = placement({ userId: "sub-1", lobbyId: "lobby-1", teamNumber: null, sourceLobbyIndex: 0 });
    // A sub source should not be offered other subs as swap targets.
    expect(buildSwapCandidates(event, fromSub).map((c) => c.value)).not.toContain("sub-1");

    const fromTeam = placement({ userId: "team-1", lobbyId: "lobby-1", teamNumber: 1, sourceLobbyIndex: 0 });
    expect(buildSwapCandidates(event, fromTeam).map((c) => c.value)).toContain("sub-1");
  });

  it("excludes unplaced<->unplaced swaps", () => {
    const unplacedA = buildEventRegistration({ user_id: "u-a" });
    const unplacedB = buildEventRegistration({ user_id: "u-b" });
    const event = buildEventGroupEvent({ lobbies: [], unplaced: [unplacedA, unplacedB] });

    const source = placement({ userId: "u-a", lobbyId: null, teamNumber: undefined, sourceLobbyIndex: null });
    expect(buildSwapCandidates(event, source)).toEqual([]);
  });

  it("allows a team player to swap with an unplaced registrant", () => {
    const teamPlayer = buildLobbyPlayer({ user_id: "team-1" });
    const lobby = buildEventLobby({ teams: [buildEventTeam({ players: [teamPlayer] })], subs: [] });
    const unplaced = buildEventRegistration({ user_id: "u-a", display_name: "Unplaced Guy", discord_name: "unplaced#0001" });
    const event = buildEventGroupEvent({ lobbies: [lobby], unplaced: [unplaced] });

    const source = placement({ userId: "team-1", lobbyId: lobby.id, teamNumber: 1, sourceLobbyIndex: 0 });
    expect(buildSwapCandidates(event, source).map((c) => c.value)).toContain("u-a");
  });

  it("sorts candidates alphabetically by label", () => {
    const p1 = buildLobbyPlayer({ user_id: "p1", display_name: "Zed", discord_name: "zed#0001" });
    const p2 = buildLobbyPlayer({ user_id: "p2", display_name: "Amy", discord_name: "amy#0001" });
    const lobby = buildEventLobby({
      teams: [buildEventTeam({ team_number: 1, players: [p1] }), buildEventTeam({ team_number: 2, players: [p2] })],
    });
    const event = buildEventGroupEvent({ lobbies: [lobby], unplaced: [] });
    const source = placement({ userId: "other", lobbyId: null, teamNumber: undefined, sourceLobbyIndex: null });

    const labels = buildSwapCandidates(event, source).map((c) => c.label);
    expect(labels).toEqual([...labels].sort((a, b) => a.localeCompare(b)));
  });
});
