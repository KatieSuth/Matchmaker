// Tests for the Discord-markdown roster message builder (host's "share teams" paste target).
import { describe, expect, it } from "vitest";
import {
  buildEventGroupDetail,
  buildEventGroupEvent,
  buildEventLobby,
  buildEventRegistration,
  buildEventTeam,
  buildLobbyPlayer,
} from "@/test/fixtures";
import { buildDiscordPingMessage, resolveLobbyHostDiscordName, sequentialTeamNumber } from "./discordPings";

describe("sequentialTeamNumber", () => {
  it("numbers lobby 0's teams 1 and 2 by default (2 teams per lobby)", () => {
    expect(sequentialTeamNumber(0, 1)).toBe(1);
    expect(sequentialTeamNumber(0, 2)).toBe(2);
  });

  it("continues numbering across lobbies", () => {
    expect(sequentialTeamNumber(1, 1)).toBe(3);
    expect(sequentialTeamNumber(1, 2)).toBe(4);
  });

  it("respects a custom teamsPerLobby", () => {
    expect(sequentialTeamNumber(1, 1, 3)).toBe(4);
  });
});

describe("resolveLobbyHostDiscordName", () => {
  const host = buildLobbyPlayer({ user_id: "host-1", discord_name: "Host#0001" });
  const other = buildLobbyPlayer({ user_id: "player-1", discord_name: "Player#0001" });

  it("returns null when the lobby has no host", () => {
    const lobby = buildEventLobby({ host_id: null, teams: [buildEventTeam({ players: [other] })] });
    const event = buildEventGroupEvent({ unplaced: [] });
    expect(resolveLobbyHostDiscordName(lobby, event)).toBeNull();
  });

  it("finds the host among team players", () => {
    const lobby = buildEventLobby({ host_id: "host-1", teams: [buildEventTeam({ players: [host, other] })] });
    const event = buildEventGroupEvent({ unplaced: [] });
    expect(resolveLobbyHostDiscordName(lobby, event)).toBe("Host#0001");
  });

  it("finds the host among subs", () => {
    const lobby = buildEventLobby({ host_id: "host-1", teams: [], subs: [host] });
    const event = buildEventGroupEvent({ unplaced: [] });
    expect(resolveLobbyHostDiscordName(lobby, event)).toBe("Host#0001");
  });

  it("returns null when the host_id doesn't match anyone (data inconsistency)", () => {
    const lobby = buildEventLobby({ host_id: "missing", teams: [buildEventTeam({ players: [other] })] });
    const event = buildEventGroupEvent({ unplaced: [] });
    expect(resolveLobbyHostDiscordName(lobby, event)).toBeNull();
  });

  it("finds the host among an event's unplaced players", () => {
    const lobby = buildEventLobby({ host_id: "host-1", teams: [] });
    const event = buildEventGroupEvent({
      unplaced: [buildEventRegistration({ user_id: "host-1", discord_name: "Host#0001" })],
    });
    expect(resolveLobbyHostDiscordName(lobby, event)).toBe("Host#0001");
  });
});

describe("buildDiscordPingMessage", () => {
  it("titles the message with the group name when set", () => {
    const group = buildEventGroupDetail({ name: "Friday Night Customs", events: [] });
    expect(buildDiscordPingMessage(group)).toContain("# Here are the teams for Friday Night Customs!");
  });

  it("falls back to a game-name title when the group is unnamed", () => {
    const group = buildEventGroupDetail({ name: "", game_name: "Valorant", events: [] });
    expect(buildDiscordPingMessage(group)).toContain("upcoming Valorant custom matches!");
  });

  it("skips games with no lobbies (not locked in yet)", () => {
    const noLobbiesEvent = buildEventGroupEvent({ lobbies: [] });
    const group = buildEventGroupDetail({ events: [noLobbiesEvent] });
    const message = buildDiscordPingMessage(group);
    expect(message).not.toContain("Lobby Host");
  });

  it("includes lobby host, team players, and substitutes for a locked-in game", () => {
    const p1 = buildLobbyPlayer({ user_id: "p1", discord_name: "PlayerOne#0001", can_lobby_host: true });
    const p2 = buildLobbyPlayer({ user_id: "p2", discord_name: "PlayerTwo#0002" });
    const sub = buildLobbyPlayer({ user_id: "sub1", discord_name: "SubOne#0003" });
    const lobby = buildEventLobby({
      host_id: "p1",
      teams: [buildEventTeam({ team_number: 1, players: [p1] }), buildEventTeam({ team_number: 2, players: [p2] })],
      subs: [sub],
    });
    const event = buildEventGroupEvent({ lobbies: [lobby], unplaced: [] });
    const group = buildEventGroupDetail({ name: "Test Group", events: [event] });

    const message = buildDiscordPingMessage(group);
    expect(message).toContain("### Lobby Host: @PlayerOne#0001");
    expect(message).toContain("### Team 1");
    expect(message).toContain("@PlayerOne#0001");
    expect(message).toContain("### Team 2");
    expect(message).toContain("@PlayerTwo#0002");
    expect(message).toContain("### Substitutes");
    expect(message).toContain("@SubOne#0003");
  });

  it("shows the 'no substitutes' message when a lobby has none", () => {
    const p1 = buildLobbyPlayer({ user_id: "p1" });
    const lobby = buildEventLobby({ teams: [buildEventTeam({ players: [p1] })], subs: [] });
    const event = buildEventGroupEvent({ lobbies: [lobby] });
    const group = buildEventGroupDetail({ events: [event] });

    expect(buildDiscordPingMessage(group)).toContain("There are no substitutes available");
  });

  it("labels a single-lobby game's host section without a lobby number", () => {
    const lobby = buildEventLobby({ host_id: null, teams: [buildEventTeam()] });
    const event = buildEventGroupEvent({ lobbies: [lobby] });
    const group = buildEventGroupDetail({ events: [event] });

    expect(buildDiscordPingMessage(group)).toContain("### Lobby Host");
    expect(buildDiscordPingMessage(group)).not.toContain("### Lobby 1 Host");
  });

  it("numbers lobby host sections when a game has multiple lobbies", () => {
    const lobbyA = buildEventLobby({ host_id: null, teams: [buildEventTeam()] });
    const lobbyB = buildEventLobby({ host_id: null, teams: [buildEventTeam()] });
    const event = buildEventGroupEvent({ lobbies: [lobbyA, lobbyB] });
    const group = buildEventGroupDetail({ events: [event] });

    const message = buildDiscordPingMessage(group);
    expect(message).toContain("### Lobby 1 Host");
    expect(message).toContain("### Lobby 2 Host");
  });

  it("numbers game sections and pluralizes the title when a group has multiple locked-in games", () => {
    const eventA = buildEventGroupEvent({ lobbies: [buildEventLobby()] });
    const eventB = buildEventGroupEvent({ lobbies: [buildEventLobby()] });
    const group = buildEventGroupDetail({ name: "", game_name: "Valorant", events: [eventA, eventB] });

    const message = buildDiscordPingMessage(group);
    expect(message).toContain("upcoming Valorant custom matches!");
    expect(message).toContain("## Game 1 -");
    expect(message).toContain("## Game 2 -");
  });

  it("skips a team/sub player with no discord_name (unlinked Discord account)", () => {
    const linked = buildLobbyPlayer({ discord_name: "Linked#0001" });
    const unlinked = buildLobbyPlayer({ discord_name: "" });
    const lobby = buildEventLobby({
      teams: [buildEventTeam({ players: [linked, unlinked] })],
      subs: [unlinked],
    });
    const event = buildEventGroupEvent({ lobbies: [lobby] });
    const group = buildEventGroupDetail({ events: [event] });

    const message = buildDiscordPingMessage(group);
    expect(message).toContain("@Linked#0001");
    // The unlinked player contributes no mention line at all (not even a blank "@").
    expect(message).not.toMatch(/^@$/m);
  });
});
