// Labeled on/off control with optional subtext, for settings-style forms.
interface ToggleRowProps {
  label: string;
  description?: string;
  checked: boolean;
  onChange: (val: boolean) => void;
}

export function ToggleRow({ label, description, checked, onChange }: ToggleRowProps) {
  return (
    <label className="flex items-center justify-between gap-4 cursor-pointer group">
      <div>
        <p className="text-sm text-[var(--color-text-soft)] group-hover:text-[var(--color-text)] transition-colors">
          {label}
        </p>
        {description && (
          <p className="text-xs text-[var(--color-text-muted)] mt-0.5">
            {description}
          </p>
        )}
      </div>
      <button
        type="button"
        role="switch"
        aria-checked={checked}
        onClick={() => onChange(!checked)}
        className={[
          "flex-shrink-0 w-11 h-6 rounded-full border transition-all duration-200",
          "flex items-center px-1",
          checked
            ? "bg-[var(--color-accent-blue)]/25 border-[var(--color-accent-blue)]/50"
            : "bg-white/5 border-white/10",
        ].join(" ")}
      >
        <span
          className={[
            "w-4 h-4 rounded-full transition-all duration-200 shadow-sm flex-shrink-0",
            checked
              ? "translate-x-[1.125rem] bg-[var(--color-accent-blue)]"
              : "translate-x-0 bg-white/30",
          ].join(" ")}
        />
      </button>
    </label>
  );
}
