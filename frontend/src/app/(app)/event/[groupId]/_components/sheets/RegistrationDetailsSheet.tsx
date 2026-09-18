"use client";

import { LobbyHostInfoHint } from "@/app/_components/LobbyHostInfoHint";
import { ResponsiveSheet } from "@/app/_components/ResponsiveSheet";
import { EMPTY_VALUE } from "@/app/_lib/constants";
import { EventRegistration } from "@/app/_types/types";
import { formatDateTime } from "../../_lib/formatters";

export function RegistrationDetailsSheet({
  isOpen,
  onClose,
  registration,
}: {
  isOpen: boolean;
  onClose: () => void;
  registration: EventRegistration | null;
}) {
  return (
    <ResponsiveSheet isOpen={isOpen} onClose={onClose} title="Registration Details">
      {registration ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-sm">
          <div>
            <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Discord</p>
            <p className="text-[var(--color-text-soft)]">{registration.discord_name || EMPTY_VALUE}</p>
          </div>
          <div>
            <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Display Name</p>
            <p className="text-[var(--color-text-soft)]">
              {registration.display_name?.trim() || EMPTY_VALUE}
            </p>
          </div>
          <div>
            <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">In-game name</p>
            <p className="text-[var(--color-text-soft)]">{registration.in_game_name?.trim() || EMPTY_VALUE}</p>
          </div>
          <div>
            <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Pronouns</p>
            <p className="text-[var(--color-text-soft)]">{registration.pronouns || EMPTY_VALUE}</p>
          </div>
          <div>
            <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Current rank</p>
            <p className="text-[var(--color-text-soft)]">{registration.current_rank_name || EMPTY_VALUE}</p>
          </div>
          <div>
            <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Average rank</p>
            <p className="text-[var(--color-text-soft)]">{registration.avg_rank_name || EMPTY_VALUE}</p>
          </div>
          <div>
            <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Peak rank</p>
            <p className="text-[var(--color-text-soft)]">{registration.peak_rank_name || EMPTY_VALUE}</p>
          </div>
          <div>
            <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Sign up time</p>
            <p className="text-[var(--color-text-soft)]">{formatDateTime(registration.created_at)}</p>
          </div>
          <div>
            <div className="flex items-center gap-1.5">
              <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Can lobby host</p>
              <LobbyHostInfoHint />
            </div>
            <p className="text-[var(--color-text-soft)]">{registration.can_lobby_host ? "Yes" : "No"}</p>
          </div>
          <div>
            <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Can substitute</p>
            <p className="text-[var(--color-text-soft)]">{registration.can_substitute ? "Yes" : "No"}</p>
          </div>
          <div>
            <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Duo request</p>
            <p className="text-[var(--color-text-soft)]">{registration.duo_request || EMPTY_VALUE}</p>
          </div>
        </div>
      ) : (
        <p className="text-sm text-[var(--color-text-muted)]">No registration selected.</p>
      )}
    </ResponsiveSheet>
  );
}
