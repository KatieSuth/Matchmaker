// Shared date formatter for the My Events dashboard (event cards + active filter summary).
// Delegates to the shared `formatDate` helper (rather than caching an `Intl.DateTimeFormat` here)
// so the time zone can be overridden per-call in tests instead of being baked in at import time.
import { formatDate } from "@/app/_lib/dateTime";

/** Formats a `Date` as a medium-length date string in the given (or the user's local) time zone. */
export function formatMyEventsDate(date: Date, timeZone?: string): string {
  return formatDate(date, timeZone);
}
