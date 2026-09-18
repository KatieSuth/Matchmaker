// Lobby fairness-warning predicates/copy for the event group detail page.
import { EventGroupEvent, EventLobby } from "@/app/_types/types";

/** True when any lobby in the game is currently flagged unfair. */
export function eventHasUnfairLobby(event: EventGroupEvent): boolean {
  return (event.lobbies ?? []).some((lobby) => lobby.fairness_warning);
}

/** Returns the lobby fairness banner copy for lock-in vs post-edit warnings. */
export function lobbyFairnessWarningMessage(lobby: EventLobby): string {
  if (lobby.fairness_warning_at_lock) {
    return "Teams were formed with the best available balance, but rank spread was too wide for fully fair teams in this lobby.";
  }
  return "This lobby was fair when teams were locked in, but a manual roster change has made the rank spread too wide for fully fair teams.";
}
