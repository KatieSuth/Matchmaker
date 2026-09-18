// Pure display-formatting helpers for the event group detail page.
import { EMPTY_VALUE } from "@/app/_lib/constants";
import { sequentialTeamNumber } from "@/app/_lib/discordPings";
import { formatDateTime as formatDateTimeValue } from "@/app/_lib/dateTime";
import { formatUserDisplayLabel } from "@/app/_lib/userDisplayName";
import { EventGroupEvent } from "@/app/_types/types";

/**
 * Formats an ISO date-time string as a medium date + short time (e.g. "Sep 4, 2026, 3:00 PM").
 * Delegates to the shared `formatDateTime` (which builds its `Intl.DateTimeFormat` per call rather
 * than caching one at module scope) so `timeZone` can be overridden in tests.
 */
export function formatDateTime(value: string, timeZone?: string) {
  if (!value.trim()) return EMPTY_VALUE;
  const date = new Date(value);
  if (!Number.isFinite(date.getTime())) return EMPTY_VALUE;
  return formatDateTimeValue(date, timeZone);
}

/** Joins Discord server names for the lock-denial sentence (or / Oxford or). */
export function joinDiscordGuildNames(names: string[]): string {
  if (names.length === 0) return "a required Discord server";
  if (names.length === 1) return names[0];
  if (names.length === 2) return `${names[0]} or ${names[1]}`;
  return `${names.slice(0, -1).join(", ")}, or ${names[names.length - 1]}`;
}

/** Body copy for the Discord lock denial: named events keep the host title; unnamed use the game name. */
export function discordLockLeadSentence(eventNamed: boolean, title: string, serverNames: string): string {
  if (eventNamed && title) {
    return `${title} is locked to ${serverNames}.`;
  }
  const game = title || "game";
  return `This ${game} event is locked to ${serverNames}.`;
}

export function formatPlayerCount(count: number) {
  return `${count} ${count === 1 ? "Player" : "Players"}`;
}

export function formatGroupTeamSizeLabel(events: EventGroupEvent[]) {
  const sizes = [...new Set(events.map((e) => e.team_size))];
  if (sizes.length === 0) return "—";
  if (sizes.length === 1) return String(sizes[0]);
  return "Varies";
}

export function formatHostDisplayLabel(
  isViewerHost: boolean,
  ownerDisplayName: string,
  ownerName: string,
  ownerPronouns: string,
) {
  if (isViewerHost) return "You";
  const label = formatUserDisplayLabel(ownerDisplayName, ownerName);
  const pronouns = ownerPronouns.trim();
  if (pronouns) return `${label} (${pronouns})`;
  return label;
}

/** Compact subtitle fragment for chips / roster rows (game mode · formatted time). */
export function formatGameModeAndTime(modeName: string, startISO: string) {
  return `${modeName} · ${formatDateTime(startISO)}`;
}

/** Human-readable placement label for swap candidates (team, subs, or unplaced). */
export function formatPlacementCategory(
  sourceLobbyIndex: number | null,
  targetLobbyIndex: number,
  teamNumber: number | null | undefined,
): string {
  if (teamNumber === undefined) {
    return "Unplaced";
  }
  if (teamNumber === null) {
    // Substitutes are a single per-game pool in the UI (not per lobby).
    return "Subs";
  }
  const displayTeam = sequentialTeamNumber(targetLobbyIndex, teamNumber);
  if (sourceLobbyIndex !== null && sourceLobbyIndex === targetLobbyIndex) {
    return `Team ${displayTeam}`;
  }
  return `Lobby ${targetLobbyIndex + 1} · Team ${displayTeam}`;
}

/** Formats a swap dropdown option as "Name (Category) · Rank". */
export function formatSwapCandidateLabel(
  displayName: string,
  discordName: string,
  category: string,
  currentRankName?: string,
): string {
  const name = formatUserDisplayLabel(displayName, discordName);
  const rank = currentRankName?.trim() || EMPTY_VALUE;
  return `${name} (${category}) · ${rank}`;
}
