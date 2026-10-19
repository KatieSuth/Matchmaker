// Lobby fairness-warning predicates/copy for the event group detail page.
import { EventGroupEvent, EventLobby } from "@/app/_types/types";

/** True when any lobby in the game is currently flagged unfair. */
export function eventHasUnfairLobby(event: EventGroupEvent): boolean {
  return (event.lobbies ?? []).some((lobby) => lobby.fairness_warning);
}

/** Explains which fairness check(s) failed, or null when the lobby predates recorded reasons. */
function fairnessReason(lobby: EventLobby): string | null {
  const outlier = "one player is ranked far above everyone else in this lobby";
  const teamGap = "the two teams' average ranks are too far apart";
  if (lobby.fairness_outlier && lobby.fairness_team_separation) {
    return `${outlier}, and ${teamGap}`;
  }
  if (lobby.fairness_outlier) return outlier;
  if (lobby.fairness_team_separation) return teamGap;
  return null;
}

/** Returns the lobby fairness banner copy for lock-in vs post-edit warnings, naming the failed check. */
export function lobbyFairnessWarningMessage(lobby: EventLobby): string {
  const reason = fairnessReason(lobby);
  if (lobby.fairness_warning_at_lock) {
    return reason
      ? `Teams were formed with the best available balance, but ${reason}.`
      : "Teams were formed with the best available balance, but rank spread was too wide for fully fair teams in this lobby.";
  }
  return reason
    ? `This lobby was fair when teams were locked in, but after a manual roster change ${reason}.`
    : "This lobby was fair when teams were locked in, but a manual roster change has made the rank spread too wide for fully fair teams.";
}
