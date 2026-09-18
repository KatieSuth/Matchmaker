"use client";

// Full-screen confirmation overlay for permanently deleting an event group.
export function DeleteEventDialog({
  isOpen,
  deleteError,
  isDeleting,
  onCancel,
  onConfirm,
}: {
  isOpen: boolean;
  deleteError: string | null;
  isDeleting: boolean;
  onCancel: () => void;
  onConfirm: () => void;
}) {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/70 px-4">
      <div className="w-full max-w-md rounded-xl border border-white/10 bg-[var(--color-bg)] p-5 shadow-[0_30px_90px_rgba(0,0,0,0.7)]">
        <h3 className="text-base font-semibold text-[var(--color-text)]">Delete Event</h3>
        <p className="mt-2 text-sm text-[var(--color-text-soft)]">
          This action cannot be undone. All games, registrations, and teams in this event group will be permanently deleted.
        </p>
        {deleteError && <p className="mt-3 text-xs text-[var(--color-text-danger)]">{deleteError}</p>}
        <div className="mt-4 flex justify-end gap-2">
          <button
            type="button"
            onClick={onCancel}
            disabled={isDeleting}
            className="px-3 py-2 rounded-lg text-sm font-medium border border-white/10 bg-white/[0.03] text-[var(--color-text-muted)] hover:text-[var(--color-text-soft)] transition-colors disabled:opacity-50"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={onConfirm}
            disabled={isDeleting}
            className="px-3 py-2 rounded-lg text-sm font-medium border border-[var(--color-text-danger)]/40 bg-[var(--color-text-danger)]/10 text-[var(--color-text-danger)] hover:bg-[var(--color-text-danger)]/20 transition-colors disabled:opacity-50"
          >
            {isDeleting ? "Deleting..." : "Delete Permanently"}
          </button>
        </div>
      </div>
    </div>
  );
}
