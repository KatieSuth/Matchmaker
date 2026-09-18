"use client";

import { ResponsiveSheet } from "@/app/_components/ResponsiveSheet";

export function DeleteTeamsWarningSheet({
  isOpen,
  onClose,
  working,
  onConfirm,
}: {
  isOpen: boolean;
  onClose: () => void;
  working: boolean;
  onConfirm: () => void;
}) {
  return (
    <ResponsiveSheet isOpen={isOpen} onClose={onClose} title="Delete teams">
      <div className="flex flex-col gap-4">
        <p className="text-sm text-[var(--color-text-soft)]">
          All lobbies and teams for this event will be deleted, but registrations will remain. Registration stays closed until you open it from Edit. This action cannot be undone.
        </p>
        <div className="flex justify-end gap-2">
          <button
            type="button"
            onClick={onClose}
            className="rounded-lg border border-white/10 bg-white/[0.03] px-3 py-2 text-sm text-[var(--color-text-soft)]"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={onConfirm}
            disabled={working}
            className="rounded-lg border border-[var(--color-text-danger)]/40 bg-[var(--color-text-danger)]/10 px-3 py-2 text-sm text-[var(--color-text-danger)] disabled:opacity-40"
          >
            {working ? "Deleting..." : "Delete Teams"}
          </button>
        </div>
      </div>
    </ResponsiveSheet>
  );
}
