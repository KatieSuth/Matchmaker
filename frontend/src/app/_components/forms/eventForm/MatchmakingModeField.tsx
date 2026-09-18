"use client";

// Radio-card picker for the event's matchmaking sort logic (balanced vs. rank grouping).
interface MatchmakingModeOption {
  value: "balanced" | "ranked";
  title: string;
  description: string;
}

const MATCHMAKING_MODE_OPTIONS: readonly MatchmakingModeOption[] = [
  {
    value: "balanced",
    title: "Balanced",
    description:
      "Puts similar ranks on opposite teams (one rank band per player slot) so each side gets a matching mix. Best for casual games.",
  },
  {
    value: "ranked",
    title: "Rank Grouping",
    description:
      "Keeps players of similar skill in the same lobby. Best for serious practice.",
  },
];

interface MatchmakingModeFieldProps {
  value: "balanced" | "ranked";
  onChange: (next: "balanced" | "ranked") => void;
  disabled?: boolean;
}

export function MatchmakingModeField({ value, onChange, disabled = false }: MatchmakingModeFieldProps) {
  return (
    <div
      className="flex flex-col gap-2"
      role="radiogroup"
      aria-labelledby="matchmaking-mode-label"
    >
      {MATCHMAKING_MODE_OPTIONS.map((opt) => {
        const selected = value === opt.value;
        return (
          <label
            key={opt.value}
            className={`flex gap-3 rounded-lg border p-3 transition-colors ${
              disabled
                ? "cursor-not-allowed opacity-60"
                : "cursor-pointer"
            } ${
              selected
                ? "border-[var(--color-accent-blue)]/50 bg-[var(--color-accent-blue)]/10"
                : disabled
                  ? "border-white/10 bg-white/[0.02]"
                  : "border-white/10 bg-white/[0.02] hover:bg-white/[0.04]"
            }`}
          >
            <input
              type="radio"
              name="sort_logic_form"
              value={opt.value}
              checked={selected}
              disabled={disabled}
              onChange={() => onChange(opt.value)}
              className="mt-1 h-4 w-4 shrink-0 accent-[var(--color-accent-blue)]"
            />
            <div className="min-w-0 flex flex-col gap-1">
              <span className="text-sm font-medium text-[var(--color-text-soft)]">{opt.title}</span>
              <p className="text-xs leading-relaxed text-[var(--color-text-faint)]">{opt.description}</p>
            </div>
          </label>
        );
      })}
    </div>
  );
}
