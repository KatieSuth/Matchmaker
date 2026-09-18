// Shared date/time helpers used across event scheduling and event listing pages.

/** Resolves the browser's IANA time zone, falling back to UTC when Intl is unavailable (SSR). */
export function getUserTimeZone(): string {
  return typeof Intl !== "undefined" ? Intl.DateTimeFormat().resolvedOptions().timeZone : "UTC";
}

/**
 * Formats a `Date` as a medium-length date string (e.g. "Sep 4, 2026") in the given time zone.
 *
 * `timeZone` defaults to the browser's local zone via `getUserTimeZone()`. The formatter is built
 * fresh on every call (rather than cached at module scope) so callers — and tests — can pass an
 * explicit `timeZone` to get deterministic output independent of the host machine's local zone.
 */
export function formatDate(date: Date, timeZone: string = getUserTimeZone()): string {
  return new Intl.DateTimeFormat(undefined, { dateStyle: "medium", timeZone }).format(date);
}

/**
 * Formats a `Date` as a medium date + short time string (e.g. "Sep 4, 2026, 3:00 PM") in the given
 * time zone. See `formatDate` for why the formatter isn't cached at module scope.
 */
export function formatDateTime(date: Date, timeZone: string = getUserTimeZone()): string {
  return new Intl.DateTimeFormat(undefined, { dateStyle: "medium", timeStyle: "short", timeZone }).format(date);
}
