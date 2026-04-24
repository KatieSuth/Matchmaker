"use client";

// Event group detail: metadata, per-game registration panels, host controls (teams, registration),
// and participant registration / profile actions. Large presentational pieces live in local helpers below.
import { useCallback, useEffect, useMemo, useState } from "react";
import { useParams } from "next/navigation";
import { ResponsiveSheet } from "@/app/_components/ResponsiveSheet";
import { EventForm } from "@/app/_components/forms/EventForm";
import { useAuth } from "@/app/_context/AuthContext";
import { inputCls } from "@/app/_lib/styles";
import {
  createTeams,
  deleteRegistration,
  deleteTeamsAndOpenRegistration,
  fetchEventGroup,
  setEventGroupRegistrationOpen,
  upsertMyRegistration,
} from "@/app/_services/events";
import { EventGroupDetail, EventGroupEvent, EventRegistration } from "@/app/_types/types";

const DATE_TIME_FMT = new Intl.DateTimeFormat(undefined, {
  dateStyle: "medium",
  timeStyle: "short",
});

interface RegistrationDraft {
  can_substitute: boolean;
  can_lobby_host: boolean;
  duo_request: string;
}

function formatDateTime(value: string) {
  return DATE_TIME_FMT.format(new Date(value));
}

function ActionButton({
  label,
  onClick,
  tone = "default",
  disabled = false,
}: {
  label: string;
  onClick: () => void;
  tone?: "default" | "danger";
  disabled?: boolean;
}) {
  const toneClass =
    tone === "danger"
      ? "border-[var(--color-text-danger)]/40 bg-[var(--color-text-danger)]/10 text-[var(--color-text-danger)] hover:bg-[var(--color-text-danger)]/20"
      : "border-white/10 bg-white/[0.03] text-[var(--color-text-soft)] hover:bg-white/[0.08]";

  return (
    <button
      type="button"
      disabled={disabled}
      onClick={onClick}
      className={[
        "w-full rounded-lg px-3 py-2 text-left text-sm font-medium transition-colors",
        "disabled:opacity-40 disabled:cursor-not-allowed",
        toneClass,
      ].join(" ")}
    >
      {label}
    </button>
  );
}

function PlayerCard({
  registration,
  isHostView,
  currentUserId,
  onShowDetails,
  onDeleteRegistration,
}: {
  registration: EventRegistration;
  isHostView: boolean;
  currentUserId?: string;
  onShowDetails: (registration: EventRegistration) => void;
  onDeleteRegistration: (registration: EventRegistration) => void;
}) {
  const [menuOpen, setMenuOpen] = useState(false);
  const canDelete = isHostView || registration.user_id === currentUserId;

  return (
    <div className="card rounded-xl p-4 flex flex-col gap-3 relative overflow-hidden">
      <div className="absolute top-0 left-4 right-4 h-px bg-top-edge opacity-20 rounded-full" />
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="text-sm font-semibold text-[var(--color-text)] truncate">
            {registration.discord_name || "Unknown user"}
          </p>
          <p className="text-xs text-[var(--color-text-muted)] mt-0.5 truncate">
            {registration.pronouns || "\u2014"}
          </p>
        </div>
        {isHostView && (
          <div className="relative">
            <button
              type="button"
              onClick={() => setMenuOpen((v) => !v)}
              className="inline-flex h-8 w-8 items-center justify-center rounded-lg border border-white/10 bg-white/[0.03] text-[var(--color-text-muted)] hover:text-[var(--color-text-soft)]"
              aria-label="Registration actions"
            >
              <svg width="12" height="12" viewBox="0 0 12 12" fill="none">
                <circle cx="6" cy="2.5" r="1" fill="currentColor" />
                <circle cx="6" cy="6" r="1" fill="currentColor" />
                <circle cx="6" cy="9.5" r="1" fill="currentColor" />
              </svg>
            </button>
            {menuOpen && (
              <div className="absolute right-0 top-10 z-20 w-48 rounded-lg border border-white/10 bg-[var(--color-bg-soft)] p-1.5 shadow-[0_18px_45px_rgba(0,0,0,0.55)]">
                <button
                  type="button"
                  onClick={() => {
                    setMenuOpen(false);
                    onShowDetails(registration);
                  }}
                  className="w-full rounded-md px-2.5 py-2 text-left text-sm text-[var(--color-text-soft)] hover:bg-white/[0.08]"
                >
                  Show More Details
                </button>
                {canDelete && (
                  <button
                    type="button"
                    onClick={() => {
                      setMenuOpen(false);
                      onDeleteRegistration(registration);
                    }}
                    className="w-full rounded-md px-2.5 py-2 text-left text-sm text-[var(--color-text-danger)] hover:bg-[var(--color-text-danger)]/10"
                  >
                    Delete Registration
                  </button>
                )}
              </div>
            )}
          </div>
        )}
      </div>

      <div className="h-px bg-white/[0.06]" />

      <div className="grid grid-cols-2 gap-3">
        <div>
          <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Rank</p>
          <p className="text-xs text-[var(--color-text-soft)]">{registration.current_rank_name || "\u2014"}</p>
        </div>
        <div>
          <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Signed up</p>
          <p className="text-xs text-[var(--color-text-soft)]">{formatDateTime(registration.created_at)}</p>
        </div>
      </div>
    </div>
  );
}

function EventPanel({
  event,
  isHostView,
  currentUserId,
  onShowDetails,
  onDeleteRegistration,
}: {
  event: EventGroupEvent;
  isHostView: boolean;
  currentUserId?: string;
  onShowDetails: (registration: EventRegistration) => void;
  onDeleteRegistration: (registration: EventRegistration) => void;
}) {
  return (
    <div className="flex flex-col gap-3">
      {event.registrations.length === 0 ? (
        <div className="rounded-xl border border-dashed border-white/[0.08] py-8 text-center text-sm text-[var(--color-text-muted)]">
          No registered players yet.
        </div>
      ) : (
        event.registrations.map((registration) => (
          <PlayerCard
            key={registration.user_id}
            registration={registration}
            isHostView={isHostView}
            currentUserId={currentUserId}
            onShowDetails={onShowDetails}
            onDeleteRegistration={onDeleteRegistration}
          />
        ))
      )}
    </div>
  );
}

// EventGroupPage - loads GET /events/:groupId and drives host vs guest UI, sheets, and mutations.
export default function EventGroupPage() {
  const params = useParams<{ groupId: string }>();
  const groupId = params?.groupId;
  const { user, isAuthenticated, isLoading: authLoading } = useAuth();

  const [group, setGroup] = useState<EventGroupDetail | null>(null);
  const [pageError, setPageError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [working, setWorking] = useState(false);
  const [activeEventId, setActiveEventId] = useState<string | null>(null);
  const [showAllEvents, setShowAllEvents] = useState(false);
  const [hostMenuOpen, setHostMenuOpen] = useState(false);
  const [editSheetOpen, setEditSheetOpen] = useState(false);
  const [registrationSheetOpen, setRegistrationSheetOpen] = useState(false);
  const [detailsSheetOpen, setDetailsSheetOpen] = useState(false);
  const [warningSheetOpen, setWarningSheetOpen] = useState(false);
  const [selectedRegistration, setSelectedRegistration] = useState<EventRegistration | null>(null);
  const [shareStatus, setShareStatus] = useState<"idle" | "success" | "error">("idle");
  const [registrationDraft, setRegistrationDraft] = useState<RegistrationDraft>({
    can_substitute: false,
    can_lobby_host: false,
    duo_request: "",
  });
  const [toast, setToast] = useState<string | null>(null);

  const loadGroup = useCallback(async (signal?: AbortSignal) => {
    if (!groupId) return;
    setLoading(true);
    setPageError(null);
    try {
      const data = await fetchEventGroup(groupId, signal);
      if (signal?.aborted) return;
      setGroup(data);
      if (!activeEventId || !data.events.some((event) => event.id === activeEventId)) {
        setActiveEventId(data.events[0]?.id ?? null);
      }
    } catch (err) {
      const canceled =
        signal?.aborted ||
        (err as { code?: string; name?: string })?.code === "ERR_CANCELED" ||
        (err as { name?: string })?.name === "CanceledError";
      if (canceled) return;
      setPageError("Could not load this event group.");
    } finally {
      if (!signal?.aborted) {
        setLoading(false);
      }
    }
  }, [activeEventId, groupId]);

  useEffect(() => {
    if (authLoading) return;
    if (!isAuthenticated || !groupId) return;
    const ac = new AbortController();
    const timer = window.setTimeout(() => {
      void loadGroup(ac.signal);
    }, 0);
    return () => {
      window.clearTimeout(timer);
      ac.abort();
    };
  }, [authLoading, groupId, isAuthenticated, loadGroup]);

  useEffect(() => {
    if (!toast) return;
    const timer = window.setTimeout(() => setToast(null), 2500);
    return () => window.clearTimeout(timer);
  }, [toast]);

  useEffect(() => {
    if (shareStatus === "idle") return;
    const timer = window.setTimeout(() => setShareStatus("idle"), 1600);
    return () => window.clearTimeout(timer);
  }, [shareStatus]);

  const activeEvent = useMemo(
    () => group?.events.find((event) => event.id === activeEventId) ?? group?.events[0] ?? null,
    [activeEventId, group?.events]
  );

  const firstEventStart = group?.events[0]?.start_time ?? "";
  const isHost = !!(group && user?.id && user.id === group.owner_id);
  const hasAnyLobbies = !!group?.events.some((event) => event.lobbies_count > 0);
  const myRegistration =
    activeEvent?.registrations.find((registration) => registration.user_id === user?.id) ?? null;

  const refreshAndCloseMenus = async () => {
    await loadGroup();
    setHostMenuOpen(false);
    setWarningSheetOpen(false);
  };

  const withHostAction = async (action: () => Promise<void>) => {
    try {
      setWorking(true);
      await action();
      await refreshAndCloseMenus();
    } catch {
      setPageError("Could not complete that action.");
    } finally {
      setWorking(false);
    }
  };

  const handleShare = async () => {
    const shareUrl = typeof window !== "undefined" ? window.location.href : "";
    if (!shareUrl) return;
    try {
      await navigator.clipboard.writeText(shareUrl);
      setShareStatus("success");
    } catch {
      setShareStatus("error");
    }
  };

  const handleOpenRegistrationSheet = () => {
    setRegistrationDraft({
      can_substitute: myRegistration?.can_substitute ?? false,
      can_lobby_host: myRegistration?.can_lobby_host ?? false,
      duo_request: myRegistration?.duo_request ?? "",
    });
    setRegistrationSheetOpen(true);
  };

  const handleSaveRegistration = async () => {
    if (!activeEvent) return;
    try {
      setWorking(true);
      await upsertMyRegistration(activeEvent.id, registrationDraft);
      setRegistrationSheetOpen(false);
      await loadGroup();
    } catch {
      setPageError("Could not save your registration.");
    } finally {
      setWorking(false);
    }
  };

  const handleDeleteRegistration = async (registration: EventRegistration) => {
    if (!activeEvent) return;
    try {
      setWorking(true);
      await deleteRegistration(activeEvent.id, registration.user_id);
      await loadGroup();
    } catch {
      setPageError("Could not delete registration.");
    } finally {
      setWorking(false);
    }
  };

  const hostActionButtons = () => {
    if (!isHost || !group) return null;
    if (group.registration_open) {
      return (
        <>
          <ActionButton
            label="Close registration"
            onClick={() => withHostAction(() => setEventGroupRegistrationOpen(group.id, false))}
            disabled={working}
          />
          <ActionButton
            label="Close registration & Create Teams"
            onClick={() => withHostAction(() => createTeams(group.id))}
            disabled={working}
          />
        </>
      );
    }

    if (!hasAnyLobbies) {
      return (
        <>
          <ActionButton
            label="Open registration"
            onClick={() => withHostAction(() => setEventGroupRegistrationOpen(group.id, true))}
            disabled={working}
          />
          <ActionButton
            label="Create Teams"
            onClick={() => withHostAction(() => createTeams(group.id))}
            disabled={working}
          />
        </>
      );
    }

    return (
      <ActionButton
        label="Delete Teams & Open Registration"
        tone="danger"
        onClick={() => setWarningSheetOpen(true)}
        disabled={working}
      />
    );
  };

  if (authLoading || (isAuthenticated && loading)) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <p className="text-sm text-[var(--color-text-muted)]">Loading event group...</p>
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <p className="text-sm text-[var(--color-text-muted)]">Please sign in to view this page.</p>
      </div>
    );
  }

  if (!group) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <p className="text-sm text-[var(--color-text-danger)]">{pageError ?? "Event group not found."}</p>
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col items-center px-4 py-8">
      <div className="w-full max-w-3xl flex flex-col gap-5" style={{ animation: "var(--animate-rise)" }}>
        <div className="card rounded-xl p-4 sm:p-5 flex flex-col gap-4 relative overflow-visible">
          <div className="absolute top-0 left-4 right-4 h-px bg-top-edge opacity-20 rounded-full" />
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <h1 className="text-xl font-semibold text-[var(--color-text)] tracking-tight">{group.game_name}</h1>
              <p className="text-sm text-[var(--color-text-muted)] mt-1">
                {group.game_mode_name} · {group.region}
              </p>
              <p className="text-xs text-[var(--color-text-faint)] mt-1">
                First event: {firstEventStart ? formatDateTime(firstEventStart) : "Not scheduled"}
              </p>
            </div>
            <div className="flex items-center gap-2">
              <div className="relative">
                <button
                  type="button"
                  onClick={handleShare}
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
                {shareStatus !== "idle" && (
                  <div
                    className={[
                      "pointer-events-none absolute left-1/2 top-full mt-1.5 -translate-x-1/2 whitespace-nowrap rounded-md border px-2 py-1 text-[11px] shadow-[0_10px_24px_rgba(0,0,0,0.45)]",
                      shareStatus === "success"
                        ? "border-emerald-500/30 bg-[var(--color-bg)] text-emerald-300"
                        : "border-[var(--color-text-danger)]/30 bg-[var(--color-bg)] text-[var(--color-text-danger)]",
                    ].join(" ")}
                  >
                    {shareStatus === "success" ? "Link copied" : "Copy failed"}
                  </div>
                )}
              </div>
              {isHost && (
                <button
                  type="button"
                  onClick={() => setEditSheetOpen(true)}
                  className="rounded-lg border border-white/10 bg-white/[0.03] px-3 py-2 text-sm font-medium text-[var(--color-text-soft)] hover:bg-white/[0.08]"
                >
                  Edit
                </button>
              )}
            </div>
          </div>

          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
            <div>
              <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Status</p>
              <p className="text-xs text-[var(--color-text-soft)]">
                {group.registration_open ? "Registration Open" : "Registration Closed"}
              </p>
            </div>
            <div>
              <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Team Size</p>
              <p className="text-xs text-[var(--color-text-soft)]">{group.team_size}</p>
            </div>
            <div>
              <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Host</p>
              <p className="text-xs text-[var(--color-text-soft)]">{isHost ? "You" : group.owner_name}</p>
            </div>
            {isHost && (
              <div className="relative">
                <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Registration</p>
                <button
                  type="button"
                  onClick={() => setHostMenuOpen((v) => !v)}
                  className="mt-1 w-full rounded-lg border border-white/10 bg-white/[0.03] px-3 py-2 text-left text-sm text-[var(--color-text-soft)] hover:bg-white/[0.08]"
                >
                  Manage Registration
                </button>
                {hostMenuOpen && (
                  <div className="absolute right-0 mt-2 z-20 w-72 rounded-lg border border-white/10 bg-[var(--color-bg)] p-2 shadow-[0_18px_45px_rgba(0,0,0,0.7)] flex flex-col gap-2">
                    {hostActionButtons()}
                  </div>
                )}
              </div>
            )}
          </div>
        </div>

        <div className="card rounded-xl p-3 sm:p-4 flex flex-col gap-3">
          <div className="flex items-center justify-between">
            <p className="text-sm font-medium text-[var(--color-text-soft)]">Games in this group</p>
            <button
              type="button"
              onClick={() => setShowAllEvents((v) => !v)}
              className="text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text-soft)] underline underline-offset-2"
            >
              {showAllEvents ? "Show one at a time" : "View all"}
            </button>
          </div>
          <div className="flex gap-2 overflow-x-auto pb-1">
            {group.events.map((event, index) => (
              <button
                key={event.id}
                type="button"
                onClick={() => {
                  setActiveEventId(event.id);
                  setShowAllEvents(false);
                }}
                className={[
                  "shrink-0 rounded-full border px-3 py-1.5 text-xs font-medium transition-colors",
                  activeEvent?.id === event.id
                    ? "border-[var(--color-accent-blue)]/35 bg-[var(--color-accent-blue)]/10 text-[var(--color-accent-blue)]"
                    : "border-white/10 bg-white/[0.02] text-[var(--color-text-muted)] hover:text-[var(--color-text-soft)]",
                ].join(" ")}
              >
                Game {index + 1} · {DATE_TIME_FMT.format(new Date(event.start_time))}
              </button>
            ))}
          </div>
        </div>

        {!showAllEvents && activeEvent && group.registration_open && (
          <div className="flex justify-center">
            <button
              type="button"
              onClick={handleOpenRegistrationSheet}
              className="rounded-lg border border-[var(--color-accent-blue)]/35 bg-[var(--color-accent-blue)]/10 px-5 py-2.5 text-sm font-medium text-[var(--color-accent-blue)] hover:bg-[var(--color-accent-blue)]/20"
            >
              {myRegistration ? "Edit My Registration" : "Register Now"}
            </button>
          </div>
        )}

        {pageError && (
          <p className="rounded-lg border border-[var(--color-text-danger)]/30 bg-[var(--color-text-danger)]/10 px-3 py-2 text-sm text-[var(--color-text-danger)]">
            {pageError}
          </p>
        )}

        {showAllEvents ? (
          <div className="flex flex-col gap-4">
            {group.events.map((event, index) => (
              <div key={event.id} className="flex flex-col gap-3">
                <h2 className="text-sm font-semibold text-[var(--color-text)]">
                  Game {index + 1} · {formatDateTime(event.start_time)}
                </h2>
                {!group.registration_open && event.lobbies_count > 0 ? (
                  <div className="rounded-xl border border-dashed border-white/[0.08] p-4 text-sm text-[var(--color-text-muted)]">
                    Teams display placeholder. Team assignment UI is coming next, but this section already uses the same
                    registration cards for consistency.
                    <div className="mt-3">
                      <EventPanel
                        event={event}
                        isHostView={isHost}
                        currentUserId={user?.id}
                        onShowDetails={(registration) => {
                          setSelectedRegistration(registration);
                          setDetailsSheetOpen(true);
                        }}
                        onDeleteRegistration={handleDeleteRegistration}
                      />
                    </div>
                  </div>
                ) : (
                  <EventPanel
                    event={event}
                    isHostView={isHost}
                    currentUserId={user?.id}
                    onShowDetails={(registration) => {
                      setSelectedRegistration(registration);
                      setDetailsSheetOpen(true);
                    }}
                    onDeleteRegistration={handleDeleteRegistration}
                  />
                )}
              </div>
            ))}
          </div>
        ) : activeEvent ? (
          !group.registration_open && activeEvent.lobbies_count > 0 ? (
            <div className="rounded-xl border border-dashed border-white/[0.08] p-4 text-sm text-[var(--color-text-muted)]">
              Teams display placeholder. Team assignment UI is coming next, but player cards are reused here.
              <div className="mt-3">
                <EventPanel
                  event={activeEvent}
                  isHostView={isHost}
                  currentUserId={user?.id}
                  onShowDetails={(registration) => {
                    setSelectedRegistration(registration);
                    setDetailsSheetOpen(true);
                  }}
                  onDeleteRegistration={handleDeleteRegistration}
                />
              </div>
            </div>
          ) : (
            <EventPanel
              event={activeEvent}
              isHostView={isHost}
              currentUserId={user?.id}
              onShowDetails={(registration) => {
                setSelectedRegistration(registration);
                setDetailsSheetOpen(true);
              }}
              onDeleteRegistration={handleDeleteRegistration}
            />
          )
        ) : null}
      </div>

      <ResponsiveSheet
        isOpen={editSheetOpen}
        onClose={() => setEditSheetOpen(false)}
        title="Edit event settings"
      >
        <EventForm
          mode="edit"
          eventGroupId={group.id}
          initialValues={{
            game_id: group.game_id,
            game_mode_id: group.game_mode_id,
            region: group.region,
            sub_min: group.sub_min,
            games_to_run: group.events.length,
            start_time_local: firstEventStart ? new Date(firstEventStart).toISOString().slice(0, 16) : "",
            registration_open: group.registration_open,
          }}
          onCancel={() => setEditSheetOpen(false)}
          onSubmitted={() => {
            void loadGroup();
            setToast("Event settings updated.");
          }}
        />
      </ResponsiveSheet>

      <ResponsiveSheet
        isOpen={registrationSheetOpen}
        onClose={() => setRegistrationSheetOpen(false)}
        title={myRegistration ? "Edit Registration" : "Register"}
      >
        <div className="flex flex-col gap-4">
          <div className="rounded-lg border border-white/[0.06] bg-white/[0.02] p-3 text-xs text-[var(--color-text-muted)]">
            Registration form placeholder: this currently captures the core fields and submits to the API.
          </div>
          <label className="flex items-center justify-between gap-3">
            <span className="text-sm text-[var(--color-text-soft)]">Can substitute</span>
            <input
              type="checkbox"
              checked={registrationDraft.can_substitute}
              onChange={(event) =>
                setRegistrationDraft((prev) => ({ ...prev, can_substitute: event.target.checked }))
              }
            />
          </label>
          <label className="flex items-center justify-between gap-3">
            <span className="text-sm text-[var(--color-text-soft)]">Can lobby host</span>
            <input
              type="checkbox"
              checked={registrationDraft.can_lobby_host}
              onChange={(event) =>
                setRegistrationDraft((prev) => ({ ...prev, can_lobby_host: event.target.checked }))
              }
            />
          </label>
          <div className="flex flex-col gap-1.5">
            <label className="text-sm text-[var(--color-text-soft)]">Duo request</label>
            <input
              className={inputCls}
              placeholder="Optional Discord name"
              value={registrationDraft.duo_request}
              onChange={(event) =>
                setRegistrationDraft((prev) => ({ ...prev, duo_request: event.target.value }))
              }
            />
          </div>
          <div className="flex justify-end">
            <button
              type="button"
              onClick={handleSaveRegistration}
              disabled={working}
              className="rounded-lg border border-[var(--color-accent-blue)]/35 bg-[var(--color-accent-blue)]/10 px-4 py-2 text-sm font-medium text-[var(--color-accent-blue)] disabled:opacity-40"
            >
              {working ? "Saving..." : "Save Registration"}
            </button>
          </div>
        </div>
      </ResponsiveSheet>

      <ResponsiveSheet
        isOpen={detailsSheetOpen}
        onClose={() => setDetailsSheetOpen(false)}
        title="Registration Details"
      >
        {selectedRegistration ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-sm">
            <div>
              <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Discord</p>
              <p className="text-[var(--color-text-soft)]">{selectedRegistration.discord_name || "\u2014"}</p>
            </div>
            <div>
              <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Pronouns</p>
              <p className="text-[var(--color-text-soft)]">{selectedRegistration.pronouns || "\u2014"}</p>
            </div>
            <div>
              <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Rank</p>
              <p className="text-[var(--color-text-soft)]">{selectedRegistration.current_rank_name || "\u2014"}</p>
            </div>
            <div>
              <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Sign up time</p>
              <p className="text-[var(--color-text-soft)]">{formatDateTime(selectedRegistration.created_at)}</p>
            </div>
            <div>
              <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Can lobby host</p>
              <p className="text-[var(--color-text-soft)]">{selectedRegistration.can_lobby_host ? "Yes" : "No"}</p>
            </div>
            <div>
              <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Can substitute</p>
              <p className="text-[var(--color-text-soft)]">{selectedRegistration.can_substitute ? "Yes" : "No"}</p>
            </div>
            <div className="sm:col-span-2">
              <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Duo request</p>
              <p className="text-[var(--color-text-soft)]">{selectedRegistration.duo_request || "\u2014"}</p>
            </div>
          </div>
        ) : (
          <p className="text-sm text-[var(--color-text-muted)]">No registration selected.</p>
        )}
      </ResponsiveSheet>

      <ResponsiveSheet
        isOpen={warningSheetOpen}
        onClose={() => setWarningSheetOpen(false)}
        title="Delete Teams & Open Registration"
      >
        <div className="flex flex-col gap-4">
          <p className="text-sm text-[var(--color-text-soft)]">
            This action cannot be undone. All generated teams for this event group will be deleted.
          </p>
          <div className="flex justify-end gap-2">
            <button
              type="button"
              onClick={() => setWarningSheetOpen(false)}
              className="rounded-lg border border-white/10 bg-white/[0.03] px-3 py-2 text-sm text-[var(--color-text-soft)]"
            >
              Cancel
            </button>
            <button
              type="button"
              onClick={() => withHostAction(() => deleteTeamsAndOpenRegistration(group.id))}
              disabled={working}
              className="rounded-lg border border-[var(--color-text-danger)]/40 bg-[var(--color-text-danger)]/10 px-3 py-2 text-sm text-[var(--color-text-danger)] disabled:opacity-40"
            >
              {working ? "Deleting..." : "Delete Teams"}
            </button>
          </div>
        </div>
      </ResponsiveSheet>

      {toast && (
        <div className="fixed bottom-4 left-1/2 -translate-x-1/2 rounded-lg border border-white/10 bg-[var(--color-bg-soft)] px-3 py-2 text-xs text-[var(--color-text-soft)] shadow-[0_18px_45px_rgba(0,0,0,0.45)]">
          {toast}
        </div>
      )}
    </div>
  );
}
