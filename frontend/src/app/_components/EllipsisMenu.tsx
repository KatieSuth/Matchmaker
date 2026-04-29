"use client";

import { useEffect, useRef, useState } from "react";

export interface EllipsisMenuOption {
  label: string;
  onSelect: () => void;
  tone?: "default" | "danger";
  disabled?: boolean;
}

interface EllipsisMenuProps {
  options: EllipsisMenuOption[];
  ariaLabel?: string;
  className?: string;
}

// Reusable vertical-ellipsis dropdown for contextual actions.
export function EllipsisMenu({ options, ariaLabel = "More actions", className = "" }: EllipsisMenuProps) {
  const [open, setOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const visibleOptions = options.filter((option) => !option.disabled);

  useEffect(() => {
    if (!open) return;
    const handlePointerDown = (event: PointerEvent) => {
      if (!containerRef.current?.contains(event.target as Node)) {
        setOpen(false);
      }
    };
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setOpen(false);
      }
    };
    window.addEventListener("pointerdown", handlePointerDown);
    window.addEventListener("keydown", handleKeyDown);
    return () => {
      window.removeEventListener("pointerdown", handlePointerDown);
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, [open]);

  if (visibleOptions.length === 0) return null;

  return (
    <div ref={containerRef} className={["relative", className].join(" ").trim()}>
      <button
        type="button"
        onClick={() => setOpen((prev) => !prev)}
        className="inline-flex h-8 w-8 items-center justify-center rounded-lg border border-white/10 bg-white/[0.03] text-[var(--color-text-muted)] hover:text-[var(--color-text-soft)]"
        aria-label={ariaLabel}
        aria-expanded={open}
        aria-haspopup="menu"
      >
        <svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden>
          <circle cx="6" cy="2.5" r="1" fill="currentColor" />
          <circle cx="6" cy="6" r="1" fill="currentColor" />
          <circle cx="6" cy="9.5" r="1" fill="currentColor" />
        </svg>
      </button>
      {open && (
        <div className="absolute right-0 top-10 z-20 w-48 rounded-lg border border-white/10 bg-[var(--color-bg)] p-2 shadow-[0_18px_45px_rgba(0,0,0,0.7)]">
          {visibleOptions.map((option) => (
            <button
              key={option.label}
              type="button"
              onClick={() => {
                setOpen(false);
                option.onSelect();
              }}
              className={[
                "w-full rounded-md px-2.5 py-2 text-left text-sm",
                option.tone === "danger"
                  ? "text-[var(--color-text-danger)] hover:bg-[var(--color-text-danger)]/10"
                  : "text-[var(--color-text-soft)] hover:bg-white/[0.08]",
              ].join(" ")}
            >
              {option.label}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
