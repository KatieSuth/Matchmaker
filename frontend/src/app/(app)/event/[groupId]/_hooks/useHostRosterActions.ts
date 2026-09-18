"use client";

// Host-only roster management: lock-in/delete teams, swap players, move to subs/unplaced, and
// lobby-host reassignment. All mutations funnel through `withHostAction`, which refreshes the
// group and surfaces failures on the shared page error banner.
import { useCallback, useMemo, useState } from "react";
import { SelectOption } from "@/app/_components/Select";
import {
  createTeams,
  deleteTeams,
  moveSubToUnplaced,
  moveUnplacedToSubs,
  setLobbyHost,
  swapPlayers,
} from "@/app/_services/events";
import { extractApiError } from "@/app/_services/games";
import { EventGroupDetail } from "@/app/_types/types";
import { buildLobbyHostVolunteers, buildSwapCandidates, isTeamAssignedPlacement } from "../_lib/placements";
import { PendingLobbyHostChange, PlayerPlacement } from "../_types";

export function useHostRosterActions(
  group: EventGroupDetail | null,
  loadGroup: () => Promise<void>,
  setWorking: (working: boolean) => void,
  setPageError: (message: string | null) => void,
) {
  const [warningSheetOpen, setWarningSheetOpen] = useState(false);
  const [subCapacitySheetOpen, setSubCapacitySheetOpen] = useState(false);
  const [swapSheetOpen, setSwapSheetOpen] = useState(false);
  const [pendingSwap, setPendingSwap] = useState<PlayerPlacement | null>(null);
  const [swapTargetUserId, setSwapTargetUserId] = useState("");
  const [swapError, setSwapError] = useState<string | null>(null);
  const [moveToSubsSheetOpen, setMoveToSubsSheetOpen] = useState(false);
  const [pendingMoveToSubs, setPendingMoveToSubs] = useState<PlayerPlacement | null>(null);
  const [moveToSubsLobbyId, setMoveToSubsLobbyId] = useState("");
  const [moveToSubsError, setMoveToSubsError] = useState<string | null>(null);
  const [lobbyHostConfirmOpen, setLobbyHostConfirmOpen] = useState(false);
  const [pendingLobbyHostChange, setPendingLobbyHostChange] = useState<PendingLobbyHostChange | null>(null);

  const refreshAndCloseMenus = async () => {
    await loadGroup();
    setWarningSheetOpen(false);
  };

  const withHostAction = useCallback(
    async (action: () => Promise<void>) => {
      try {
        setWorking(true);
        await action();
        await refreshAndCloseMenus();
      } catch (err) {
        setPageError(extractApiError(err, "Could not complete that action."));
      } finally {
        setWorking(false);
      }
    },
    // refreshAndCloseMenus recreated each render but only closes over loadGroup/setWarningSheetOpen
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [loadGroup, setPageError, setWorking],
  );

  const lockInTeams = async () => {
    if (!group) return;
    try {
      setWorking(true);
      const result = await createTeams(group.id);
      await loadGroup();
      if (result.sub_capacity_adjusted) {
        setSubCapacitySheetOpen(true);
      }
    } catch (err) {
      setPageError(extractApiError(err, "Could not complete that action."));
    } finally {
      setWorking(false);
    }
  };

  const confirmDeleteTeams = () => {
    if (!group) return;
    void withHostAction(() => deleteTeams(group.id));
  };

  const openSwapSheet = useCallback((placement: PlayerPlacement) => {
    setPendingSwap(placement);
    setSwapTargetUserId("");
    setSwapError(null);
    setSwapSheetOpen(true);
  }, []);

  const closeSwapSheet = useCallback(() => {
    setSwapSheetOpen(false);
    setPendingSwap(null);
    setSwapTargetUserId("");
    setSwapError(null);
  }, []);

  const handleSwapSubmit = async () => {
    if (!pendingSwap || !swapTargetUserId) return;
    try {
      setWorking(true);
      setSwapError(null);
      await swapPlayers(pendingSwap.eventId, pendingSwap.userId, swapTargetUserId);
      await loadGroup();
      closeSwapSheet();
    } catch (err) {
      setSwapError(extractApiError(err, "Could not complete that action."));
    } finally {
      setWorking(false);
    }
  };

  const closeMoveToSubsSheet = useCallback(() => {
    setMoveToSubsSheetOpen(false);
    setPendingMoveToSubs(null);
    setMoveToSubsLobbyId("");
    setMoveToSubsError(null);
  }, []);

  const handleMoveToUnplaced = (placement: PlayerPlacement) => {
    void withHostAction(() => moveSubToUnplaced(placement.eventId, placement.userId));
  };

  const handleMoveToSubs = (placement: PlayerPlacement) => {
    if (!group) return;
    const event = group.events.find((item) => item.id === placement.eventId);
    if (!event) return;
    const lobbies = event.lobbies ?? [];
    if (lobbies.length === 1) {
      void withHostAction(() =>
        moveUnplacedToSubs(placement.eventId, placement.userId, lobbies[0].id),
      );
      return;
    }
    setPendingMoveToSubs(placement);
    setMoveToSubsLobbyId("");
    setMoveToSubsError(null);
    setMoveToSubsSheetOpen(true);
  };

  const handleMoveToSubsSubmit = async () => {
    if (!pendingMoveToSubs || !moveToSubsLobbyId) return;
    try {
      setWorking(true);
      setMoveToSubsError(null);
      await moveUnplacedToSubs(pendingMoveToSubs.eventId, pendingMoveToSubs.userId, moveToSubsLobbyId);
      await loadGroup();
      closeMoveToSubsSheet();
    } catch (err) {
      setMoveToSubsError(extractApiError(err, "Could not complete that action."));
    } finally {
      setWorking(false);
    }
  };

  const closeLobbyHostConfirm = useCallback(() => {
    setLobbyHostConfirmOpen(false);
    setPendingLobbyHostChange(null);
  }, []);

  const submitLobbyHostChange = useCallback(
    async (placement: PlayerPlacement) => {
      try {
        setWorking(true);
        await setLobbyHost(placement.eventId, placement.userId);
        await loadGroup();
        closeLobbyHostConfirm();
      } catch (err) {
        setPageError(extractApiError(err, "Could not complete that action."));
      } finally {
        setWorking(false);
      }
    },
    [closeLobbyHostConfirm, loadGroup, setPageError, setWorking],
  );

  const handleMakeLobbyHost = useCallback(
    (placement: PlayerPlacement) => {
      if (!group) return;
      const event = group.events.find((item) => item.id === placement.eventId);
      if (!event || !isTeamAssignedPlacement(placement)) return;

      const lobby = (event.lobbies ?? []).find((item) => item.id === placement.lobbyId);
      if (!lobby) return;
      const lobbyIndex = (event.lobbies ?? []).findIndex((item) => item.id === placement.lobbyId);

      const player = lobby.teams
        .flatMap((team) => team.players)
        .find((item) => item.user_id === placement.userId);
      if (!player) return;

      if (player.can_lobby_host) {
        void submitLobbyHostChange(placement);
        return;
      }

      setPendingLobbyHostChange({
        placement,
        volunteerOptions: buildLobbyHostVolunteers(
          lobby,
          lobbyIndex >= 0 ? lobbyIndex : 0,
          placement.userId,
          lobby.host_id,
        ),
      });
      setLobbyHostConfirmOpen(true);
    },
    [group, submitLobbyHostChange],
  );

  const swapEvent = useMemo(() => {
    if (!group || !pendingSwap) return null;
    return group.events.find((event) => event.id === pendingSwap.eventId) ?? null;
  }, [group, pendingSwap]);

  const swapCandidateOptions = useMemo(() => {
    if (!swapEvent || !pendingSwap) return [];
    return buildSwapCandidates(swapEvent, pendingSwap);
  }, [swapEvent, pendingSwap]);

  const moveToSubsEvent = useMemo(() => {
    if (!group || !pendingMoveToSubs) return null;
    return group.events.find((event) => event.id === pendingMoveToSubs.eventId) ?? null;
  }, [group, pendingMoveToSubs]);

  const moveToSubsLobbyOptions = useMemo((): SelectOption[] => {
    if (!moveToSubsEvent) return [];
    return (moveToSubsEvent.lobbies ?? []).map((lobby, lobbyIndex) => ({
      value: lobby.id,
      label: `Lobby ${lobbyIndex + 1}`,
    }));
  }, [moveToSubsEvent]);

  return {
    warningSheetOpen,
    setWarningSheetOpen,
    subCapacitySheetOpen,
    setSubCapacitySheetOpen,
    lockInTeams,
    confirmDeleteTeams,
    swapSheetOpen,
    pendingSwap,
    swapTargetUserId,
    setSwapTargetUserId,
    swapError,
    setSwapError,
    openSwapSheet,
    closeSwapSheet,
    handleSwapSubmit,
    swapCandidateOptions,
    moveToSubsSheetOpen,
    pendingMoveToSubs,
    moveToSubsLobbyId,
    setMoveToSubsLobbyId,
    moveToSubsError,
    setMoveToSubsError,
    closeMoveToSubsSheet,
    handleMoveToUnplaced,
    handleMoveToSubs,
    handleMoveToSubsSubmit,
    moveToSubsLobbyOptions,
    lobbyHostConfirmOpen,
    pendingLobbyHostChange,
    closeLobbyHostConfirm,
    submitLobbyHostChange,
    handleMakeLobbyHost,
  };
}
