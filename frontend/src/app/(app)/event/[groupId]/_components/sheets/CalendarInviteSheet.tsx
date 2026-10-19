"use client";

// Modal shown right after a registration is saved: offers the .ics invitation for the viewer's games.
import { ResponsiveSheet } from "@/app/_components/ResponsiveSheet";

interface Props {
  isOpen: boolean;
  onClose: () => void;
  onDownload: () => void;
}

export function CalendarInviteSheet({ isOpen, onClose, onDownload }: Props) {
  return (
    <ResponsiveSheet isOpen={isOpen} onClose={onClose} title="You're registered!">
      <div className="flex flex-col gap-4">
        <p className="text-sm text-[var(--color-text-soft)]">
          Add this event to your calendar to get reminders 24 hours, 90 minutes, and 10 minutes before your first
          game. You can download it again later from the calendar button at the top of this page.
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
