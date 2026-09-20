"use client";

// Generic labeled +/- numeric input; used by event scheduling forms (sub minimums, game counts).
import { useId } from "react";
import { inputCls } from "@/app/_lib/styles";

interface NumberStepperProps {
  label: string;
  value: number;
  min: number;
  onChange: (next: number) => void;
  hint?: string;
  disabled?: boolean;
}

export function NumberStepper({ label, value, min, onChange, hint, disabled = false }: NumberStepperProps) {
  const decrement = () => onChange(Math.max(min, value - 1));
  const increment = () => onChange(value + 1);
  // Associates the visible <label> with the <input> (was previously an unassociated <label>, which
  // fails the axe "form elements must have labels" check — a screen reader couldn't tell what the
  // number field was for). `useId` keeps this collision-safe across multiple instances on one page.
  const inputId = useId();

  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={inputId} className="text-xs font-medium tracking-wide text-[var(--color-text-soft)]">
        {label}
      </label>
      <div className="flex items-center gap-2">
        <button
          type="button"
          onClick={decrement}
          disabled={disabled || value <= min}
          className="h-9 w-9 rounded-lg border border-white/10 bg-white/[0.03] text-[var(--color-text-soft)] transition-colors hover:bg-white/[0.08] disabled:opacity-40 disabled:cursor-not-allowed"
          aria-label={`Decrease ${label}`}
        >
          <svg width="12" height="12" viewBox="0 0 12 12" fill="none" className="mx-auto">
            <path d="M2.25 6h7.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
          </svg>
        </button>
        <input
          id={inputId}
          type="number"
          min={min}
          value={value}
          disabled={disabled}
          onChange={(event) => onChange(Math.max(min, Number(event.target.value) || min))}
          // hide-number-spinners lives in globals.css so it wins over inputCls's appearance-none.
          className={`${inputCls} hide-number-spinners text-center`}
        />
        <button
          type="button"
          onClick={increment}
          disabled={disabled}
          className="h-9 w-9 rounded-lg border border-white/10 bg-white/[0.03] text-[var(--color-text-soft)] transition-colors hover:bg-white/[0.08]"
          aria-label={`Increase ${label}`}
        >
          <svg width="12" height="12" viewBox="0 0 12 12" fill="none" className="mx-auto">
            <path
              d="M6 2.25v7.5M2.25 6h7.5"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
            />
          </svg>
        </button>
      </div>
      {hint && <p className="text-xs text-[var(--color-text-faint)]">{hint}</p>}
    </div>
  );
}
