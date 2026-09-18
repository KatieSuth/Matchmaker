// Roster-placement predicates and swap/host-volunteer candidate builders for the event group page.
import { SelectOption } from "@/app/_components/Select";
import { sequentialTeamNumber } from "@/app/_lib/discordPings";
import { formatUserDisplayLabel } from "@/app/_lib/userDisplayName";
import {
  EventGroupEvent,
  EventLobby,
  EventRegistration,
  LobbyPlayer,
} from "@/app/_types/types";
import { formatPlacementCategory, formatSwapCandidateLabel } from "./formatters";
import { LobbyHostVolunteer, PlayerPlacement } from "../_types";

/** True when a player is assigned to a team (not subs or unplaced). */
export function isTeamAssignedPlacement(
  placement?: PlayerPlacement,
): placement is PlayerPlacement & { lobbyId: string; teamNumber: number } {
  return !!placement && placement.lobbyId !== null && typeof placement.teamNumber === "number";
}

/** True when a player is in a lobby sub pool. */
export function isSubPlacement(
  placement?: PlayerPlacement,
): placement is PlayerPlacement & { lobbyId: string; teamNumber: null } {
  return !!placement && placement.lobbyId !== null && placement.teamNumber === null;
}

/** True when a player is registered but not on a team or sub pool. */
export function isUnplacedPlacement(
  placement?: PlayerPlacement,
): placement is PlayerPlacement & { lobbyId: null; teamNumber: undefined } {
  return !!placement && placement.teamNumber === undefined;
}

/** Lists team players in a lobby who volunteered to host, excluding the selected player. */
export function buildLobbyHostVolunteers(
  lobby: EventLobby,
  lobbyIndex: number,
  excludeUserId: string,
  lobbyHostId?: string | null,
): LobbyHostVolunteer[] {
  const volunteers: LobbyHostVolunteer[] = [];
  for (const team of lobby.teams) {
    for (const player of team.players) {
      if (player.user_id === excludeUserId || !player.can_lobby_host) {
        continue;
      }
      volunteers.push({
        userId: player.user_id,
        discordName: formatUserDisplayLabel(player.display_name, player.discord_name),
        teamNumber: sequentialTeamNumber(lobbyIndex, team.team_number),
        isCurrentHost: !!lobbyHostId && player.user_id === lobbyHostId,
      });
    }
  }
  return volunteers;
}

/** Lists eligible swap targets for a player, excluding same-team roster mates and other substitutes when the source is a sub. */
export function buildSwapCandidates(event: EventGroupEvent, source: PlayerPlacement): SelectOption[] {
  const options: SelectOption[] = [];
  const lobbies = event.lobbies ?? [];
  const sourceIsSub = isSubPlacement(source);
  const sourceIsUnplaced = isUnplacedPlacement(source);

  for (let lobbyIndex = 0; lobbyIndex < lobbies.length; lobbyIndex++) {
    const lobby = lobbies[lobbyIndex];
    for (const team of lobby.teams) {
      for (const player of team.players) {
        if (player.user_id === source.userId) {
          continue;
        }
        if (
          source.teamNumber !== undefined &&
          source.teamNumber !== null &&
          source.lobbyId === lobby.id &&
          source.teamNumber === team.team_number
        ) {
          continue;
        }
        const category = formatPlacementCategory(source.sourceLobbyIndex, lobbyIndex, team.team_number);
        options.push({
          value: player.user_id,
          label: formatSwapCandidateLabel(player.display_name, player.discord_name, category, player.avg_rank_name),
        });
      }
    }
    for (const player of lobby.subs) {
      if (player.user_id === source.userId) {
        continue;
      }
      // Unplaced cannot swap with subs; subs cannot swap with other game subs.
      if (sourceIsUnplaced || sourceIsSub) {
        continue;
      }
      const category = formatPlacementCategory(source.sourceLobbyIndex, lobbyIndex, null);
      options.push({
        value: player.user_id,
        label: formatSwapCandidateLabel(player.display_name, player.discord_name, category, player.avg_rank_name),
      });
    }
  }

  for (const registration of event.unplaced ?? []) {
    if (registration.user_id === source.userId) {
      continue;
    }
    // Sub↔unplaced and unplaced↔unplaced are rejected by the backend.
    if (sourceIsSub || sourceIsUnplaced) {
      continue;
    }
    options.push({
      value: registration.user_id,
      label: formatSwapCandidateLabel(
        registration.display_name,
        registration.discord_name,
        "Unplaced",
        registration.avg_rank_name,
      ),
    });
  }

  options.sort((a, b) => a.label.localeCompare(b.label));
  return options;
}

/** Lists every lobby the given user hosts across all games in the group. */
export function findUserLobbyHostAssignments(
  events: EventGroupEvent[],
  userId: string,
): { gameNumber: number; lobbyNumber: number }[] {
  const assignments: { gameNumber: number; lobbyNumber: number }[] = [];
  events.forEach((event, eventIndex) => {
    (event.lobbies ?? []).forEach((lobby, lobbyIndex) => {
      if (lobby.host_id === userId) {
        assignments.push({
          gameNumber: eventIndex + 1,
          lobbyNumber: lobbyIndex + 1,
        });
      }
    });
  });
  return assignments;
}

/** Adapts a lobby player row so TeamsPanel can reuse PlayerCard and registration actions. */
export function lobbyPlayerAsRegistration(player: LobbyPlayer, eventId: string): EventRegistration {
  return {
    event_id: eventId,
    user_id: player.user_id,
    discord_name: player.discord_name,
    display_name: player.display_name,
    in_game_name: player.in_game_name,
    pronouns: player.pronouns,
    current_rank_name: player.current_rank_name,
    peak_rank_name: player.peak_rank_name,
    avg_rank_name: player.avg_rank_name,
    can_substitute: player.can_substitute,
    can_lobby_host: player.can_lobby_host,
    duo_request: player.duo_request,
    created_at: player.created_at,
    updated_at: player.updated_at,
  };
}

/** Resolves the lobby host's UI label from roster, subs, or unplaced players. */
export function lobbyHostName(lobby: EventLobby, event: EventGroupEvent): string | null {
  if (!lobby.host_id) return null;
  const allPlayers = [
    ...lobby.teams.flatMap((team) => team.players),
    ...lobby.subs,
    ...(event.unplaced ?? []),
  ];
  const host = allPlayers.find((p) => p.user_id === lobby.host_id);
  if (!host) return null;
  return formatUserDisplayLabel(host.display_name, host.discord_name);
}
