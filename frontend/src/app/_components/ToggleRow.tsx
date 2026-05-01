// Labeled on/off control with optional subtext, for settings-style forms.
import type { ReactNode } from "react";
import { ToggleSwitch } from "@/app/_components/ToggleSwitch";

interface ToggleRowProps {
  label: string;
  /** Rendered inline after the label (e.g. info icon). */
  labelAccessory?: ReactNode;
  description?: string;
  checked: boolean;
  onChange: (val: boolean) => void;
  disabled?: boolean;
}

export function ToggleRow({
  label,
  labelAccessory,
  description,
  checked,
  onChange,
  disabled = false,
}: ToggleRowProps) {
  return (
    <div
      className={[
        "flex items-start justify-between gap-4",
        disabled ? "opacity-50 cursor-not-allowed" : "cursor-pointer group",
      ].join(" ")}
    >
      <div className="min-w-0 flex-1">
        <div className="flex items-start gap-2">
          <p
            className={[
              "text-sm text-[var(--color-text-soft)] transition-colors",
              disabled ? "" : "group-hover:text-[var(--color-text)]",
            ].join(" ")}
          >
            {label}
          </p>
          {labelAccessory}
        </div>
        {description && (
          <p className="text-xs text-[var(--color-text-muted)] mt-0.5">
            {description}
          </p>
        )}
      </div>
      <div className="shrink-0 pt-0.5">
        <ToggleSwitch checked={checked} onChange={onChange} disabled={disabled} />
      </div>
    </div>
  );
}
