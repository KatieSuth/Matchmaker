"use client";

// Title/meta card at the top of the event group page: name, first-event time, Discord-pings /
// lock-in-teams / share / settings actions, and the status/team-size/host summary grid.
import { CopyStatusTooltip } from "@/app/_components/CopyStatusTooltip";
import { CopyStatus } from "@/app/_hooks/useCopyStatus";
import { EventGroupDetail } from "@/app/_types/types";
import { formatDateTime, formatGroupTeamSizeLabel, formatHostDisplayLabel } from "../_lib/formatters";

export function EventGroupHeaderCard({
  group,
  isHost,
  hasAnyLobbies,
  working,
  firstEventStart,
  pingStatus,
  onCopyDiscordPings,
  shareStatus,
  onShare,
  onOpenEditSheet,
  onLockInClick,
}: {
  group: EventGroupDetail;
  isHost: boolean;
  hasAnyLobbies: boolean;
  working: boolean;
  firstEventStart: string;
  pingStatus: CopyStatus;
  onCopyDiscordPings: () => void;
  shareStatus: CopyStatus;
  onShare: () => void;
  onOpenEditSheet: () => void;
  onLockInClick: () => void;
}) {
  return (
    <div className="card rounded-xl p-4 sm:p-5 flex flex-col gap-4 relative overflow-visible">
      <div className="absolute top-0 left-4 right-4 h-px bg-top-edge opacity-20 rounded-full" />
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h1 className="text-xl font-semibold text-[var(--color-text)] tracking-tight">
            {group.name.trim() ? group.name : group.game_name}
          </h1>
          <p className="text-sm text-[var(--color-text-muted)] mt-1">
            {group.name.trim()
              ? `${group.game_name} · ${group.game_mode_name} · ${group.region}`
              : `${group.game_mode_name} · ${group.region}`}
          </p>
          <p className="text-xs text-[var(--color-text-faint)] mt-1">
            First event: {firstEventStart ? formatDateTime(firstEventStart) : "Not scheduled"}
          </p>
        </div>
        <div className="flex items-center gap-2">
          {isHost && hasAnyLobbies && (
            <div className="relative">
              <button
                type="button"
                onClick={onCopyDiscordPings}
                className={[
                  "rounded-lg border px-3 py-2 text-sm font-medium",
                  pingStatus === "success"
                    ? "border-emerald-500/35 bg-white/[0.03] text-emerald-400"
                    : pingStatus === "error"
                      ? "border-[var(--color-text-danger)]/35 bg-white/[0.03] text-[var(--color-text-danger)]"
                      : "border-white/10 bg-white/[0.03] text-[var(--color-text-soft)] hover:bg-white/[0.08]",
                ].join(" ")}
              >
                Copy Discord Pings
              </button>
              <CopyStatusTooltip status={pingStatus} successLabel="Ping message copied" />
            </div>
          )}
          {isHost && (
            <button
              type="button"
              disabled={working}
              onClick={onLockInClick}
              className={[
                "rounded-lg border px-3 py-2 text-sm font-medium transition-colors",
                "disabled:opacity-40 disabled:cursor-not-allowed",
                !group.registration_open && hasAnyLobbies
                  ? "border-[var(--color-text-danger)]/40 bg-[var(--color-text-danger)]/10 text-[var(--color-text-danger)] hover:bg-[var(--color-text-danger)]/20"
                  : "border-white/10 bg-white/[0.03] text-[var(--color-text-soft)] hover:bg-white/[0.08]",
              ].join(" ")}
            >
              {group.registration_open
                ? "Lock In & Create Teams"
                : hasAnyLobbies
                  ? "Delete teams"
                  : "Create teams"}
            </button>
          )}
          <div className="relative">
            <button
              type="button"
              onClick={onShare}
              className={[
                "inline-flex h-9 w-9 items-center justify-center rounded-lg border bg-white/[0.03] hover:bg-white/[0.08]",
                shareStatus === "success"
                  ? "border-emerald-500/35 text-emerald-400"
                  : shareStatus === "error"
                    ? "border-[var(--color-text-danger)]/35 text-[var(--color-text-danger)]"
                    : "border-white/10 text-[var(--color-text-soft)]",
              ].join(" ")}
              aria-label="Copy share link"
            >
              {shareStatus === "success" ? (
                <svg width="16" height="16" viewBox="0 0 16 16" fill="none" aria-hidden="true">
                  <path
                    d="M3.5 8.2l3 3L12.5 5.2"
                    stroke="currentColor"
                    strokeWidth="1.7"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  />
                </svg>
              ) : (
                <svg width="16" height="16" viewBox="0 0 16 16" fill="none" aria-hidden="true">
                  <circle cx="4" cy="8" r="1.5" stroke="currentColor" strokeWidth="1.4" />
                  <circle cx="12" cy="4" r="1.5" stroke="currentColor" strokeWidth="1.4" />
                  <circle cx="12" cy="12" r="1.5" stroke="currentColor" strokeWidth="1.4" />
                  <path
                    d="M5.4 7.3L10.6 4.7M5.4 8.7l5.2 2.6"
                    stroke="currentColor"
                    strokeWidth="1.4"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  />
                </svg>
              )}
            </button>
            <CopyStatusTooltip status={shareStatus} successLabel="Link copied" />
          </div>
          <button
            type="button"
            onClick={onOpenEditSheet}
            className="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-white/10 bg-white/[0.03] text-[var(--color-text-soft)] hover:bg-white/[0.08]"
            aria-label="Event settings"
          >
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none" aria-hidden="true">
              <path
                d="M6.4 1.6h3.2l.35 1.4a4.8 4.8 0 0 1 1.15.66l1.4-.45 1.6 2.77-1.05 1.02c.07.33.1.67.1 1 0 .33-.03.67-.1 1l1.05 1.02-1.6 2.77-1.4-.45a4.8 4.8 0 0 1-1.15.66l-.35 1.4H6.4l-.35-1.4a4.8 4.8 0 0 1-1.15-.66l-1.4.45L2 10.62l1.05-1.02A4.9 4.9 0 0 1 2.95 8c0-.33.03-.67.1-1L2 5.98 3.6 3.21l1.4.45c.35-.27.73-.5 1.15-.66L6.4 1.6Z"
                stroke="currentColor"
                strokeWidth="1.3"
                strokeLinejoin="round"
              />
              <circle cx="8" cy="8" r="2" stroke="currentColor" strokeWidth="1.3" />
            </svg>
          </button>
        </div>
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
        <div>
          <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Status</p>
          <p className="text-xs text-[var(--color-text-soft)]">
            {group.registration_open ? "Registration Open" : "Registration Closed"}
          </p>
        </div>
        <div>
          <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Team Size</p>
          <p className="text-xs text-[var(--color-text-soft)]">
            {formatGroupTeamSizeLabel(group.events)}
          </p>
        </div>
        <div>
          <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Host</p>
          <p className="text-xs text-[var(--color-text-soft)]">
            {formatHostDisplayLabel(
              isHost,
              group.owner_display_name,
              group.owner_name,
              group.owner_pronouns ?? "",
            )}
          </p>
        </div>
      </div>
    </div>
  );
}
