"use client";

// Event group detail: metadata, per-game registration panels, host controls (teams, registration),
// and participant registration / profile actions. Large presentational pieces live in local helpers below.
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useParams } from "next/navigation";
import { EllipsisMenu, EllipsisMenuOption } from "@/app/_components/EllipsisMenu";
import { ResponsiveSheet } from "@/app/_components/ResponsiveSheet";
import { Select, SelectOption } from "@/app/_components/Select";
import { LobbyHostInfoHint } from "@/app/_components/LobbyHostInfoHint";
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
  deleteTeams,
  fetchEventGroup,
  swapPlayers,
  upsertMyGroupRegistrations,
} from "@/app/_services/events";
import { fetchCurrentUserGames, upsertCurrentUserGame } from "@/app/_services/users";
import {
  EventGroupDetail,
  EventGroupEvent,
  EventLobby,
  EventRegistration,
  GameRank,
  LobbyPlayer,
} from "@/app/_types/types";

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

/** Identifies a roster/sub/unplaced player within a locked-in game for host swap actions. */
interface PlayerPlacement {
  eventId: string;
  userId: string;
  discordName: string;
  lobbyId: string | null;
  sourceLobbyIndex: number | null;
  teamNumber: number | null | undefined;
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
  if (!value.trim()) return EMPTY_VALUE;
  const date = new Date(value);
  if (!Number.isFinite(date.getTime())) return EMPTY_VALUE;
  return DATE_TIME_FMT.format(date);
}

function formatPlayerCount(count: number) {
  return `${count} ${count === 1 ? "Player" : "Players"}`;
}

/** Per-player skill value used for team averages — mirrors backend (current + peak) / 2. */
function playerSkillOrder(player: LobbyPlayer): number | null {
  if (player.current_rank_order <= 0 || player.peak_rank_order <= 0) {
    return null;
  }
  return (player.current_rank_order + player.peak_rank_order) / 2;
}

/** Maps a numeric skill value to the closest game rank name by order distance. */
function nearestRankName(ranks: GameRank[], skillOrder: number): string {
  if (ranks.length === 0) return EMPTY_VALUE;
  let nearest = ranks[0];
  let bestDistance = Math.abs(nearest.order - skillOrder);
  for (const rank of ranks) {
    const distance = Math.abs(rank.order - skillOrder);
    if (distance < bestDistance) {
      bestDistance = distance;
      nearest = rank;
    }
  }
  return nearest.name;
}

/** Returns the nearest rank name for a team's mean skill, or EMPTY_VALUE when no rank data exists. */
function teamAverageRankLabel(players: LobbyPlayer[], ranks: GameRank[]): string {
  const skillOrders = players
    .map(playerSkillOrder)
    .filter((order): order is number => order !== null);
  if (skillOrders.length === 0) {
    return EMPTY_VALUE;
  }
  const average =
    skillOrders.reduce((sum, order) => sum + order, 0) / skillOrders.length;
  return nearestRankName(ranks, average);
}

function formatGroupTeamSizeLabel(events: EventGroupEvent[]) {
  const sizes = [...new Set(events.map((e) => e.team_size))];
  if (sizes.length === 0) return "—";
  if (sizes.length === 1) return String(sizes[0]);
  return "Varies";
}

function formatHostDisplayLabel(isViewerHost: boolean, ownerName: string, ownerPronouns: string) {
  if (isViewerHost) return "You";
  const pronouns = ownerPronouns.trim();
  if (pronouns) return `${ownerName} (${pronouns})`;
  return ownerName;
}

/** Compact subtitle fragment for chips / roster rows (game mode · formatted time). */
function formatGameModeAndTime(modeName: string, startISO: string) {
  return `${modeName} · ${formatDateTime(startISO)}`;
}

/** Human-readable placement label for swap candidates (team, subs, or unplaced). */
function formatPlacementCategory(
  sourceLobbyIndex: number | null,
  targetLobbyIndex: number,
  teamNumber: number | null | undefined,
): string {
  if (teamNumber === undefined) {
    return "Unplaced";
  }
  if (teamNumber === null) {
    if (sourceLobbyIndex !== null && sourceLobbyIndex === targetLobbyIndex) {
      return "Subs";
    }
    return `Lobby ${targetLobbyIndex + 1} · Subs`;
  }
  if (sourceLobbyIndex !== null && sourceLobbyIndex === targetLobbyIndex) {
    return `Team ${teamNumber}`;
  }
  return `Lobby ${targetLobbyIndex + 1} · Team ${teamNumber}`;
}

/** Formats a swap dropdown option as "Name (Category) · Rank". */
function formatSwapCandidateLabel(
  discordName: string,
  category: string,
  currentRankName?: string,
): string {
  const name = discordName || "Unknown user";
  const rank = currentRankName?.trim() || EMPTY_VALUE;
  return `${name} (${category}) · ${rank}`;
}

/** Lists eligible swap targets for a player, excluding same-team roster mates and same-lobby subs. */
function buildSwapCandidates(event: EventGroupEvent, source: PlayerPlacement): SelectOption[] {
  const options: SelectOption[] = [];
  const lobbies = event.lobbies ?? [];

  for (let lobbyIndex = 0; lobbyIndex < lobbies.length; lobbyIndex++) {
    const lobby = lobbies[lobbyIndex];
    for (const team of lobby.teams) {
      for (const player of team.players) {
        if (player.user_id === source.userId) {
          continue;
        }
        if (
          source.teamNumber !== undefined &&
          source.teamNumber !== null &&
          source.lobbyId === lobby.id &&
          source.teamNumber === team.team_number
        ) {
          continue;
        }
        const category = formatPlacementCategory(source.sourceLobbyIndex, lobbyIndex, team.team_number);
        options.push({
          value: player.user_id,
          label: formatSwapCandidateLabel(player.discord_name, category, player.current_rank_name),
        });
      }
    }
    for (const player of lobby.subs) {
      if (player.user_id === source.userId) {
        continue;
      }
      if (source.teamNumber === null && source.lobbyId === lobby.id) {
        continue;
      }
      const category = formatPlacementCategory(source.sourceLobbyIndex, lobbyIndex, null);
      options.push({
        value: player.user_id,
        label: formatSwapCandidateLabel(player.discord_name, category, player.current_rank_name),
      });
    }
  }

  for (const registration of event.unplaced ?? []) {
    if (registration.user_id === source.userId) {
      continue;
    }
    options.push({
      value: registration.user_id,
      label: formatSwapCandidateLabel(registration.discord_name, "Unplaced", registration.current_rank_name),
    });
  }

  options.sort((a, b) => a.label.localeCompare(b.label));
  return options;
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
  placement,
  onSwap,
  showDuoRequest = false,
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
  placement?: PlayerPlacement;
  onSwap?: (placement: PlayerPlacement) => void;
  showDuoRequest?: boolean;
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
  if (isHostView && placement && onSwap) {
    menuOptions.push({
      label: "Swap",
      onSelect: () => onSwap(placement),
    });
  }
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
        <div>
          <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Can substitute</p>
          <p className="text-xs text-[var(--color-text-soft)]">{registration.can_substitute ? "Yes" : "No"}</p>
        </div>
        {showDuoRequest && (
          <div>
            <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Duo request</p>
            <p className="text-xs text-[var(--color-text-soft)] truncate" title={registration.duo_request || undefined}>
              {registration.duo_request || EMPTY_VALUE}
            </p>
          </div>
        )}
      </div>
    </div>
  );
}

/** True when any lobby in the game is currently flagged unfair. */
function eventHasUnfairLobby(event: EventGroupEvent): boolean {
  return (event.lobbies ?? []).some((lobby) => lobby.fairness_warning);
}

/** Returns the lobby fairness banner copy for lock-in vs post-edit warnings. */
function lobbyFairnessWarningMessage(lobby: EventLobby): string {
  if (lobby.fairness_warning_at_lock) {
    return "Teams were formed with the best available balance, but rank spread was too wide for fully fair teams in this lobby.";
  }
  return "This lobby was fair when teams were locked in, but a manual roster change has made the rank spread too wide for fully fair teams.";
}

/** Adapts a lobby player row so TeamsPanel can reuse PlayerCard and registration actions. */
function lobbyPlayerAsRegistration(player: LobbyPlayer, eventId: string): EventRegistration {
  return {
    event_id: eventId,
    user_id: player.user_id,
    discord_name: player.discord_name,
    pronouns: player.pronouns,
    current_rank_name: player.current_rank_name,
    can_substitute: player.can_substitute,
    can_lobby_host: player.can_lobby_host,
    duo_request: player.duo_request,
    created_at: player.created_at,
    updated_at: player.updated_at,
  };
}

/** Resolves the lobby host's display name from roster, subs, or unplaced players. */
function lobbyHostName(lobby: EventLobby, event: EventGroupEvent): string | null {
  if (!lobby.host_id) return null;
  const allPlayers = [
    ...lobby.teams.flatMap((team) => team.players),
    ...lobby.subs,
    ...event.unplaced,
  ];
  const host = allPlayers.find((p) => p.user_id === lobby.host_id);
  return host?.discord_name ?? null;
}

function TeamsPanel({
  event,
  gameNumber,
  eventRegion,
  currentUserRegion,
  isHostView,
  currentUserId,
  gameRanks,
  onShowDetails,
  onDeleteRegistrationForGame,
  onDeleteAllFromUser,
  onSwapPlayer,
}: {
  event: EventGroupEvent;
  gameNumber: number;
  eventRegion: string;
  currentUserRegion?: string | null;
  isHostView: boolean;
  currentUserId?: string;
  gameRanks: GameRank[];
  onShowDetails: (registration: EventRegistration) => void;
  onDeleteRegistrationForGame: (registration: EventRegistration, gameNumber: number) => void;
  onDeleteAllFromUser: (registration: EventRegistration, gameNumber: number) => void;
  onSwapPlayer?: (placement: PlayerPlacement) => void;
}) {
  const lobbies = event.lobbies ?? [];
  if (lobbies.length === 0) {
    return null;
  }

  return (
    <div className="flex flex-col gap-4">
      {lobbies.map((lobby, lobbyIndex) => (
        <div key={lobby.id} className="flex flex-col gap-3">
          {lobby.fairness_warning && (
            <div className="rounded-lg border border-amber-400/30 bg-amber-400/10 px-3 py-2 text-xs text-amber-200">
              {lobbyFairnessWarningMessage(lobby)}
            </div>
          )}
          <div className="flex items-center gap-2 text-sm font-semibold text-[var(--color-text)]">
            <span>
              Lobby {lobbyIndex + 1}
              {lobbyHostName(lobby, event) ? ` · Host: ${lobbyHostName(lobby, event)}` : ""}
            </span>
            {lobby.fairness_warning && <span className="text-amber-400" aria-label="Unfair lobby">⚠</span>}
          </div>
          <div className="grid gap-3 sm:grid-cols-2">
            {lobby.teams.map((team) => {
              const averageRank = isHostView ? teamAverageRankLabel(team.players, gameRanks) : EMPTY_VALUE;
              return (
              <div key={team.team_number} className="flex flex-col gap-2">
                <p className="text-xs font-medium uppercase tracking-wide text-[var(--color-text-faint)]">
                  Team {team.team_number}
                  {averageRank !== EMPTY_VALUE ? ` · Average: ${averageRank}` : ""}
                </p>
                {team.players.map((player) => (
                  <PlayerCard
                    key={player.user_id}
                    registration={lobbyPlayerAsRegistration(player, event.id)}
                    gameNumber={gameNumber}
                    eventRegion={eventRegion}
                    currentUserRegion={currentUserRegion}
                    isHostView={isHostView}
                    currentUserId={currentUserId}
                    canEditRegistration={false}
                    onShowDetails={onShowDetails}
                    onDeleteRegistrationForGame={onDeleteRegistrationForGame}
                    onDeleteAllFromUser={onDeleteAllFromUser}
                    showDuoRequest
                    placement={{
                      eventId: event.id,
                      userId: player.user_id,
                      discordName: player.discord_name || "Unknown user",
                      lobbyId: lobby.id,
                      sourceLobbyIndex: lobbyIndex,
                      teamNumber: team.team_number,
                    }}
                    onSwap={onSwapPlayer}
                  />
                ))}
              </div>
            );
            })}
          </div>
          {lobby.subs.length > 0 && (
            <div className="flex flex-col gap-2">
              <p className="text-xs font-medium uppercase tracking-wide text-[var(--color-text-faint)]">Subs</p>
              {lobby.subs.map((player) => (
                <PlayerCard
                  key={player.user_id}
                  registration={lobbyPlayerAsRegistration(player, event.id)}
                  gameNumber={gameNumber}
                  eventRegion={eventRegion}
                  currentUserRegion={currentUserRegion}
                  isHostView={isHostView}
                  currentUserId={currentUserId}
                  canEditRegistration={false}
                  onShowDetails={onShowDetails}
                  onDeleteRegistrationForGame={onDeleteRegistrationForGame}
                  onDeleteAllFromUser={onDeleteAllFromUser}
                  showDuoRequest
                  placement={{
                    eventId: event.id,
                    userId: player.user_id,
                    discordName: player.discord_name || "Unknown user",
                    lobbyId: lobby.id,
                    sourceLobbyIndex: lobbyIndex,
                    teamNumber: null,
                  }}
                  onSwap={onSwapPlayer}
                />
              ))}
            </div>
          )}
        </div>
      ))}
      {(event.unplaced ?? []).length > 0 && (
        <div className="flex flex-col gap-2">
          <p className="text-xs font-medium uppercase tracking-wide text-[var(--color-text-faint)]">Unplaced</p>
          <p className="text-xs text-[var(--color-text-muted)]">
            Registered but not assigned to a team or sub pool for this game.
          </p>
          {event.unplaced.map((registration) => (
            <PlayerCard
              key={registration.user_id}
              registration={registration}
              gameNumber={gameNumber}
              eventRegion={eventRegion}
              currentUserRegion={currentUserRegion}
              isHostView={isHostView}
              currentUserId={currentUserId}
              canEditRegistration={false}
              onShowDetails={onShowDetails}
              onDeleteRegistrationForGame={onDeleteRegistrationForGame}
              onDeleteAllFromUser={onDeleteAllFromUser}
              showDuoRequest
              placement={{
                eventId: event.id,
                userId: registration.user_id,
                discordName: registration.discord_name || "Unknown user",
                lobbyId: null,
                sourceLobbyIndex: null,
                teamNumber: undefined,
              }}
              onSwap={onSwapPlayer}
            />
          ))}
        </div>
      )}
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
  const groupId = Array.isArray(params?.groupId) ? params.groupId[0] : params?.groupId;
  const { user, isAuthenticated, isLoading: authLoading } = useAuth();

  const [group, setGroup] = useState<EventGroupDetail | null>(null);
  const [pageError, setPageError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [working, setWorking] = useState(false);
  const [activeEventId, setActiveEventId] = useState<string | null>(null);
  const [showAllEvents, setShowAllEvents] = useState(false);
  const [editSheetOpen, setEditSheetOpen] = useState(false);
  const [registrationEditorOpen, setRegistrationEditorOpen] = useState(false);
  const [detailsSheetOpen, setDetailsSheetOpen] = useState(false);
  const [warningSheetOpen, setWarningSheetOpen] = useState(false);
  const [deleteWarningSheetOpen, setDeleteWarningSheetOpen] = useState(false);
  const [swapSheetOpen, setSwapSheetOpen] = useState(false);
  const [selectedRegistration, setSelectedRegistration] = useState<EventRegistration | null>(null);
  const [pendingDeleteAction, setPendingDeleteAction] = useState<PendingDeleteAction | null>(null);
  const [pendingSwap, setPendingSwap] = useState<PlayerPlacement | null>(null);
  const [swapTargetUserId, setSwapTargetUserId] = useState("");
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
  const [gameRanks, setGameRanks] = useState<GameRank[]>([]);
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
      if (user?.id === data.owner_id) {
        const ranks = await fetchGameRanks(data.game_id, signal);
        if (signal?.aborted) return;
        setGameRanks(ranks);
      } else {
        setGameRanks([]);
      }
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
  }, [activeEventId, groupId, user]);

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
    setWarningSheetOpen(false);
  };

  const withHostAction = async (action: () => Promise<void>) => {
    try {
      setWorking(true);
      await action();
      await refreshAndCloseMenus();
    } catch (err) {
      setPageError(extractApiError(err, "Could not complete that action."));
    } finally {
      setWorking(false);
    }
  };

  const openSwapSheet = useCallback((placement: PlayerPlacement) => {
    setPendingSwap(placement);
    setSwapTargetUserId("");
    setSwapSheetOpen(true);
  }, []);

  const closeSwapSheet = useCallback(() => {
    setSwapSheetOpen(false);
    setPendingSwap(null);
    setSwapTargetUserId("");
  }, []);

  const handleSwapSubmit = async () => {
    if (!pendingSwap || !swapTargetUserId) return;
    try {
      setWorking(true);
      await swapPlayers(pendingSwap.eventId, pendingSwap.userId, swapTargetUserId);
      await loadGroup();
      closeSwapSheet();
    } catch (err) {
      setPageError(extractApiError(err, "Could not complete that action."));
    } finally {
      setWorking(false);
    }
  };

  const swapEvent = useMemo(() => {
    if (!group || !pendingSwap) return null;
    return group.events.find((event) => event.id === pendingSwap.eventId) ?? null;
  }, [group, pendingSwap]);

  const swapCandidateOptions = useMemo(() => {
    if (!swapEvent || !pendingSwap) return [];
    return buildSwapCandidates(swapEvent, pendingSwap);
  }, [swapEvent, pendingSwap]);

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
                <>
                  <button
                    type="button"
                    onClick={() => setEditSheetOpen(true)}
                    className="rounded-lg border border-white/10 bg-white/[0.03] px-3 py-2 text-sm font-medium text-[var(--color-text-soft)] hover:bg-white/[0.08]"
                  >
                    Edit
                  </button>
                  <button
                    type="button"
                    disabled={working}
                    onClick={() => {
                      if (!group.registration_open && hasAnyLobbies) {
                        setWarningSheetOpen(true);
                        return;
                      }
                      void withHostAction(() => createTeams(group.id));
                    }}
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
                </>
              )}
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
                {formatHostDisplayLabel(isHost, group.owner_name, group.owner_pronouns ?? "")}
              </p>
            </div>
          </div>
        </div>

        <div className="card rounded-xl p-3 sm:p-4 flex flex-col gap-3">
          <div className="flex items-center justify-between gap-2">
            <p className="text-sm font-medium text-[var(--color-text-soft)]">Games in this group</p>
            {group.events.length > 1 && (
              <button
                type="button"
                disabled={registrationEditorOpen}
                onClick={() => setShowAllEvents((v) => !v)}
                className="text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text-soft)] underline underline-offset-2 disabled:opacity-40 disabled:cursor-not-allowed disabled:no-underline shrink-0"
              >
                {showAllEvents ? "Show one at a time" : "View all"}
              </button>
            )}
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
                <span className="inline-flex items-center gap-1">
                  Game {index + 1} · {formatGameModeAndTime(event.game_mode_name, event.start_time)}
                  {eventHasUnfairLobby(event) && (
                    <span className="text-amber-400" aria-label="Contains unfair lobby">
                      ⚠
                    </span>
                  )}
                </span>
              </button>
            ))}
          </div>
        </div>

        {(registrationEditorOpen || showAllEvents || (!showAllEvents && activeEvent)) && (
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
              <span
                title={!group.registration_open ? "Registration is closed" : undefined}
                className={[
                  "inline-flex rounded-lg",
                  !group.registration_open ? "cursor-not-allowed" : "",
                ]
                  .filter(Boolean)
                  .join(" ")}
              >
                <button
                  type="button"
                  disabled={!group.registration_open}
                  onClick={() => {
                    void handleOpenRegistrationSheet();
                  }}
                  className={[
                    "rounded-lg border px-5 py-2.5 text-sm font-medium",
                    group.registration_open
                      ? "border-[var(--color-accent-blue)]/35 bg-[var(--color-accent-blue)]/10 text-[var(--color-accent-blue)] hover:bg-[var(--color-accent-blue)]/20"
                      : "pointer-events-none border-white/10 bg-white/[0.03] text-[var(--color-text-muted)] opacity-50",
                  ].join(" ")}
                >
                  {myRegistrationsByEvent.size > 0 ? "Edit My Registration" : "Register Now"}
                </button>
              </span>
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
                        Game {index + 1} · {event.game_mode_name} · {formatDateTime(event.start_time)}
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
                        labelAccessory={<LobbyHostInfoHint />}
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
              <label className="text-sm text-[var(--color-text-soft)]">Duo Request (They must list you here too. Applies to each selected event. Cannot be guaranteed)</label>
              <input
                className={inputCls}
                placeholder="Discord Name"
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
                <h2 className="relative flex items-center justify-between gap-2 text-sm font-semibold text-[var(--color-text)]">
                  <span className="relative z-[1] flex min-w-0 shrink flex-wrap items-baseline gap-x-2 gap-y-0">
                    <span className="shrink-0 whitespace-nowrap">Game {index + 1}</span>
                    <span className="text-[0.65rem] font-normal uppercase tracking-wide text-[var(--color-text-faint)]">
                      {event.game_mode_name}
                    </span>
                  </span>
                  <span className="relative z-[1] shrink-0 text-center text-xs font-normal text-[var(--color-text-muted)]">
                    {formatDateTime(event.start_time)}
                  </span>
                  <span className="relative z-[1] shrink-0 text-right text-xs font-normal text-[var(--color-text-soft)]">
                    {formatPlayerCount(event.registrations.length)}
                  </span>
                </h2>
                {!group.registration_open && event.lobbies_count > 0 ? (
                  <TeamsPanel
                    event={event}
                    gameNumber={index + 1}
                    eventRegion={group.region}
                    currentUserRegion={user?.region}
                    isHostView={isHost}
                    currentUserId={user?.id}
                    gameRanks={gameRanks}
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
                    onSwapPlayer={isHost ? openSwapSheet : undefined}
                  />
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
            <h2 className="relative flex items-center justify-between gap-2 text-sm font-semibold text-[var(--color-text)]">
              <span className="relative z-[1] flex min-w-0 shrink flex-wrap items-baseline gap-x-2 gap-y-0">
                <span className="shrink-0 whitespace-nowrap">Game {activeEventNumber}</span>
                <span className="text-[0.65rem] font-normal uppercase tracking-wide text-[var(--color-text-faint)]">
                  {activeEvent.game_mode_name}
                </span>
              </span>
              <span className="relative z-[1] shrink-0 text-center text-xs font-normal text-[var(--color-text-muted)]">
                {formatDateTime(activeEvent.start_time)}
              </span>
              <span className="relative z-[1] shrink-0 text-right text-xs font-normal text-[var(--color-text-soft)]">
                {formatPlayerCount(activeEvent.registrations.length)}
              </span>
            </h2>
            {!group.registration_open && activeEvent.lobbies_count > 0 ? (
              <TeamsPanel
                event={activeEvent}
                gameNumber={activeEventNumber}
                eventRegion={group.region}
                currentUserRegion={user?.region}
                isHostView={isHost}
                currentUserId={user?.id}
                gameRanks={gameRanks}
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
                onSwapPlayer={isHost ? openSwapSheet : undefined}
              />
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
          editSchedule={group.events.map((e) => ({
            id: e.id,
            start_time: e.start_time,
            game_mode_id: e.game_mode_id,
          }))}
          initialValues={{
            game_id: group.game_id,
            region: group.region,
            sub_min: group.sub_min,
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
              <div className="flex items-center gap-1.5">
                <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Can lobby host</p>
                <LobbyHostInfoHint />
              </div>
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
        isOpen={swapSheetOpen}
        onClose={closeSwapSheet}
        title={pendingSwap ? `Swap ${pendingSwap.discordName}` : "Swap player"}
      >
        <div className="flex flex-col gap-4">
          <p className="text-sm text-[var(--color-text-soft)]">
            Choose a player from another team, lobby, sub pool, or unplaced list to swap with.
          </p>
          <Select
            value={swapTargetUserId}
            onChange={setSwapTargetUserId}
            options={swapCandidateOptions}
            placeholder={swapCandidateOptions.length === 0 ? "No swap candidates" : "— Select player —"}
            disabled={working || swapCandidateOptions.length === 0}
          />
          <div className="flex justify-end gap-2">
            <button
              type="button"
              onClick={closeSwapSheet}
              className="rounded-lg border border-white/10 bg-white/[0.03] px-3 py-2 text-sm text-[var(--color-text-soft)]"
            >
              Cancel
            </button>
            <button
              type="button"
              onClick={() => {
                void handleSwapSubmit();
              }}
              disabled={working || !pendingSwap || !swapTargetUserId}
              className="rounded-lg border border-[var(--color-accent-blue)]/40 bg-[var(--color-accent-blue)]/10 px-3 py-2 text-sm text-[var(--color-accent-blue)] disabled:opacity-40"
            >
              {working ? "Swapping..." : "Submit"}
            </button>
          </div>
        </div>
      </ResponsiveSheet>

      <ResponsiveSheet
        isOpen={warningSheetOpen}
        onClose={() => setWarningSheetOpen(false)}
        title="Delete teams"
      >
        <div className="flex flex-col gap-4">
          <p className="text-sm text-[var(--color-text-soft)]">
            All lobbies and teams for this event will be deleted, but registrations will remain. Registration stays closed until you open it from Edit. This action cannot be undone.
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
              onClick={() => withHostAction(() => deleteTeams(group.id))}
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
