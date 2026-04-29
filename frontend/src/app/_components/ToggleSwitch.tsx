// Pill-shaped switch matching ToggleRow styling; reusable standalone or inside ToggleRow.
interface ToggleSwitchProps {
  checked: boolean;
  onChange: (checked: boolean) => void;
  disabled?: boolean;
  id?: string;
  className?: string;
}

export function ToggleSwitch({ checked, onChange, disabled = false, id, className = "" }: ToggleSwitchProps) {
  return (
    <button
      type="button"
      id={id}
      role="switch"
      aria-checked={checked}
      disabled={disabled}
      onClick={() => {
        if (!disabled) onChange(!checked);
      }}
      className={[
        "flex-shrink-0 w-11 h-6 rounded-full border transition-all duration-200",
        "flex items-center px-1",
        disabled ? "opacity-40 cursor-not-allowed" : "cursor-pointer",
        checked
          ? "bg-[var(--color-accent-blue)]/25 border-[var(--color-accent-blue)]/50"
          : "bg-white/5 border-white/10",
        className,
      ].join(" ")}
    >
      <span
        className={[
          "w-4 h-4 rounded-full transition-all duration-200 shadow-sm flex-shrink-0",
          checked ? "translate-x-[1.125rem] bg-[var(--color-accent-blue)]" : "translate-x-0 bg-white/30",
        ].join(" ")}
      />
    </button>
  );
}
