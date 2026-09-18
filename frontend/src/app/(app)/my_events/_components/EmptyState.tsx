"use client";

// Empty-list placeholder, worded differently depending on active tab/time filter/whether filters are applied.
import { Tab, TimeFilter } from "../_types";

export function EmptyState({
  tab,
  time,
  hasFilters,
}: {
  tab: Tab;
  time: TimeFilter;
  hasFilters: boolean;
}) {
  return (
    <div className="flex flex-col items-center justify-center py-14 rounded-xl border border-dashed border-white/[0.08] text-center gap-3">
      <div className="w-10 h-10 rounded-full bg-white/[0.04] border border-white/[0.08] flex items-center justify-center">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" className="text-[var(--color-text-faint)]">
          <rect x="3" y="4" width="18" height="18" rx="3" stroke="currentColor" strokeWidth="1.5" />
          <path d="M3 10h18M8 2v4M16 2v4" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
        </svg>
      </div>
      <div>
        <p className="text-sm font-medium text-[var(--color-text-soft)]">
          {hasFilters
            ? "No events match your filters"
            : time === "past"
            ? "No past events"
            : tab === "hosting"
            ? "You're not hosting any upcoming events"
            : "You're not registered for any upcoming events"}
        </p>
      </div>
    </div>
  );
}
