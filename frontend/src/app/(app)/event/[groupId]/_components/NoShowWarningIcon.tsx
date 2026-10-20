"use client";

// Host-only orange warning shown left of a player's name when their recent no-shows at this
// host meet the deprioritization limit. Hover, keyboard focus, or a tap reveals the count.
import { useId, useState } from "react";

interface Props {
  count: number;
}

/** "1 time" / "N times" for the tooltip copy. */
function formatNoShowTimes(count: number): string {
  return count === 1 ? "1 time" : `${count} times`;
}

export function NoShowWarningIcon({ count }: Props) {
  const [open, setOpen] = useState(false);
  const tooltipId = useId();
  const message = `No-show ${formatNoShowTimes(count)} at your events in the last 6 months`;

  return (
    <span
      className="relative inline-flex shrink-0 items-center"
      onMouseEnter={() => setOpen(true)}
      onMouseLeave={() => setOpen(false)}
    >
      <button
        type="button"
        aria-label={message}
        aria-describedby={open ? tooltipId : undefined}
        onFocus={() => setOpen(true)}
        onBlur={() => setOpen(false)}
        // Touch screens have no hover, and Safari does not focus buttons on tap: a tap opens it,
        // and tapping elsewhere (blur) closes it. Open-only so a mouse click never toggles it shut.
        onClick={() => setOpen(true)}
        className="inline-flex text-orange-400 rounded focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-orange-400/60"
      >
        <svg width="14" height="14" viewBox="0 0 16 16" fill="none" aria-hidden="true">
          <path d="M8 1.75 15 14.25H1L8 1.75Z" stroke="currentColor" strokeWidth="1.5" strokeLinejoin="round" />
          <path d="M8 6.25v3.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
          <circle cx="8" cy="11.9" r="0.85" fill="currentColor" />
        </svg>
      </button>
      {open && (
        <span
          id={tooltipId}
          role="tooltip"
          className="pointer-events-none absolute left-0 top-full z-50 mt-1.5 whitespace-nowrap rounded-md border border-orange-400/30 bg-[var(--color-bg)] px-2 py-1 text-[11px] font-normal text-orange-200 shadow-[0_10px_24px_rgba(0,0,0,0.45)]"
        >
          {message}
        </span>
      )}
    </span>
  );
}
