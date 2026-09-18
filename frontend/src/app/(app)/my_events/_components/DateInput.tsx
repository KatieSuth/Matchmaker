"use client";

import { forwardRef } from "react";

// Custom react-datepicker trigger button that matches the shared input styling.
// Must forward a ref: DatePicker cloneElements `customInput` and attaches one.
interface DateInputProps {
  value?: string;
  onClick?: () => void;
  placeholder: string;
  /** Survives react-datepicker's cloneElement, which overwrites `placeholder` with placeholderText. */
  emptyLabel?: string;
  isClearable?: boolean;
  onClear?: () => void;
  id?: string;
}

export const DateInput = forwardRef<HTMLButtonElement, DateInputProps>(function DateInput(
  { value, onClick, placeholder, emptyLabel, isClearable, onClear, id },
  ref,
) {
  const displayText = value || emptyLabel || placeholder;
  const clearLabel = emptyLabel || placeholder;

  return (
    <div className="relative flex items-center w-full">
      <button
        ref={ref}
        id={id}
        type="button"
        onClick={onClick}
        aria-label={displayText}
        className={[
          // Re-use inputCls as a base but override padding for the icon
          "h-9 w-full pl-9 pr-3 rounded-lg text-sm text-left",
          "bg-white/5 border transition-all duration-150",
          "focus:outline-none focus:border-[var(--color-accent-blue)] focus:ring-1 focus:ring-[var(--color-accent-blue)]/30",
          value
            ? "border-white/20 text-[var(--color-text)]"
            : "border-white/10 text-[var(--color-text-muted)]",
        ].join(" ")}
      >
        {displayText}
      </button>
      {/* Calendar icon */}
      <svg
        className="absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none"
        width="13"
        height="13"
        viewBox="0 0 16 16"
        fill="none"
        aria-hidden="true"
      >
        <rect x="1" y="3" width="14" height="12" rx="2" stroke="rgba(180,200,235,0.45)" strokeWidth="1.5" />
        <path d="M1 7h14" stroke="rgba(180,200,235,0.45)" strokeWidth="1.5" />
        <path d="M5 1v4M11 1v4" stroke="rgba(180,200,235,0.45)" strokeWidth="1.5" strokeLinecap="round" />
      </svg>
      {/* Clear button */}
      {isClearable && value && (
        <button
          type="button"
          aria-label={`Clear ${clearLabel}`}
          onClick={(e) => { e.stopPropagation(); onClear?.(); }}
          className="absolute right-2.5 top-1/2 -translate-y-1/2 text-[rgba(180,200,235,0.45)] hover:text-[var(--color-text-soft)] transition-colors"
        >
          <svg width="10" height="10" viewBox="0 0 10 10" fill="currentColor">
            <path d="M1 1l8 8M9 1L1 9" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
          </svg>
        </button>
      )}
    </div>
  );
});
