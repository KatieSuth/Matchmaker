"use client";

// Event group detail: metadata, per-game registration panels, host controls (teams, registration),
// and participant registration / profile actions. Large presentational pieces live in local helpers below.
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useParams } from "next/navigation";
import { EllipsisMenu, EllipsisMenuOption } from "@/app/_components/EllipsisMenu";
import { ResponsiveSheet } from "@/app/_components/ResponsiveSheet";
import { ToggleRow } from "@/app/_components/ToggleRow";
import { ToggleSwitch } from "@/app/_components/ToggleSwitch";
import { EventForm } from "@/app/_components/forms/EventForm";
import { UserGameEditor, UserGameEditorValue } from "@/app/_components/forms/UserGameEditor";
import { useAuth } from "@/app/_context/AuthContext";
import { EMPTY_VALUE } from "@/app/_lib/constants";
import { inputCls } from "@/app/_lib/styles";
import { extractApiError, fetchGameRanks } from "@/app/_services/games";
import {
  createTeams,
  deleteRegistration,
  deleteTeamsAndOpenRegistration,
  fetchEventGroup,
  setEventGroupRegistrationOpen,
  upsertMyGroupRegistrations,
} from "@/app/_services/events";
import { fetchCurrentUserGames, upsertCurrentUserGame } from "@/app/_services/users";
import { EventGroupDetail, EventGroupEvent, EventRegistration, GameRank } from "@/app/_types/types";

const DATE_TIME_FMT = new Intl.DateTimeFormat(undefined, {
  dateStyle: "medium",
  timeStyle: "short",
});
interface EventRegistrationDraft {
  can_substitute: boolean;
  can_lobby_host: boolean;
}

interface RegistrationDraft {
  selected_event_ids: string[];
  per_event: Record<string, EventRegistrationDraft>;
  duo_request: string;
}

interface PendingDeleteAction {
  mode: "single" | "all";
  userId: string;
  userName: string;
  eventId: string;
  gameNumber: number;
  registrationsInGroup: number;
}

function emptyUserGameDraft(gameId: string): UserGameEditorValue {
  return {
    game_id: gameId,
    in_game_name: "",
    current_rank: "",
    peak_rank: "",
    show_rank: false,
  };
}

function formatDateTime(value: string) {
  return DATE_TIME_FMT.format(new Date(value));
}

function formatPlayerCount(count: number) {
  return `${count} ${count === 1 ? "Player" : "Players"}`;
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
  gameNumber,
  eventRegion,
  currentUserRegion,
  isHostView,
  currentUserId,
  canEditRegistration,
  onShowDetails,
  onDeleteRegistrationForGame,
  onDeleteAllFromUser,
}: {
  registration: EventRegistration;
  gameNumber: number;
  eventRegion: string;
  currentUserRegion?: string | null;
  isHostView: boolean;
  currentUserId?: string;
  canEditRegistration: boolean;
  onShowDetails: (registration: EventRegistration) => void;
  onDeleteRegistrationForGame: (registration: EventRegistration, gameNumber: number) => void;
  onDeleteAllFromUser: (registration: EventRegistration, gameNumber: number) => void;
}) {
  const canOpenMenu = isHostView || canEditRegistration;
  const canDelete = isHostView || registration.user_id === currentUserId;
  const regionMismatch =
    canEditRegistration &&
    !!currentUserRegion &&
    currentUserRegion.trim().toUpperCase() !== eventRegion.trim().toUpperCase();
  const menuOptions: EllipsisMenuOption[] = [];
  menuOptions.push({
    label: "Show More Details",
    onSelect: () => onShowDetails(registration),
  });
  if (canDelete) {
    menuOptions.push({
      label: `Delete for Game ${gameNumber}`,
      onSelect: () => onDeleteRegistrationForGame(registration, gameNumber),
      tone: "danger",
    });
    menuOptions.push({
      label: `Delete All`,
      onSelect: () => onDeleteAllFromUser(registration, gameNumber),
      tone: "danger",
    });
  }

  return (
    <div
      className={[
        "card rounded-xl p-4 flex flex-col gap-3 relative overflow-visible",
        regionMismatch ? "ring-1 ring-amber-400/35" : "",
      ].join(" ")}
    >
      <div className="absolute top-0 left-4 right-4 h-px bg-top-edge opacity-20 rounded-full" />
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="text-sm font-semibold text-[var(--color-text)] truncate">
            {registration.discord_name || "Unknown user"}
          </p>
          <p className="text-xs text-[var(--color-text-muted)] mt-0.5 truncate">
            {registration.pronouns || EMPTY_VALUE}
          </p>
          {regionMismatch && (
            <p className="text-[11px] text-amber-300 mt-1">
              Region: {currentUserRegion}
            </p>
          )}
        </div>
        {canOpenMenu && (
          <EllipsisMenu options={menuOptions} ariaLabel="Registration actions" />
        )}
      </div>

      <div className="h-px bg-white/[0.06]" />

      <div className="grid grid-cols-2 gap-3">
        <div>
          <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Current Rank</p>
          <p className="text-xs text-[var(--color-text-soft)]">{registration.current_rank_name || EMPTY_VALUE}</p>
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
  gameNumber,
  eventRegion,
  currentUserRegion,
  isHostView,
  currentUserId,
  onShowDetails,
  onDeleteRegistrationForGame,
  onDeleteAllFromUser,
}: {
  event: EventGroupEvent;
  gameNumber: number;
  eventRegion: string;
  currentUserRegion?: string | null;
  isHostView: boolean;
  currentUserId?: string;
  onShowDetails: (registration: EventRegistration) => void;
  onDeleteRegistrationForGame: (registration: EventRegistration, gameNumber: number) => void;
  onDeleteAllFromUser: (registration: EventRegistration, gameNumber: number) => void;
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
            gameNumber={gameNumber}
            eventRegion={eventRegion}
            currentUserRegion={currentUserRegion}
            isHostView={isHostView}
            currentUserId={currentUserId}
            canEditRegistration={registration.user_id === currentUserId}
            onShowDetails={onShowDetails}
            onDeleteRegistrationForGame={onDeleteRegistrationForGame}
            onDeleteAllFromUser={onDeleteAllFromUser}
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
  const [registrationEditorOpen, setRegistrationEditorOpen] = useState(false);
  const [detailsSheetOpen, setDetailsSheetOpen] = useState(false);
  const [warningSheetOpen, setWarningSheetOpen] = useState(false);
  const [deleteWarningSheetOpen, setDeleteWarningSheetOpen] = useState(false);
  const [selectedRegistration, setSelectedRegistration] = useState<EventRegistration | null>(null);
  const [pendingDeleteAction, setPendingDeleteAction] = useState<PendingDeleteAction | null>(null);
  const [shareStatus, setShareStatus] = useState<"idle" | "success" | "error">("idle");
  const [registrationDraft, setRegistrationDraft] = useState<RegistrationDraft>({
    selected_event_ids: [],
    per_event: {},
    duo_request: "",
  });
  const [userGameDraft, setUserGameDraft] = useState<UserGameEditorValue>({
    game_id: "",
    in_game_name: "",
    current_rank: "",
    peak_rank: "",
    show_rank: false,
  });
  const [userGameRanks, setUserGameRanks] = useState<GameRank[]>([]);
  const [registrationError, setRegistrationError] = useState<string | null>(null);
  const [registrationLoading, setRegistrationLoading] = useState(false);
  const [toast, setToast] = useState<string | null>(null);
  const topAnchorRef = useRef<HTMLDivElement | null>(null);
  const eventSectionRefs = useRef<Record<string, HTMLDivElement | null>>({});

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
  const activeEventNumber = useMemo(() => {
    if (!group || !activeEvent) return 1;
    const idx = group.events.findIndex((event) => event.id === activeEvent.id);
    return idx >= 0 ? idx + 1 : 1;
  }, [activeEvent, group]);
  const scrollToEventSection = useCallback((eventId: string) => {
    eventSectionRefs.current[eventId]?.scrollIntoView({
      behavior: "smooth",
      block: "start",
    });
  }, []);
  const scrollToTop = useCallback(() => {
    topAnchorRef.current?.scrollIntoView({
      behavior: "smooth",
      block: "start",
    });
  }, []);

  const firstEventStart = group?.events[0]?.start_time ?? "";
  const isHost = !!(group && user?.id && user.id === group.owner_id);
  const hasAnyLobbies = !!group?.events.some((event) => event.lobbies_count > 0);
  const myRegistrationsByEvent = useMemo(() => {
    const map = new Map<string, EventRegistration>();
    if (!group || !user?.id) return map;
    for (const event of group.events) {
      const registration = event.registrations.find((item) => item.user_id === user.id);
      if (registration) {
        map.set(event.id, registration);
      }
    }
    return map;
  }, [group, user]);

  const selectedValidEventIds = useMemo(() => {
    if (!group) return [];
    const validIds = new Set(group.events.map((event) => event.id));
    return Array.from(new Set(registrationDraft.selected_event_ids)).filter((eventId) => validIds.has(eventId));
  }, [group, registrationDraft.selected_event_ids]);
  const userGameErrors = useMemo(
    () => ({
      in_game_name: userGameDraft.in_game_name.trim() ? undefined : "In-game name is required",
      current_rank: userGameDraft.current_rank ? undefined : "Current rank is required",
      peak_rank: userGameDraft.peak_rank ? undefined : "Peak rank is required",
    }),
    [userGameDraft.current_rank, userGameDraft.in_game_name, userGameDraft.peak_rank]
  );
  const hasUserGameErrors = !!(userGameErrors.in_game_name || userGameErrors.current_rank || userGameErrors.peak_rank);
  const registrationIsEditMode = myRegistrationsByEvent.size > 0;
  const canDeleteAllViaSave =
    registrationIsEditMode && selectedValidEventIds.length === 0 && !!user?.id && !working && !registrationLoading;
  const canSaveRegistration =
    selectedValidEventIds.length > 0 && !working && !registrationLoading && !hasUserGameErrors;
  const canSubmitRegistration = canSaveRegistration || canDeleteAllViaSave;
  const regionMismatchWarning = (() => {
    if (!group || !user?.region || selectedValidEventIds.length === 0) return null;
    const preferredRegion = user.region.trim();
    const eventRegion = group.region.trim();
    if (!preferredRegion || !eventRegion) return null;
    if (preferredRegion.toUpperCase() === eventRegion.toUpperCase()) return null;
    return `Heads up: your preferred region is ${preferredRegion}, but this event is in ${eventRegion}.`;
  })();

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

  const handleOpenRegistrationSheet = async (registration?: EventRegistration) => {
    if (!group) return;
    setRegistrationError(null);
    setRegistrationLoading(true);
    const selectedEventIds: string[] = [];
    const perEvent: Record<string, EventRegistrationDraft> = {};
    let duoRequest = "";

    for (const event of group.events) {
      const existing = myRegistrationsByEvent.get(event.id);
      if (!existing) continue;
      selectedEventIds.push(event.id);
      perEvent[event.id] = {
        can_substitute: existing.can_substitute,
        can_lobby_host: existing.can_lobby_host,
      };
      if (!duoRequest && existing.duo_request) {
        duoRequest = existing.duo_request;
      }
    }

    if (selectedEventIds.length === 0) {
      for (const event of group.events) {
        selectedEventIds.push(event.id);
        perEvent[event.id] = {
          can_substitute: true,
          can_lobby_host: false,
        };
      }
    }

    if (registration && !selectedEventIds.includes(registration.event_id)) {
      selectedEventIds.push(registration.event_id);
      perEvent[registration.event_id] = {
        can_substitute: registration.can_substitute,
        can_lobby_host: registration.can_lobby_host,
      };
      if (!duoRequest && registration.duo_request) {
        duoRequest = registration.duo_request;
      }
    }

    setRegistrationDraft({
      selected_event_ids: selectedEventIds,
      per_event: perEvent,
      duo_request: duoRequest,
    });
    try {
      const [userGames, ranks] = await Promise.all([
        fetchCurrentUserGames(),
        fetchGameRanks(group.game_id),
      ]);
      const existing = userGames.find((userGame) => userGame.game_id === group.game_id);
      setUserGameRanks(ranks);
      setUserGameDraft(
        existing
          ? {
              game_id: group.game_id,
              in_game_name: existing.in_game_name ?? "",
              current_rank: existing.current_rank ?? "",
              peak_rank: existing.peak_rank ?? "",
              show_rank: existing.show_rank,
            }
          : emptyUserGameDraft(group.game_id)
      );
    } catch (err) {
      setUserGameRanks([]);
      setUserGameDraft(emptyUserGameDraft(group.game_id));
      setRegistrationError(extractApiError(err, "Could not load your game settings. Please try again."));
    } finally {
      setRegistrationLoading(false);
    }
    setRegistrationEditorOpen(true);
  };

  const handleCloseRegistrationEditor = async () => {
    setRegistrationError(null);
    setRegistrationEditorOpen(false);
    await loadGroup();
  };

  const deleteAllRegistrationsForUserInGroup = async (targetUserId: string) => {
    if (!group) return;
    const eventIds = group.events
      .filter((event) => event.registrations.some((item) => item.user_id === targetUserId))
      .map((event) => event.id);
    await Promise.all(eventIds.map((eventId) => deleteRegistration(eventId, targetUserId)));
  };

  const handleSaveRegistration = async () => {
    if (!group) return;
    setRegistrationError(null);
    if (selectedValidEventIds.length === 0) {
      if (registrationIsEditMode && user?.id) {
        openDeleteAllForCurrentUserConfirmation();
        return;
      }
      setRegistrationError("Select at least one event, then try saving again.");
      return;
    }
    if (hasUserGameErrors) {
      setRegistrationError("Complete your game profile (in-game name, current rank, and peak rank) before saving.");
      return;
    }
    setWorking(true);
    try {
      await upsertCurrentUserGame(group.game_id, {
        in_game_name: userGameDraft.in_game_name.trim(),
        current_rank: userGameDraft.current_rank,
        peak_rank: userGameDraft.peak_rank,
        show_rank: userGameDraft.show_rank,
      });
    } catch (err) {
      setRegistrationError(
        `${extractApiError(err)}. Complete your game profile, then try again.`
      );
      setWorking(false);
      return;
    }

    try {
      await upsertMyGroupRegistrations(group.id, {
        duo_request: registrationDraft.duo_request,
        events: selectedValidEventIds.map((eventId) => {
          const eventDraft = registrationDraft.per_event[eventId] ?? {
            can_substitute: true,
            can_lobby_host: false,
          };
          return {
            event_id: eventId,
            can_substitute: eventDraft.can_substitute,
            can_lobby_host: eventDraft.can_lobby_host,
          };
        }),
      });
      setRegistrationEditorOpen(false);
      await loadGroup();
    } catch (err) {
      setRegistrationError(
        `Your game settings were saved, but registration update failed: ${extractApiError(
          err
        )}. Complete your game profile, then retry Save Registration.`
      );
    } finally {
      setWorking(false);
    }
  };

  const openDeleteConfirmation = (
    registration: EventRegistration,
    gameNumber: number,
    mode: "single" | "all"
  ) => {
    if (!group) return;
    const registrationsInGroup = group.events.reduce(
      (count, event) => count + (event.registrations.some((item) => item.user_id === registration.user_id) ? 1 : 0),
      0
    );
    setPendingDeleteAction({
      mode,
      userId: registration.user_id,
      userName: registration.discord_name || "User",
      eventId: registration.event_id,
      gameNumber,
      registrationsInGroup,
    });
    setDeleteWarningSheetOpen(true);
  };

  const closeDeleteWarningSheet = () => {
    setDeleteWarningSheetOpen(false);
    setPendingDeleteAction(null);
  };

  const openDeleteAllForCurrentUserConfirmation = () => {
    if (!group || !user?.id) return;
    const currentUserRegistration =
      group.events
        .flatMap((event, index) => event.registrations.map((item) => ({ item, gameNumber: index + 1 })))
        .find(({ item }) => item.user_id === user.id) ?? null;
    if (!currentUserRegistration) return;
    openDeleteConfirmation(currentUserRegistration.item, currentUserRegistration.gameNumber, "all");
  };

  const handleDeleteRegistration = async () => {
    if (!group || !pendingDeleteAction) return;
    try {
      setWorking(true);
      if (pendingDeleteAction.mode === "single") {
        await deleteRegistration(pendingDeleteAction.eventId, pendingDeleteAction.userId);
      } else {
        await deleteAllRegistrationsForUserInGroup(pendingDeleteAction.userId);
      }
      closeDeleteWarningSheet();
      await loadGroup();
    } catch {
      setPageError("Could not delete registration.");
    } finally {
      setWorking(false);
    }
  };
  const deletingSelf = !!(pendingDeleteAction && user?.id && pendingDeleteAction.userId === user.id);

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
        <div ref={topAnchorRef} />
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
              disabled={registrationEditorOpen}
              onClick={() => setShowAllEvents((v) => !v)}
              className="text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text-soft)] underline underline-offset-2 disabled:opacity-40 disabled:cursor-not-allowed disabled:no-underline"
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
                  if (showAllEvents) {
                    scrollToEventSection(event.id);
                    return;
                  }
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

        {group.registration_open &&
          (registrationEditorOpen || showAllEvents || (!showAllEvents && activeEvent)) && (
          <div className="flex justify-center">
            {registrationEditorOpen ? (
              <button
                type="button"
                onClick={() => {
                  void handleCloseRegistrationEditor();
                }}
                disabled={working}
                className="rounded-lg border border-[var(--color-text-danger)]/40 bg-[var(--color-text-danger)]/10 px-5 py-2.5 text-sm font-medium text-[var(--color-text-danger)] hover:bg-[var(--color-text-danger)]/20 disabled:opacity-40"
              >
                Cancel
              </button>
            ) : (
              <button
                type="button"
                onClick={() => {
                  void handleOpenRegistrationSheet();
                }}
                className="rounded-lg border border-[var(--color-accent-blue)]/35 bg-[var(--color-accent-blue)]/10 px-5 py-2.5 text-sm font-medium text-[var(--color-accent-blue)] hover:bg-[var(--color-accent-blue)]/20"
              >
                {myRegistrationsByEvent.size > 0 ? "Edit My Registration" : "Register Now"}
              </button>
            )}
          </div>
        )}

        {pageError && (
          <p className="rounded-lg border border-[var(--color-text-danger)]/30 bg-[var(--color-text-danger)]/10 px-3 py-2 text-sm text-[var(--color-text-danger)]">
            {pageError}
          </p>
        )}

        {registrationEditorOpen ? (
          <div className="card rounded-xl p-4 sm:p-5 flex flex-col gap-4 relative overflow-hidden">
            <div className="absolute top-0 left-4 right-4 h-px bg-top-edge opacity-20 rounded-full" />
            <h2 className="text-sm font-semibold text-[var(--color-text)]">
              {myRegistrationsByEvent.size > 0 ? "Edit registration" : "Register"}
            </h2>
            {regionMismatchWarning && (
              <p className="text-xs text-amber-300">{regionMismatchWarning}</p>
            )}
            <UserGameEditor
              hideGameSelector
              gameLabel={group.game_name}
              value={userGameDraft}
              ranks={userGameRanks}
              ranksLoading={registrationLoading}
              errors={userGameErrors}
              onChange={(next) => setUserGameDraft(next)}
            />

            <div className="flex flex-col gap-4">
              <p className="text-sm text-[var(--color-text-soft)]">Choose games to register for</p>
              {group.events.map((event, index) => {
                const checked = registrationDraft.selected_event_ids.includes(event.id);
                const settings = registrationDraft.per_event[event.id] ?? {
                  can_substitute: true,
                  can_lobby_host: false,
                };
                return (
                  <div key={event.id} className="flex flex-col gap-3">
                    <div className="flex items-start gap-3 select-none">
                      <ToggleSwitch
                        checked={checked}
                        onChange={(nextChecked) => {
                          setRegistrationDraft((prev) => {
                            if (nextChecked) {
                              const nextIds = prev.selected_event_ids.includes(event.id)
                                ? prev.selected_event_ids
                                : [...prev.selected_event_ids, event.id];
                              return {
                                ...prev,
                                selected_event_ids: nextIds,
                                per_event: {
                                  ...prev.per_event,
                                  [event.id]: prev.per_event[event.id] ?? {
                                    can_substitute: true,
                                    can_lobby_host: false,
                                  },
                                },
                              };
                            }
                            return {
                              ...prev,
                              selected_event_ids: prev.selected_event_ids.filter((id) => id !== event.id),
                            };
                          });
                        }}
                        className="mt-0.5"
                      />
                      <span className="text-sm font-semibold text-[var(--color-text)] leading-snug pt-0.5">
                        Game {index + 1} · {formatDateTime(event.start_time)}
                      </span>
                    </div>
                    <div className="ml-14 rounded-lg border border-white/[0.06] bg-white/[0.02] p-3 flex flex-col gap-2">
                      <ToggleRow
                        label="Can substitute"
                        checked={settings.can_substitute}
                        disabled={!checked}
                        onChange={(val) =>
                          setRegistrationDraft((prev) => ({
                            ...prev,
                            per_event: {
                              ...prev.per_event,
                              [event.id]: {
                                ...(prev.per_event[event.id] ?? { can_substitute: true, can_lobby_host: false }),
                                can_substitute: val,
                              },
                            },
                          }))
                        }
                      />
                      <ToggleRow
                        label="Can lobby host"
                        checked={settings.can_lobby_host}
                        disabled={!checked}
                        onChange={(val) =>
                          setRegistrationDraft((prev) => ({
                            ...prev,
                            per_event: {
                              ...prev.per_event,
                              [event.id]: {
                                ...(prev.per_event[event.id] ?? { can_substitute: true, can_lobby_host: false }),
                                can_lobby_host: val,
                              },
                            },
                          }))
                        }
                      />
                    </div>
                  </div>
                );
              })}
            </div>

            <div className="flex flex-col gap-1.5">
              <label className="text-sm text-[var(--color-text-soft)]">Duo request (applies to each selected event; cannot be guaranteed)</label>
              <input
                className={inputCls}
                placeholder="Optional Discord name"
                value={registrationDraft.duo_request}
                onChange={(event) =>
                  setRegistrationDraft((prev) => ({ ...prev, duo_request: event.target.value }))
                }
              />
            </div>
            {selectedValidEventIds.length === 0 && !canDeleteAllViaSave && (
              <p className="text-xs text-[var(--color-text-danger)]">Select at least one event to save your registration.</p>
            )}
            {registrationError && (
              <p className="text-xs text-[var(--color-text-danger)]">{registrationError}</p>
            )}
          </div>
        ) : showAllEvents ? (
          <div className="flex flex-col gap-4">
            {group.events.map((event, index) => (
              <div
                key={event.id}
                ref={(node) => {
                  eventSectionRefs.current[event.id] = node;
                }}
                className="flex flex-col gap-3 scroll-mt-24"
              >
                <h2 className="relative flex items-center justify-between text-sm font-semibold text-[var(--color-text)]">
                  <span className="relative z-[1]">Game {index + 1}</span>
                  <span className="pointer-events-none absolute left-1/2 -translate-x-1/2 whitespace-nowrap text-center">
                    {formatDateTime(event.start_time)}
                  </span>
                  <span className="relative z-[1] text-right">{formatPlayerCount(event.registrations.length)}</span>
                </h2>
                {!group.registration_open && event.lobbies_count > 0 ? (
                  <div className="rounded-xl border border-dashed border-white/[0.08] p-4 text-sm text-[var(--color-text-muted)]">
                    Teams display placeholder. Team assignment UI is coming next, but this section already uses the same
                    registration cards for consistency.
                    <div className="mt-3">
                      <EventPanel
                        event={event}
                        gameNumber={index + 1}
                        eventRegion={group.region}
                        currentUserRegion={user?.region}
                        isHostView={isHost}
                        currentUserId={user?.id}
                        onShowDetails={(registration) => {
                          setSelectedRegistration(registration);
                          setDetailsSheetOpen(true);
                        }}
                        onDeleteRegistrationForGame={(registration, gameNumber) =>
                          openDeleteConfirmation(registration, gameNumber, "single")
                        }
                        onDeleteAllFromUser={(registration, gameNumber) =>
                          openDeleteConfirmation(registration, gameNumber, "all")
                        }
                      />
                    </div>
                  </div>
                ) : (
                  <EventPanel
                    event={event}
                    gameNumber={index + 1}
                    eventRegion={group.region}
                    currentUserRegion={user?.region}
                    isHostView={isHost}
                    currentUserId={user?.id}
                    onShowDetails={(registration) => {
                      setSelectedRegistration(registration);
                      setDetailsSheetOpen(true);
                    }}
                    onDeleteRegistrationForGame={(registration, gameNumber) =>
                      openDeleteConfirmation(registration, gameNumber, "single")
                    }
                    onDeleteAllFromUser={(registration, gameNumber) =>
                      openDeleteConfirmation(registration, gameNumber, "all")
                    }
                  />
                )}
                <div className="flex justify-end">
                  <button
                    type="button"
                    onClick={scrollToTop}
                    className="rounded-lg border border-white/10 bg-white/[0.03] px-3 py-1.5 text-xs font-medium text-[var(--color-text-muted)] hover:bg-white/[0.08] hover:text-[var(--color-text-soft)]"
                  >
                    Back to top
                  </button>
                </div>
              </div>
            ))}
          </div>
        ) : activeEvent ? (
          <div className="flex flex-col gap-3">
            <h2 className="relative flex items-center justify-between text-sm font-semibold text-[var(--color-text)]">
              <span className="relative z-[1]">Game {activeEventNumber}</span>
              <span className="pointer-events-none absolute left-1/2 -translate-x-1/2 whitespace-nowrap text-center">
                {formatDateTime(activeEvent.start_time)}
              </span>
              <span className="relative z-[1] text-right">{formatPlayerCount(activeEvent.registrations.length)}</span>
            </h2>
            {!group.registration_open && activeEvent.lobbies_count > 0 ? (
              <div className="rounded-xl border border-dashed border-white/[0.08] p-4 text-sm text-[var(--color-text-muted)]">
                Teams display placeholder. Team assignment UI is coming next, but player cards are reused here.
                <div className="mt-3">
                  <EventPanel
                    event={activeEvent}
                    gameNumber={activeEventNumber}
                    eventRegion={group.region}
                    currentUserRegion={user?.region}
                    isHostView={isHost}
                    currentUserId={user?.id}
                    onShowDetails={(registration) => {
                      setSelectedRegistration(registration);
                      setDetailsSheetOpen(true);
                    }}
                    onDeleteRegistrationForGame={(registration, gameNumber) =>
                      openDeleteConfirmation(registration, gameNumber, "single")
                    }
                    onDeleteAllFromUser={(registration, gameNumber) =>
                      openDeleteConfirmation(registration, gameNumber, "all")
                    }
                  />
                </div>
              </div>
            ) : (
              <EventPanel
                event={activeEvent}
                gameNumber={activeEventNumber}
                eventRegion={group.region}
                currentUserRegion={user?.region}
                isHostView={isHost}
                currentUserId={user?.id}
                onShowDetails={(registration) => {
                  setSelectedRegistration(registration);
                  setDetailsSheetOpen(true);
                }}
                onDeleteRegistrationForGame={(registration, gameNumber) =>
                  openDeleteConfirmation(registration, gameNumber, "single")
                }
                onDeleteAllFromUser={(registration, gameNumber) =>
                  openDeleteConfirmation(registration, gameNumber, "all")
                }
              />
            )}
            <div className="flex justify-end">
              <button
                type="button"
                onClick={scrollToTop}
                className="rounded-lg border border-white/10 bg-white/[0.03] px-3 py-1.5 text-xs font-medium text-[var(--color-text-muted)] hover:bg-white/[0.08] hover:text-[var(--color-text-soft)]"
              >
                Back to top
              </button>
            </div>
          </div>
        ) : null}

        {registrationEditorOpen && (
          <div className="w-full flex justify-end pt-2 pb-1 border-t border-white/[0.08]">
            <div className="flex flex-col items-end gap-2">
              {!canDeleteAllViaSave && hasUserGameErrors && (
                <p className="text-xs text-[var(--color-text-muted)]">
                  Save is disabled until your in-game name, current rank, and peak rank are filled out.
                </p>
              )}
              <button
                type="button"
                onClick={() => {
                  if (!canSubmitRegistration) return;
                  void handleSaveRegistration();
                }}
                disabled={!canSubmitRegistration}
                aria-disabled={!canSubmitRegistration}
                className={[
                  "rounded-lg border px-5 py-2.5 text-sm font-medium",
                  canSubmitRegistration
                    ? canDeleteAllViaSave
                      ? "border-[var(--color-text-danger)]/40 bg-[var(--color-text-danger)]/10 text-[var(--color-text-danger)] hover:bg-[var(--color-text-danger)]/20"
                      : "border-[var(--color-accent-blue)]/35 bg-[var(--color-accent-blue)]/10 text-[var(--color-accent-blue)] hover:bg-[var(--color-accent-blue)]/20"
                    : "cursor-not-allowed border-white/10 bg-white/[0.03] text-[var(--color-text-muted)]",
                ].join(" ")}
              >
                {working ? (canDeleteAllViaSave ? "Deleting..." : "Saving...") : canDeleteAllViaSave ? "Delete Registration" : "Save Registration"}
              </button>
            </div>
          </div>
        )}
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
            sort_logic: group.sort_logic,
          }}
          onCancel={() => setEditSheetOpen(false)}
          onSubmitted={() => {
            void loadGroup();
            setToast("Event settings updated.");
          }}
        />
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
              <p className="text-[var(--color-text-soft)]">{selectedRegistration.discord_name || EMPTY_VALUE}</p>
            </div>
            <div>
              <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Pronouns</p>
              <p className="text-[var(--color-text-soft)]">{selectedRegistration.pronouns || EMPTY_VALUE}</p>
            </div>
            <div>
              <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Current rank</p>
              <p className="text-[var(--color-text-soft)]">{selectedRegistration.current_rank_name || EMPTY_VALUE}</p>
            </div>
            <div>
              <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Peak rank</p>
              <p className="text-[var(--color-text-soft)]">{selectedRegistration.peak_rank_name || EMPTY_VALUE}</p>
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
            <div>
              <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Duo request</p>
              <p className="text-[var(--color-text-soft)]">{selectedRegistration.duo_request || EMPTY_VALUE}</p>
            </div>
          </div>
        ) : (
          <p className="text-sm text-[var(--color-text-muted)]">No registration selected.</p>
        )}
      </ResponsiveSheet>

      <ResponsiveSheet
        isOpen={deleteWarningSheetOpen}
        onClose={closeDeleteWarningSheet}
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
              onClick={closeDeleteWarningSheet}
              className="rounded-lg border border-white/10 bg-white/[0.03] px-3 py-2 text-sm text-[var(--color-text-soft)]"
            >
              Cancel
            </button>
            <button
              type="button"
              onClick={() => {
                void handleDeleteRegistration();
              }}
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
