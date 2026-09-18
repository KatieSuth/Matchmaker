"use client";

import { ResponsiveSheet } from "@/app/_components/ResponsiveSheet";
import { Select, SelectOption } from "@/app/_components/Select";
import { PlayerPlacement } from "../../_types";

export function SwapPlayerSheet({
  isOpen,
  onClose,
  pendingSwap,
  swapTargetUserId,
  onChangeSwapTarget,
  swapCandidateOptions,
  swapError,
  working,
  onSubmit,
}: {
  isOpen: boolean;
  onClose: () => void;
  pendingSwap: PlayerPlacement | null;
  swapTargetUserId: string;
  onChangeSwapTarget: (value: string) => void;
  swapCandidateOptions: SelectOption[];
  swapError: string | null;
  working: boolean;
  onSubmit: () => void;
}) {
  return (
    <ResponsiveSheet isOpen={isOpen} onClose={onClose} title={pendingSwap ? `Swap ${pendingSwap.discordName}` : "Swap player"}>
      <div className="flex flex-col gap-4">
        <p className="text-sm text-[var(--color-text-soft)]">
          Choose a player from another team, lobby, the substitutes list, or unplaced to swap with.
          Swapping an unplaced player onto a roster seat moves that roster player to substitutes when they can sub, otherwise they become unplaced.
        </p>
        <Select
          inputId="swap-player-target"
          ariaLabel="Player to swap with"
          value={swapTargetUserId}
          onChange={onChangeSwapTarget}
          options={swapCandidateOptions}
          placeholder={swapCandidateOptions.length === 0 ? "No swap candidates" : "— Select player —"}
          disabled={working || swapCandidateOptions.length === 0}
        />
        {swapError && (
          <p className="rounded-lg border border-[var(--color-text-danger)]/30 bg-[var(--color-text-danger)]/10 px-3 py-2 text-sm text-[var(--color-text-danger)]">
            {swapError}
          </p>
        )}
        <div className="flex justify-end gap-2">
          <button
            type="button"
            onClick={onClose}
            className="rounded-lg border border-white/10 bg-white/[0.03] px-3 py-2 text-sm text-[var(--color-text-soft)]"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={onSubmit}
            disabled={working || !pendingSwap || !swapTargetUserId}
            className="rounded-lg border border-[var(--color-accent-blue)]/40 bg-[var(--color-accent-blue)]/10 px-3 py-2 text-sm text-[var(--color-accent-blue)] hover:bg-[var(--color-accent-blue)]/20 transition-colors disabled:opacity-40"
          >
            {working ? "Swapping..." : "Submit"}
          </button>
        </div>
      </div>
    </ResponsiveSheet>
  );
}
