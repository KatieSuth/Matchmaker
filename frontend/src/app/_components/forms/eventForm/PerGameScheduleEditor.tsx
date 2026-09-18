"use client";

// Edit-mode "Games in this series" block: per-scheduled-game start time + mode editing.
import { Dispatch, SetStateAction } from "react";
import { Select } from "@/app/_components/Select";
import { GameMode } from "@/app/_types/types";
import { EventFormDateTimePicker } from "./EventFormDateTimePicker";
import { PerGameDraftRow } from "./schema";

export function PerGameScheduleEditor({
  perGameDraft,
  setPerGameDraft,
  modes,
  modesLoading,
  modesError,
  watchedGameId,
  readOnly,
  userTz,
}: {
  perGameDraft: PerGameDraftRow[];
  setPerGameDraft: Dispatch<SetStateAction<PerGameDraftRow[]>>;
  modes: GameMode[];
  modesLoading: boolean;
  modesError: string | null;
  watchedGameId: string;
  readOnly: boolean;
  userTz: string;
}) {
  if (perGameDraft.length === 0) return null;

  return (
    <div className="flex flex-col gap-3 pt-2 border-t border-white/[0.06]">
      <p className="text-xs font-medium tracking-wide text-[var(--color-text-soft)]">
        Games in this series
      </p>
      <p className="text-xs text-[var(--color-text-faint)]">
        {readOnly
          ? `Each row is one scheduled game. Times use your local timezone (${userTz}).`
          : `Each row applies to one scheduled game. Times use your local timezone (${userTz}).`}
      </p>
      <div className="flex flex-col gap-4">
        {perGameDraft.map((row, index) => (
          <div
            key={row.eventId}
            className="rounded-lg border border-white/[0.08] bg-white/[0.02] p-3 flex flex-col gap-2"
          >
            <p className="text-xs font-semibold text-[var(--color-text-soft)]">Game {index + 1}</p>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div className="flex flex-col gap-1.5">
                <label
                  htmlFor={`event-form-game-start-${row.eventId}`}
                  className="text-xs font-medium tracking-wide text-[var(--color-text-soft)]"
                >
                  Start time *
                </label>
                <EventFormDateTimePicker
                  id={`event-form-game-start-${row.eventId}`}
                  value={row.startLocal}
                  onChange={(v) => {
                    setPerGameDraft((prev) =>
                      prev.map((r) => (r.eventId === row.eventId ? { ...r, startLocal: v } : r)),
                    );
                  }}
                  disallowPast={false}
                  disabled={readOnly}
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <label
                  htmlFor={`event-form-game-mode-${row.eventId}`}
                  className="text-xs font-medium tracking-wide text-[var(--color-text-soft)]"
                >
                  Game mode *
                </label>
                <Select
                  inputId={`event-form-game-mode-${row.eventId}`}
                  value={row.modeId}
                  onChange={(nextMode) => {
                    setPerGameDraft((prev) =>
                      prev.map((r) => (r.eventId === row.eventId ? { ...r, modeId: nextMode } : r)),
                    );
                  }}
                  disabled={readOnly || !watchedGameId || modesLoading || !!modesError}
                  placeholder={
                    !watchedGameId
                      ? "Select game first"
                      : modesLoading
                        ? "Loading game modes..."
                        : "Select game mode"
                  }
                  options={modes.map((gameMode) => ({ value: gameMode.id, label: gameMode.name }))}
                />
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
