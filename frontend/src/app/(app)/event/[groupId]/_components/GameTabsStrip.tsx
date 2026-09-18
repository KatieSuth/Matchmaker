"use client";

// "Games in this group" chip row: switches the single active game tab, or scrolls to a game's
// section when "Show all" is toggled on.
import { EventGroupEvent } from "@/app/_types/types";
import { eventHasUnfairLobby } from "../_lib/lobbyFairness";
import { formatGameModeAndTime } from "../_lib/formatters";

export function GameTabsStrip({
  events,
  activeEventId,
  showAllEvents,
  registrationEditorOpen,
  onToggleShowAll,
  onSelectEvent,
  onScrollToEvent,
}: {
  events: EventGroupEvent[];
  activeEventId: string | null;
  showAllEvents: boolean;
  registrationEditorOpen: boolean;
  onToggleShowAll: () => void;
  onSelectEvent: (eventId: string) => void;
  onScrollToEvent: (eventId: string) => void;
}) {
  return (
    <div className="card rounded-xl p-3 sm:p-4 flex flex-col gap-3">
      <div className="flex items-center justify-between gap-2">
        <p className="text-sm font-medium text-[var(--color-text-soft)]">Games in this group</p>
        {events.length > 1 && (
          <button
            type="button"
            disabled={registrationEditorOpen}
            onClick={onToggleShowAll}
            className="text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text-soft)] underline underline-offset-2 disabled:opacity-40 disabled:cursor-not-allowed disabled:no-underline shrink-0"
          >
            {showAllEvents ? "Show one at a time" : "View all"}
          </button>
        )}
      </div>
      <div className="flex gap-2 overflow-x-auto pb-1">
        {events.map((event, index) => (
          <button
            key={event.id}
            type="button"
            onClick={() => {
              if (showAllEvents) {
                onScrollToEvent(event.id);
                return;
              }
              onSelectEvent(event.id);
            }}
            className={[
              "shrink-0 rounded-full border px-3 py-1.5 text-xs font-medium transition-colors",
              !showAllEvents && activeEventId === event.id
                ? "border-[var(--color-accent-blue)]/35 bg-[var(--color-accent-blue)]/10 text-[var(--color-accent-blue)]"
                : "border-white/10 bg-white/[0.02] text-[var(--color-text-muted)] hover:text-[var(--color-text-soft)]",
            ].join(" ")}
          >
            <span className="inline-flex items-center gap-1">
              Game {index + 1} · {formatGameModeAndTime(event.game_mode_name, event.start_time)}
              {eventHasUnfairLobby(event) && (
                <span className="text-amber-400" aria-label="Contains unfair lobby">
                  ⚠
                </span>
              )}
            </span>
          </button>
        ))}
      </div>
    </div>
  );
}
