// Warning for a viewer who was marked absent at this host's games in the last 6 months, shown only
// when the event deprioritizes no-shows. Yellow while under the event's limit, red once it is met.
import { formatDate } from "@/app/_lib/dateTime";

interface Props {
  /** One date per host event where the viewer was a no-show (earliest missed game), oldest first. */
  noShowDates: string[];
  deprioritizeNoShows: boolean;
  /** No-show count that triggers deprioritization; only used while deprioritizeNoShows is on. */
  maxNoShows: number;
  timeZone?: string;
}

/** "1 time" / "N times". */
function formatTimes(count: number): string {
  return count === 1 ? "1 time" : `${count} times`;
}

/** Joins game dates as "Sep 4, 2026, Sep 11, 2026, and Sep 18, 2026" in the viewer's locale. */
function formatDateList(dates: string[], timeZone?: string): string {
  const formatted = dates.map((iso) => formatDate(new Date(iso), timeZone));
  return new Intl.ListFormat(undefined, { style: "long", type: "conjunction" }).format(formatted);
}

export function NoShowNoticeBanner({ noShowDates, deprioritizeNoShows, maxNoShows, timeZone }: Props) {
  if (!deprioritizeNoShows || noShowDates.length === 0) {
    return null;
  }

  const count = noShowDates.length;
  const excessive = maxNoShows >= 1 && count >= maxNoShows;
  const summary = `You have been marked as a no-show ${formatTimes(count)} at this host's events in the last 6 months: ${formatDateList(noShowDates, timeZone)}.`;

  if (excessive) {
    return (
      <section
        aria-label="No-show warning"
        className="rounded-lg border border-[var(--color-text-danger)]/30 bg-[var(--color-text-danger)]/10 px-4 py-3 text-sm"
      >
        <p className="font-medium text-[var(--color-text-danger)]">{summary}</p>
        <ul className="mt-2 list-disc space-y-1 pl-5 text-[var(--color-text-soft)]">
          <li>
            The rules for this event deprioritize players with {maxNoShows} or more no-shows, which makes it less
            likely (but not impossible) you will be placed on a team. You may still be called on to substitute, so
            register if you are interested in playing.
          </li>
          <li>
            If you can&apos;t make a game you&apos;ve registered for, edit your registration and un-toggle the game (remember
            to save or confirm the delete action!).
          </li>
          <li>
            If you&apos;ve been placed on a team and can&apos;t make it, let the host
            know as soon as possible so a substitute can be found in time.
          </li>
          <li>If you believe an absence was marked by mistake, let the host know.</li>
        </ul>
      </section>
    );
  }

  return (
    <section
      aria-label="No-show notice"
      className="rounded-lg border border-amber-400/30 bg-amber-400/10 px-4 py-3 text-sm"
    >
      <p className="font-medium text-amber-200">{summary}</p>
      <ul className="mt-2 list-disc space-y-1 pl-5 text-[var(--color-text-soft)]">
        <li>
          You are not currently being deprioritized, but if you no-show again you may be deprioritized for future
          games.
        </li>
        <li>
          If you can&apos;t make a game you&apos;ve registered for, edit your registration and un-toggle the game (remember
          to save or confirm the delete action!).
        </li>
        <li>
          If you&apos;ve been placed on a team and can&apos;t make it, let the host
          know as soon as possible so a substitute can be found in time.
        </li>
        <li>If you believe an absence was marked by mistake, let the host know.</li>
      </ul>
    </section>
  );
}
