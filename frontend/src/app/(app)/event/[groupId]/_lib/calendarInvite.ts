// Builds and downloads an .ics calendar invitation covering a set of games in a group: the viewer's
// registered games, or every game for the host.
import { EventGroupDetail, EventGroupEvent } from "@/app/_types/types";
import { formatDateTime } from "./formatters";

/** Description headings for the two audiences of an invitation. */
export const PLAYER_GAMES_HEADING = "Games you registered for:";
export const HOST_GAMES_HEADING = "Games in this event:";

/** Reminder offsets before the start, as RFC 5545 durations: 24 hours, 90 minutes, 10 minutes. */
export const CALENDAR_REMINDERS = ["-PT24H", "-PT90M", "-PT10M"] as const;

/** Event title as the page header shows it: the group name, or the game name when unnamed. */
export function calendarTitle(group: EventGroupDetail): string {
  return group.name.trim() ? group.name : group.game_name;
}

/**
 * The span of the viewer's registered games: the first one's start to the last one's end
 * (start + mode duration). Null when they are registered for none of the group's games.
 */
export function registeredGameWindow(
  group: EventGroupDetail,
  registeredEventIds: ReadonlySet<string>,
): { start: Date; end: Date; games: EventGroupEvent[] } | null {
  const games = group.events
    .filter((event) => registeredEventIds.has(event.id))
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
  return value.replace(/\\/g, "\\\\").replace(/;/g, "\;").replace(/,/g, "\\,").replace(/\r?\n/g, "\\n");
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
 * Builds the .ics text for the given games (the viewer's registrations, or every game for the
 * host), or null when there are none. The UID is stable per group so re-downloading after a
 * change updates the same calendar entry in clients that honor it.
 */
export function buildCalendarInvite(
  group: EventGroupDetail,
  eventIds: ReadonlySet<string>,
  pageUrl: string,
  now: Date = new Date(),
  timeZone?: string,
  heading: string = PLAYER_GAMES_HEADING,
): string | null {
  const window = registeredGameWindow(group, eventIds);
  if (!window) return null;

  const gameLines = window.games.map((game) => {
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

/** File name for the invitation: the title lower-cased with non-alphanumerics collapsed to dashes. */
export function calendarFileName(group: EventGroupDetail): string {
  const slug = calendarTitle(group)
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
  return `${slug || "event"}.ics`;
}

/** Triggers a browser download of the calendar invitation for eventIds. No-op when the set is empty. */
export function downloadCalendarInvite(
  group: EventGroupDetail,
  eventIds: ReadonlySet<string>,
  pageUrl: string,
  heading: string = PLAYER_GAMES_HEADING,
): void {
  const ics = buildCalendarInvite(group, eventIds, pageUrl, new Date(), undefined, heading);
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
