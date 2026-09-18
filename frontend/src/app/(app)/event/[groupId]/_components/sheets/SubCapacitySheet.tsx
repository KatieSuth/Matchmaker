"use client";

import { ResponsiveSheet } from "@/app/_components/ResponsiveSheet";

export function SubCapacitySheet({ isOpen, onClose }: { isOpen: boolean; onClose: () => void }) {
  return (
    <ResponsiveSheet isOpen={isOpen} onClose={onClose} title="Substitute minimum changed the teams">
      <div className="flex flex-col gap-4">
        <p className="text-sm text-[var(--color-text-soft)]">
          There were not enough players willing to substitute to keep the most even lineup
          while still filling the substitute minimum for each lobby. Some players who would
          have been a closer rank match were placed as substitutes instead. You can still
          swap players if you want a different lineup.
        </p>
        <div className="flex justify-end">
          <button
            type="button"
            onClick={onClose}
            className="rounded-lg border border-[var(--color-accent-blue)]/35 bg-[var(--color-accent-blue)]/10 px-3 py-2 text-sm font-medium text-[var(--color-accent-blue)] hover:bg-[var(--color-accent-blue)]/20"
          >
            Okay
          </button>
        </div>
      </div>
    </ResponsiveSheet>
  );
}
