// Labeled on/off control with optional subtext, for settings-style forms.
import { ToggleSwitch } from "@/app/_components/ToggleSwitch";

interface ToggleRowProps {
  label: string;
  description?: string;
  checked: boolean;
  onChange: (val: boolean) => void;
  disabled?: boolean;
}

export function ToggleRow({ label, description, checked, onChange, disabled = false }: ToggleRowProps) {
  return (
    <div
      className={[
        "flex items-center justify-between gap-4",
        disabled ? "opacity-50 cursor-not-allowed" : "cursor-pointer group",
      ].join(" ")}
    >
      <div>
        <p
          className={[
            "text-sm text-[var(--color-text-soft)] transition-colors",
            disabled ? "" : "group-hover:text-[var(--color-text)]",
          ].join(" ")}
        >
          {label}
        </p>
        {description && (
          <p className="text-xs text-[var(--color-text-muted)] mt-0.5">
            {description}
          </p>
        )}
      </div>
      <ToggleSwitch checked={checked} onChange={onChange} disabled={disabled} />
    </div>
  );
}
