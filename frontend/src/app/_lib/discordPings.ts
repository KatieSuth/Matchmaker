import {
  EventGroupDetail,
  EventGroupEvent,
  EventLobby,
} from "@/app/_types/types";

/** Backend always forms two teams per lobby; display numbers continue across lobbies within a game. */
export const TEAMS_PER_LOBBY = 2;

/**
 * Converts a per-lobby backend team_number (1|2) into a game-scoped display number.
 * Lobby 0 → 1,2; Lobby 1 → 3,4; etc.
 */
export function sequentialTeamNumber(
  lobbyIndex: number,
  teamNumber: number,
  teamsPerLobby: number = TEAMS_PER_LOBBY,
): number {
  return lobbyIndex * teamsPerLobby + teamNumber;
}

/** Resolves the lobby host's discord name from roster, subs, or unplaced players. */
export function resolveLobbyHostDiscordName(
  lobby: EventLobby,
  event: EventGroupEvent,
): string | null {
  if (!lobby.host_id) return null;
  const allPlayers = [
    ...lobby.teams.flatMap((team) => team.players),
    ...lobby.subs,
    ...(event.unplaced ?? []),
  ];
  const host = allPlayers.find((p) => p.user_id === lobby.host_id);
  return host?.discord_name ?? null;
}

function discordTimestamp(startTime: string): string {
  const unix = Math.floor(Date.parse(startTime) / 1000);
  return `<t:${unix}:F> (<t:${unix}:R>)`;
}

function mentionLine(discordName: string): string {
  return `@${discordName}`;
}

function buildTitle(group: EventGroupDetail): string {
  const name = group.name.trim();
  if (name) {
    return `# Here are the teams for ${name}!`;
  }

  const events = group.events ?? [];
  const singleMatch =
    events.length === 1 && (events[0].lobbies?.length ?? 0) === 1;
  const noun = singleMatch ? "match" : "matches";
  return `# Here are the teams for the upcoming ${group.game_name} custom ${noun}!`;
}

function buildGameSection(event: EventGroupEvent, gameIndex: number, gameCount: number): string {
  const lobbies = event.lobbies ?? [];
  const time = discordTimestamp(event.start_time);
  const heading =
    gameCount === 1 ? `## ${time}` : `## Game ${gameIndex + 1} - ${time}`;

  const parts: string[] = [heading];

  for (let lobbyIndex = 0; lobbyIndex < lobbies.length; lobbyIndex++) {
    const lobby = lobbies[lobbyIndex];
    const hostName = resolveLobbyHostDiscordName(lobby, event);
    const hostHeadingBase =
      lobbies.length === 1 ? "### Lobby Host" : `### Lobby ${lobbyIndex + 1} Host`;
    parts.push(hostName ? `${hostHeadingBase}: ${mentionLine(hostName)}` : hostHeadingBase);

    const teams = [...lobby.teams].sort((a, b) => a.team_number - b.team_number);
    for (const team of teams) {
      const displayNumber = sequentialTeamNumber(lobbyIndex, team.team_number);
      parts.push(`### Team ${displayNumber}`);
      for (const player of team.players) {
        if (player.discord_name) {
          parts.push(mentionLine(player.discord_name));
        }
      }
    }
  }

  const subs = lobbies.flatMap((lobby) => lobby.subs);
  parts.push("### Substitutes");
  if (subs.length === 0) {
    parts.push("There are no substitutes available");
  } else {
    for (const player of subs) {
      if (player.discord_name) {
        parts.push(mentionLine(player.discord_name));
      }
    }
  }

  return parts.join("\n");
}

/** Builds a Discord-markdown roster message for the event host to paste into chat. */
export function buildDiscordPingMessage(group: EventGroupDetail): string {
  const events = group.events ?? [];
  const sections = [buildTitle(group)];
  for (let i = 0; i < events.length; i++) {
    if ((events[i].lobbies?.length ?? 0) === 0) continue;
    sections.push(buildGameSection(events[i], i, events.length));
  }
  return sections.join("\n\n");
}
