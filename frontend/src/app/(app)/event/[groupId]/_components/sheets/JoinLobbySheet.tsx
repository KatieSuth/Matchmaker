"use client";

import { ResponsiveSheet } from "@/app/_components/ResponsiveSheet";
import { CopyStatusTooltip } from "@/app/_components/CopyStatusTooltip";
import { CopyStatus } from "@/app/_hooks/useCopyStatus";
import { validateLobbyJoinInput, LobbyJoinDisplayValue } from "@/app/_lib/lobbyJoin";
import { inputCls } from "@/app/_lib/styles";
import { formatDateTime } from "../../_lib/formatters";
import { PendingJoinLobby } from "../../_types";

function CopyIconButton({ onClick, status }: { onClick: () => void; status: CopyStatus }) {
  return (
    <div className="relative shrink-0">
      <button
        type="button"
        onClick={onClick}
        className={[
          "inline-flex h-9 w-9 items-center justify-center rounded-lg border bg-white/[0.03] hover:bg-white/[0.08]",
          status === "success"
            ? "border-emerald-500/35 text-emerald-400"
            : status === "error"
              ? "border-[var(--color-text-danger)]/35 text-[var(--color-text-danger)]"
              : "border-white/10 text-[var(--color-text-soft)]",
        ].join(" ")}
        aria-label="Copy lobby join info"
      >
        <svg width="16" height="16" viewBox="0 0 16 16" fill="none" aria-hidden="true">
          <rect x="5.5" y="5.5" width="7" height="8" rx="1.2" stroke="currentColor" strokeWidth="1.4" />
          <path d="M3.5 10.5V3.8A1.3 1.3 0 0 1 4.8 2.5h5.7" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" />
        </svg>
      </button>
      <CopyStatusTooltip status={status} shadow={false} />
    </div>
  );
}

export function JoinLobbySheet({
  isOpen,
  onClose,
  pendingJoinLobby,
  canEdit,
  joinLobbyDraft,
  onChangeDraft,
  joinLobbyError,
  copyStatus,
  onCopy,
  onSave,
  pendingJoinDisplay,
  joinLinkBase,
  working,
}: {
  isOpen: boolean;
  onClose: () => void;
  pendingJoinLobby: PendingJoinLobby | null;
  canEdit: boolean;
  joinLobbyDraft: string;
  onChangeDraft: (value: string) => void;
  joinLobbyError: string | null;
  copyStatus: CopyStatus;
  onCopy: () => void;
  onSave: () => void;
  pendingJoinDisplay: LobbyJoinDisplayValue | null;
  joinLinkBase: string | null | undefined;
  working: boolean;
}) {
  return (
    <ResponsiveSheet
      isOpen={isOpen}
      onClose={onClose}
      title={
        pendingJoinLobby
          ? `Join Lobby ${pendingJoinLobby.lobbyIndex + 1} (Game ${pendingJoinLobby.gameNumber} · ${formatDateTime(pendingJoinLobby.startTime)})`
          : "Join Lobby"
      }
    >
      <div className="flex flex-col gap-4">
        {canEdit ? (
          <>
            <label className="flex flex-col gap-1.5">
              <span className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">
                Lobby
              </span>
              <div className="flex items-center gap-2">
                <input
                  type="text"
                  value={joinLobbyDraft}
                  onChange={(e) => onChangeDraft(e.target.value)}
                  className={inputCls}
                  placeholder="Code or https://gg.riotgames.com/…"
                  disabled={working}
                />
                {(joinLobbyDraft.trim() || pendingJoinDisplay) && (
                  <CopyIconButton onClick={onCopy} status={copyStatus} />
                )}
              </div>
            </label>
            {joinLobbyError && (
              <p className="rounded-lg border border-[var(--color-text-danger)]/30 bg-[var(--color-text-danger)]/10 px-3 py-2 text-sm text-[var(--color-text-danger)]">
                {joinLobbyError}
              </p>
            )}
            <div className="flex justify-end gap-2">
              <button
                type="button"
                onClick={onClose}
                className="rounded-lg border border-white/10 bg-white/[0.03] px-3 py-2 text-sm text-[var(--color-text-soft)]"
              >
                Close
              </button>
              <button
                type="button"
                onClick={onSave}
                disabled={working || !!validateLobbyJoinInput(joinLobbyDraft, joinLinkBase ?? null)}
                className="rounded-lg border border-[var(--color-accent-blue)]/40 bg-[var(--color-accent-blue)]/10 px-3 py-2 text-sm text-[var(--color-accent-blue)] hover:bg-[var(--color-accent-blue)]/20 transition-colors disabled:opacity-40"
              >
                {working ? "Saving..." : "Save"}
              </button>
            </div>
          </>
        ) : pendingJoinDisplay ? (
          <>
            <div className="flex items-start gap-2">
              {pendingJoinDisplay.kind === "link" ? (
                <a
                  href={pendingJoinDisplay.value}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="min-w-0 flex-1 break-all text-sm text-[var(--color-accent-blue)] hover:underline"
                >
                  {pendingJoinDisplay.value}
                </a>
              ) : (
                <p className="min-w-0 flex-1 break-all font-mono text-sm text-[var(--color-text-soft)]">
                  {pendingJoinDisplay.value}
                </p>
              )}
              <CopyIconButton onClick={onCopy} status={copyStatus} />
            </div>
            <div className="flex justify-end">
              <button
                type="button"
                onClick={onClose}
                className="rounded-lg border border-white/10 bg-white/[0.03] px-3 py-2 text-sm text-[var(--color-text-soft)]"
              >
                Close
              </button>
            </div>
          </>
        ) : (
          <>
            <p className="text-sm text-[var(--color-text-soft)]">
              The info to join the lobby has not been added yet. If it&apos;s less than 10 minutes before the match is due to start, reach out to the
              lobby host or the event host and ask them to add it.
            </p>
            <div className="flex justify-end">
              <button
                type="button"
                onClick={onClose}
                className="rounded-lg border border-white/10 bg-white/[0.03] px-3 py-2 text-sm text-[var(--color-text-soft)]"
              >
                Close
              </button>
            </div>
          </>
        )}
      </div>
    </ResponsiveSheet>
  );
}
