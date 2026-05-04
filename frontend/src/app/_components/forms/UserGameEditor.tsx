"use client";

import { Field } from "@/app/_components/Field";
import { Select } from "@/app/_components/Select";
import { ToggleRow } from "@/app/_components/ToggleRow";
import { inputCls } from "@/app/_lib/styles";
import { Game, GameRank } from "@/app/_types/types";

export interface UserGameEditorValue {
  game_id: string;
  in_game_name: string;
  current_rank: string;
  peak_rank: string;
  show_rank: boolean;
}

export interface UserGameEditorErrors {
  game_id?: string;
  in_game_name?: string;
  current_rank?: string;
  peak_rank?: string;
}

interface UserGameEditorProps {
  value: UserGameEditorValue;
  errors?: UserGameEditorErrors;
  allGames?: Game[];
  takenGameIds?: string[];
  ranks: GameRank[];
  ranksLoading?: boolean;
  hideGameSelector?: boolean;
  gameLabel?: string;
  /** When true, the game dropdown cannot be changed (e.g. already saved on profile). Name/ranks stay editable. */
  lockGameSelect?: boolean;
  onChange: (next: UserGameEditorValue) => void;
  onRemove?: () => void;
}

export function UserGameEditor({
  value,
  errors,
  allGames = [],
  takenGameIds = [],
  ranks,
  ranksLoading = false,
  hideGameSelector = false,
  gameLabel,
  lockGameSelect = false,
  onChange,
  onRemove,
}: UserGameEditorProps) {
  return (
    <div className="card rounded-xl p-4 flex flex-col gap-4 relative overflow-hidden">
      <div className="absolute top-0 left-4 right-4 h-px bg-top-edge opacity-20 rounded-full" />

      {!hideGameSelector ? (
        <div className="flex items-center gap-2">
          <div className="flex-1">
            <Select
              value={value.game_id}
              onChange={(gameId) =>
                onChange({
                  ...value,
                  game_id: gameId,
                  current_rank: "",
                  peak_rank: "",
                })
              }
              placeholder="— Select a game —"
              disabled={lockGameSelect}
              options={allGames
                .filter((game) => game.id === value.game_id || !takenGameIds.includes(game.id))
                .map((game) => ({ value: game.id, label: game.name }))}
            />
          </div>
          {onRemove && (
            <button
              type="button"
              onClick={onRemove}
              aria-label="Remove game"
              className="flex-shrink-0 size-9 rounded-lg inline-flex items-center justify-center border border-white/10 bg-white/5 text-[var(--color-text-muted)] hover:border-[var(--color-text-danger)]/40 hover:text-[var(--color-text-danger)] hover:bg-[rgba(255,100,80,0.08)] transition-all duration-150 p-0"
            >
              <span className="block text-lg font-semibold leading-none tracking-tight select-none" aria-hidden>
                ×
              </span>
            </button>
          )}
        </div>
      ) : (
        <div className="text-sm text-[var(--color-text-soft)]">
          <span className="text-[var(--color-text-faint)]">Game:</span> {gameLabel ?? "Selected game"}
        </div>
      )}

      {errors?.game_id && <p className="text-xs text-[var(--color-text-danger)] -mt-2">{errors.game_id}</p>}

      {!value.game_id && (
        <div className="flex items-center gap-2.5 px-1 py-0.5">
          <div className="w-1 h-1 rounded-full bg-white/15 flex-shrink-0" />
          <p className="text-xs text-[var(--color-text-faint)]">Select a game above to continue</p>
        </div>
      )}

      {value.game_id && ranksLoading && (
        <div className="flex flex-col gap-3 animate-pulse">
          <div className="h-9 rounded-lg bg-white/5 border border-white/[0.06]" />
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div className="h-9 rounded-lg bg-white/5 border border-white/[0.06]" />
            <div className="h-9 rounded-lg bg-white/5 border border-white/[0.06]" />
          </div>
        </div>
      )}

      {value.game_id && !ranksLoading && (
        <>
          <div className="flex flex-col gap-3">
            <Field label="In-game name *" error={errors?.in_game_name}>
              <input
                value={value.in_game_name}
                onChange={(event) => onChange({ ...value, in_game_name: event.target.value })}
                placeholder="YourTag#1234"
                className={inputCls}
              />
            </Field>

            {ranks.length > 0 && (
              <>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <Field label="Current rank *" error={errors?.current_rank}>
                    <Select
                      value={value.current_rank}
                      onChange={(rankId) => onChange({ ...value, current_rank: rankId })}
                      placeholder="— Select rank —"
                      options={ranks.map((rank) => ({ value: rank.id, label: rank.name }))}
                    />
                  </Field>

                  <Field label="Peak rank *" error={errors?.peak_rank}>
                    <Select
                      value={value.peak_rank}
                      onChange={(rankId) => onChange({ ...value, peak_rank: rankId })}
                      placeholder="— Select rank —"
                      options={ranks.map((rank) => ({ value: rank.id, label: rank.name }))}
                    />
                  </Field>
                </div>

                <p className="text-xs text-[var(--color-text-muted)] leading-relaxed">
                  If unranked, provide the most recent rank you had or one you believe you would be placed into if you were to play.
                </p>
              </>
            )}
          </div>

          <div className="flex flex-col gap-3 pt-3 border-t border-white/[0.06]">
            <ToggleRow
              label="Show rank publicly"
              description={
                value.show_rank
                  ? "Visible to everyone in games you register for"
                  : "Visible only to the host of games you register for"
              }
              checked={value.show_rank}
              onChange={(checked) => onChange({ ...value, show_rank: checked })}
            />
          </div>
        </>
      )}
    </div>
  );
}
