// Builds and downloads an .ics calendar invitation for an event group. Every viewer's invitation
// spans the whole event (first game's start to last game's end); the description lists the games
// a player registered for, or every game for the host and for viewers not yet registered.
import { EventGroupDetail, EventGroupEvent } from "@/app/_types/types";
import { formatDateTime } from "./formatters";

/** Description headings: a registered player's games, or every game (host / not yet registered). */
export const PLAYER_GAMES_HEADING = "Games you registered for:";
export const HOST_GAMES_HEADING = "Games in this event:";

/** Reminder offsets before the start, as RFC 5545 durations: 24 hours, 90 minutes, 10 minutes. */
export const CALENDAR_REMINDERS = ["-PT24H", "-PT90M", "-PT10M"] as const;

/** The event's own label: its name, or "Game Mode" (e.g. "Valorant 5v5") when unnamed. */
export function calendarBaseTitle(group: EventGroupDetail): string {
  const name = group.name.trim();
  if (name) return name;
  return [group.game_name.trim(), group.game_mode_name.trim()].filter(Boolean).join(" ");
}

/**
 * Calendar entry title: the event's label plus its host, e.g. "Friday Customs (hosted by Hosty)",
 * so entries from different hosts' events are distinguishable. The host is their display name,
 * else their Discord name; the suffix is dropped if neither is set.
 */
export function calendarTitle(group: EventGroupDetail): string {
  const base = calendarBaseTitle(group);
  const host = group.owner_display_name.trim() || group.owner_name.trim();
  return host ? `${base} (hosted by ${host})` : base;
}

/**
 * The span of the given games: the earliest one's start to the latest one's end (start + mode
 * duration), with the games in start order. Null when none of the group's games are given.
 */
export function gameWindow(
  group: EventGroupDetail,
  eventIds: ReadonlySet<string>,
): { start: Date; end: Date; games: EventGroupEvent[] } | null {
  const games = group.events
    .filter((event) => eventIds.has(event.id))
    .sort((a, b) => Date.parse(a.start_time) - Date.parse(b.start_time));
  if (games.length === 0) return null;

  const start = new Date(games[0].start_time);
  let endMs = start.getTime();
  for (const game of games) {
    endMs = Math.max(endMs, Date.parse(game.start_time) + Math.max(game.duration_minutes, 0) * 60_000);
  }
  return { start, end: new Date(endMs), games };
}

/** Formats a Date as an RFC 5545 UTC date-time, e.g. 20260904T190000Z. */
export function formatIcsDate(date: Date): string {
  return date.toISOString().replace(/[-:]/g, "").replace(/\.\d{3}Z$/, "Z");
}

/** Escapes TEXT values per RFC 5545: backslash, semicolon, comma, and newlines. */
export function escapeIcsText(value: string): string {
  return value.replace(/\\/g, "\\\\").replace(/;/g, "\\;").replace(/,/g, "\\,").replace(/\r?\n/g, "\\n");
}

/** Folds a content line at 75 octets (continuations start with a space), never splitting a character. */
export function foldIcsLine(line: string): string {
  const encoder = new TextEncoder();
  const parts: string[] = [];
  let current = "";
  let currentBytes = 0;
  for (const char of line) {
    const charBytes = encoder.encode(char).length;
    // The first line holds 75 octets; continuation lines hold 74 after their leading space.
    const limit = parts.length === 0 ? 75 : 74;
    if (currentBytes + charBytes > limit) {
      parts.push(current);
      current = "";
      currentBytes = 0;
    }
    current += char;
    currentBytes += charBytes;
  }
  parts.push(current);
  return parts.join("\r\n ");
}

/**
 * Builds the .ics text for the group, or null when it has no games. The entry always spans the
 * whole event; listedEventIds picks the games named in the description (empty lists every game).
 * The UID is stable per group so re-downloading updates the same calendar entry in clients that
 * honor it.
 */
export function buildCalendarInvite(
  group: EventGroupDetail,
  listedEventIds: ReadonlySet<string>,
  pageUrl: string,
  now: Date = new Date(),
  timeZone?: string,
  heading: string = PLAYER_GAMES_HEADING,
): string | null {
  const window = gameWindow(group, new Set(group.events.map((event) => event.id)));
  if (!window) return null;
  const listed = (listedEventIds.size > 0 && gameWindow(group, listedEventIds)?.games) || window.games;

  const gameLines = listed.map((game) => {
    const number = group.events.findIndex((event) => event.id === game.id) + 1;
    return `• Game ${number} (${game.game_mode_name}): ${formatDateTime(game.start_time, timeZone)}`;
  });
  const description = [heading, ...gameLines, "", `Event page: ${pageUrl}`].join("\n");

  const lines = [
    "BEGIN:VCALENDAR",
    "VERSION:2.0",
    "PRODID:-//Matchmaker//Event Invitation//EN",
    "CALSCALE:GREGORIAN",
    "METHOD:PUBLISH",
    "BEGIN:VEVENT",
    `UID:${group.id}@matchmaker`,
    `DTSTAMP:${formatIcsDate(now)}`,
    `DTSTART:${formatIcsDate(window.start)}`,
    `DTEND:${formatIcsDate(window.end)}`,
    `SUMMARY:${escapeIcsText(calendarTitle(group))}`,
    `DESCRIPTION:${escapeIcsText(description)}`,
    `URL:${pageUrl}`,
    ...CALENDAR_REMINDERS.flatMap((trigger) => [
      "BEGIN:VALARM",
      "ACTION:DISPLAY",
      `DESCRIPTION:${escapeIcsText(calendarTitle(group))}`,
      `TRIGGER:${trigger}`,
      "END:VALARM",
    ]),
    "END:VEVENT",
    "END:VCALENDAR",
  ];
  return lines.map(foldIcsLine).join("\r\n") + "\r\n";
}

/** File name for the invitation: the event's label (no host) lower-cased with non-alphanumerics as dashes. */
export function calendarFileName(group: EventGroupDetail): string {
  const slug = calendarBaseTitle(group)
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
  return `${slug || "event"}.ics`;
}

/** Triggers a browser download of the group's calendar invitation. No-op when it has no games. */
export function downloadCalendarInvite(
  group: EventGroupDetail,
  listedEventIds: ReadonlySet<string>,
  pageUrl: string,
  heading: string = PLAYER_GAMES_HEADING,
): void {
  const ics = buildCalendarInvite(group, listedEventIds, pageUrl, new Date(), undefined, heading);
  if (!ics) return;
  const url = URL.createObjectURL(new Blob([ics], { type: "text/calendar;charset=utf-8" }));
  const link = document.createElement("a");
  link.href = url;
  link.download = calendarFileName(group);
  document.body.appendChild(link);
  link.click();
  link.remove();
  URL.revokeObjectURL(url);
}
