"use client";

import { ResponsiveSheet } from "@/app/_components/ResponsiveSheet";
import { PendingLobbyHostChange } from "../../_types";

export function LobbyHostConfirmSheet({
  isOpen,
  onClose,
  pendingLobbyHostChange,
  working,
  onConfirm,
}: {
  isOpen: boolean;
  onClose: () => void;
  pendingLobbyHostChange: PendingLobbyHostChange | null;
  working: boolean;
  onConfirm: () => void;
}) {
  return (
    <ResponsiveSheet
      isOpen={isOpen}
      onClose={onClose}
      title={
        pendingLobbyHostChange
          ? `Make ${pendingLobbyHostChange.placement.discordName} Lobby Host`
          : "Make Lobby Host"
      }
    >
      {pendingLobbyHostChange ? (
        <div className="flex flex-col gap-4">
          <p className="text-sm text-[var(--color-text-soft)]">
            {pendingLobbyHostChange.placement.discordName} indicated they do not want to be a lobby host.
          </p>
          {pendingLobbyHostChange.volunteerOptions.length > 0 ? (
            <div className="flex flex-col gap-2">
              <p className="text-sm text-[var(--color-text-soft)]">
                Other players on a team in this lobby who want to host:
              </p>
              <ul className="list-disc pl-5 text-sm text-[var(--color-text-soft)]">
                {pendingLobbyHostChange.volunteerOptions.map((player) => (
                  <li key={player.userId}>
                    {player.discordName} · Team {player.teamNumber}
                    {player.isCurrentHost ? " · (current host)" : ""}
                  </li>
                ))}
              </ul>
            </div>
          ) : (
            <p className="text-sm text-[var(--color-text-soft)]">
              There are no other players on a team in this lobby who want to host.
            </p>
          )}
          <p className="text-sm text-[var(--color-text-soft)]">
            Do you still want to make {pendingLobbyHostChange.placement.discordName} the lobby host?
          </p>
          <div className="flex justify-end gap-2">
            <button
              type="button"
              onClick={onClose}
              className="rounded-lg border border-white/10 bg-white/[0.03] px-3 py-2 text-sm text-[var(--color-text-soft)]"
            >
              No
            </button>
            <button
              type="button"
              onClick={onConfirm}
              disabled={working}
              className="rounded-lg border border-[var(--color-accent-blue)]/40 bg-[var(--color-accent-blue)]/10 px-3 py-2 text-sm text-[var(--color-accent-blue)] hover:bg-[var(--color-accent-blue)]/20 transition-colors disabled:opacity-40"
            >
              {working ? "Updating..." : "Yes, make lobby host"}
            </button>
          </div>
        </div>
      ) : (
        <p className="text-sm text-[var(--color-text-muted)]">No player selected.</p>
      )}
    </ResponsiveSheet>
  );
}
