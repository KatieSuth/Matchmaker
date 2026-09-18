"use client";

// Register/edit-registration sheet: per-game selection, substitute/lobby-host preferences, duo
// request, and the participant's own game profile (in-game name + ranks) editing.
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { UserGameEditorValue } from "@/app/_components/forms/UserGameEditor";
import { extractApiError, extractDiscordGuildRestriction, fetchGameRanks } from "@/app/_services/games";
import { upsertMyGroupRegistrations } from "@/app/_services/events";
import { fetchCurrentUserGames, upsertCurrentUserGame } from "@/app/_services/users";
import {
  DiscordGuildRestrictionDetails,
} from "@/app/_services/games";
import { EventGroupDetail, EventRegistration, GameRank, User } from "@/app/_types/types";
import { EventRegistrationDraft, RegistrationDraft } from "../_types";

function emptyUserGameDraft(gameId: string): UserGameEditorValue {
  return {
    game_id: gameId,
    in_game_name: "",
    current_rank: "",
    peak_rank: "",
    show_rank: false,
  };
}

export interface UseRegistrationEditorOptions {
  groupId: string | undefined;
  group: EventGroupDetail | null;
  user: User | null;
  authLoading: boolean;
  isAuthenticated: boolean;
  pageLoading: boolean;
  isHost: boolean;
  myRegistrationsByEvent: Map<string, EventRegistration>;
  hasAnyLobbies: boolean;
  working: boolean;
  setWorking: (working: boolean) => void;
  loadGroup: () => Promise<void>;
  setAccessDenial: (denial: DiscordGuildRestrictionDetails | null) => void;
  setGroup: (group: EventGroupDetail | null) => void;
  openDeleteAllForCurrentUserConfirmation: () => void;
}

/** Manages the registration editor sheet's draft state, auto-open behavior, and save/delete flow. */
export function useRegistrationEditor({
  groupId,
  group,
  user,
  authLoading,
  isAuthenticated,
  pageLoading,
  isHost,
  myRegistrationsByEvent,
  hasAnyLobbies,
  working,
  setWorking,
  loadGroup,
  setAccessDenial,
  setGroup,
  openDeleteAllForCurrentUserConfirmation,
}: UseRegistrationEditorOptions) {
  const [registrationEditorOpen, setRegistrationEditorOpen] = useState(false);
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
  /** Prevents re-opening the registration form after the user cancels an auto-open. */
  const didAutoOpenRegistrationRef = useRef(false);

  useEffect(() => {
    didAutoOpenRegistrationRef.current = false;
  }, [groupId]);

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
    [userGameDraft.current_rank, userGameDraft.in_game_name, userGameDraft.peak_rank],
  );
  const hasUserGameErrors = !!(userGameErrors.in_game_name || userGameErrors.current_rank || userGameErrors.peak_rank);
  const registrationIsEditMode = myRegistrationsByEvent.size > 0;
  const canDeleteAllViaSave =
    registrationIsEditMode &&
    selectedValidEventIds.length === 0 &&
    !!user?.id &&
    !hasAnyLobbies &&
    !working &&
    !registrationLoading;
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

  const handleOpenRegistrationSheet = useCallback(
    async (registration?: EventRegistration) => {
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
            : emptyUserGameDraft(group.game_id),
        );
      } catch (err) {
        setUserGameRanks([]);
        setUserGameDraft(emptyUserGameDraft(group.game_id));
        setRegistrationError(extractApiError(err, "Could not load your game settings. Please try again."));
      } finally {
        setRegistrationLoading(false);
      }
      setRegistrationEditorOpen(true);
    },
    [group, myRegistrationsByEvent],
  );

  const handleCloseRegistrationEditor = async () => {
    setRegistrationError(null);
    setRegistrationEditorOpen(false);
    await loadGroup();
  };

  // Auto-open register form for guests who aren't registered yet (same as clicking Register Now).
  useEffect(() => {
    if (authLoading || !isAuthenticated || !user || !group || pageLoading) return;
    if (didAutoOpenRegistrationRef.current || registrationEditorOpen) return;
    if (isHost || myRegistrationsByEvent.size > 0 || !group.registration_open) return;

    didAutoOpenRegistrationRef.current = true;
    const timer = window.setTimeout(() => {
      void handleOpenRegistrationSheet();
    }, 0);
    return () => window.clearTimeout(timer);
    // handleOpenRegistrationSheet closes over the latest group/registrations for this render.
    // eslint-disable-next-line react-hooks/exhaustive-deps -- open once per group visit via ref guard
  }, [
    authLoading,
    group,
    isAuthenticated,
    isHost,
    pageLoading,
    myRegistrationsByEvent.size,
    registrationEditorOpen,
    user,
  ]);

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
        `${extractApiError(err)}. Complete your game profile, then try again.`,
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
      const restriction = extractDiscordGuildRestriction(err);
      if (restriction) {
        setAccessDenial(restriction);
        setGroup(null);
        setRegistrationEditorOpen(false);
        return;
      }
      setRegistrationError(
        `Your game settings were saved, but registration update failed: ${extractApiError(
          err,
        )}. Complete your game profile, then retry Save Registration.`,
      );
    } finally {
      setWorking(false);
    }
  };

  return {
    registrationEditorOpen,
    registrationDraft,
    setRegistrationDraft,
    userGameDraft,
    setUserGameDraft,
    userGameRanks,
    userGameErrors,
    hasUserGameErrors,
    registrationError,
    registrationLoading,
    registrationIsEditMode,
    selectedValidEventIds,
    canDeleteAllViaSave,
    canSaveRegistration,
    canSubmitRegistration,
    regionMismatchWarning,
    handleOpenRegistrationSheet,
    handleCloseRegistrationEditor,
    handleSaveRegistration,
  };
}
