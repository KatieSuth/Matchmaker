"use client";

// Modal offering the .ics invitation: to a player right after saving a registration, or to the
// host right after creating the event.
import { ResponsiveSheet } from "@/app/_components/ResponsiveSheet";

interface Props {
  isOpen: boolean;
  onClose: () => void;
  onDownload: () => void;
  /** "registered" after a player saves a registration; "created" after the host creates the event. */
  variant?: "registered" | "created";
}

export function CalendarInviteSheet({ isOpen, onClose, onDownload, variant = "registered" }: Props) {
  const title = variant === "created" ? "Your event is created!" : "You're registered!";
  const firstGame = variant === "created" ? "the first game" : "your first game";
  return (
    <ResponsiveSheet isOpen={isOpen} onClose={onClose} title={title}>
      <div className="flex flex-col gap-4">
        <p className="text-sm text-[var(--color-text-soft)]">
          Add this event to your calendar to get reminders 24 hours, 90 minutes, and 10 minutes before {firstGame}.
          You can download it again later from the calendar button at the top of this page.
        </p>
        <div className="flex justify-end gap-2">
          <button
            type="button"
            onClick={onClose}
            className="rounded-lg border border-white/10 bg-white/[0.03] px-3 py-2 text-sm text-[var(--color-text-soft)]"
          >
            Not now
          </button>
          <button
            type="button"
            onClick={onDownload}
            className="rounded-lg border border-[var(--color-accent-blue)]/35 bg-[var(--color-accent-blue)]/10 px-3 py-2 text-sm font-medium text-[var(--color-accent-blue)] hover:bg-[var(--color-accent-blue)]/20"
          >
            Download calendar invite
          </button>
        </div>
      </div>
    </ResponsiveSheet>
  );
}
