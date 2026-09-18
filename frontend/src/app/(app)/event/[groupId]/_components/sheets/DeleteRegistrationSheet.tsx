"use client";

import { ResponsiveSheet } from "@/app/_components/ResponsiveSheet";
import { PendingDeleteAction } from "../../_types";

export function DeleteRegistrationSheet({
  isOpen,
  onClose,
  pendingDeleteAction,
  deletingSelf,
  working,
  onConfirm,
}: {
  isOpen: boolean;
  onClose: () => void;
  pendingDeleteAction: PendingDeleteAction | null;
  deletingSelf: boolean;
  working: boolean;
  onConfirm: () => void;
}) {
  return (
    <ResponsiveSheet
      isOpen={isOpen}
      onClose={onClose}
      title={
        pendingDeleteAction?.mode === "all"
          ? `Delete All Registrations From ${pendingDeleteAction.userName}`
          : `Delete Registration for Game ${pendingDeleteAction?.gameNumber ?? 1}`
      }
    >
      <div className="flex flex-col gap-4">
        <p className="text-sm text-[var(--color-text-soft)]">
          {pendingDeleteAction?.mode === "all"
            ? deletingSelf
              ? "This will delete your registrations for all games in this event series. This action cannot be undone, and you will need to register again if you want to play."
              : `This will delete ${pendingDeleteAction.userName}'s registrations for all games in this event series. This action cannot be undone, and they will need to register again if they want to play.`
            : pendingDeleteAction && pendingDeleteAction.registrationsInGroup > 1
              ? deletingSelf
                ? `You are registered for other games in this series. This action only deletes your registration for Game ${pendingDeleteAction.gameNumber}. It cannot be undone, and you will need to register again to play this game.`
                : `${pendingDeleteAction.userName} is registered for other games in this series. This action only deletes their registration for Game ${pendingDeleteAction.gameNumber}. It cannot be undone, and they will need to register again to play this game.`
              : deletingSelf
                ? `This will delete your registration for Game ${pendingDeleteAction?.gameNumber ?? 1}. This action cannot be undone, and you will need to register again to play this game.`
                : `This will delete the registration for Game ${pendingDeleteAction?.gameNumber ?? 1}. This action cannot be undone, and they will need to register again to play this game.`}
        </p>
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
            onClick={onConfirm}
            disabled={working || !pendingDeleteAction}
            className="rounded-lg border border-[var(--color-text-danger)]/40 bg-[var(--color-text-danger)]/10 px-3 py-2 text-sm text-[var(--color-text-danger)] disabled:opacity-40"
          >
            {working
              ? "Deleting..."
              : pendingDeleteAction?.mode === "all"
                ? "Delete All Registrations"
                : "Delete Registration"}
          </button>
        </div>
      </div>
    </ResponsiveSheet>
  );
}
