"use client";

import { ResponsiveSheet } from "@/app/_components/ResponsiveSheet";
import { Select, SelectOption } from "@/app/_components/Select";
import { PlayerPlacement } from "../../_types";

export function MoveToSubsSheet({
  isOpen,
  onClose,
  pendingMoveToSubs,
  moveToSubsLobbyId,
  onChangeLobby,
  moveToSubsLobbyOptions,
  moveToSubsError,
  working,
  onSubmit,
}: {
  isOpen: boolean;
  onClose: () => void;
  pendingMoveToSubs: PlayerPlacement | null;
  moveToSubsLobbyId: string;
  onChangeLobby: (value: string) => void;
  moveToSubsLobbyOptions: SelectOption[];
  moveToSubsError: string | null;
  working: boolean;
  onSubmit: () => void;
}) {
  return (
    <ResponsiveSheet
      isOpen={isOpen}
      onClose={onClose}
      title={pendingMoveToSubs ? `Move ${pendingMoveToSubs.discordName} to subs` : "Move to subs"}
    >
      <div className="flex flex-col gap-4">
        <p className="text-sm text-[var(--color-text-soft)]">
          Choose which lobby sub pool this player should join.
        </p>
        <Select
          inputId="move-to-subs-lobby"
          ariaLabel="Lobby sub pool"
          value={moveToSubsLobbyId}
          onChange={onChangeLobby}
          options={moveToSubsLobbyOptions}
          placeholder={moveToSubsLobbyOptions.length === 0 ? "No lobbies" : "— Select lobby —"}
          disabled={working || moveToSubsLobbyOptions.length === 0}
        />
        {moveToSubsError && (
          <p className="rounded-lg border border-[var(--color-text-danger)]/30 bg-[var(--color-text-danger)]/10 px-3 py-2 text-sm text-[var(--color-text-danger)]">
            {moveToSubsError}
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
            disabled={working || !pendingMoveToSubs || !moveToSubsLobbyId}
            className="rounded-lg border border-[var(--color-accent-blue)]/40 bg-[var(--color-accent-blue)]/10 px-3 py-2 text-sm text-[var(--color-accent-blue)] hover:bg-[var(--color-accent-blue)]/20 transition-colors disabled:opacity-40"
          >
            {working ? "Moving..." : "Submit"}
          </button>
        </div>
      </div>
    </ResponsiveSheet>
  );
}
