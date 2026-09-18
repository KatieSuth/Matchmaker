"use client";

// "Join Lobby" sheet: shows/edits the lobby's join code or invite link. Editable by the event
// host or that lobby's host; read-only (copy-only) for everyone else.
import { useCallback, useState } from "react";
import { useCopyStatus } from "@/app/_hooks/useCopyStatus";
import { buildLobbyJoinDisplayValue, validateLobbyJoinInput } from "@/app/_lib/lobbyJoin";
import { updateLobbyJoinCode } from "@/app/_services/events";
import { extractApiError } from "@/app/_services/games";
import { EventGroupDetail, EventLobby, User } from "@/app/_types/types";
import { PendingJoinLobby } from "../_types";

export function useJoinLobbySheet(
  group: EventGroupDetail | null,
  user: User | null,
  isHost: boolean,
  loadGroup: () => Promise<void>,
  setWorking: (working: boolean) => void,
  setToast: (message: string | null) => void,
) {
  const [joinLobbySheetOpen, setJoinLobbySheetOpen] = useState(false);
  const [pendingJoinLobby, setPendingJoinLobby] = useState<PendingJoinLobby | null>(null);
  const [joinLobbyDraft, setJoinLobbyDraft] = useState("");
  const [joinLobbyError, setJoinLobbyError] = useState<string | null>(null);
  const { status: joinLobbyCopyStatus, copy: copyJoinLobbyValue, reset: resetJoinLobbyCopyStatus } = useCopyStatus();

  const openJoinLobbySheet = useCallback(
    (lobby: EventLobby, lobbyIndex: number, gameNumber: number, startTime: string) => {
      const display = buildLobbyJoinDisplayValue(lobby.join_code, group?.join_link_base);
      setPendingJoinLobby({ lobby, lobbyIndex, gameNumber, startTime });
      setJoinLobbyDraft(display?.value ?? lobby.join_code ?? "");
      setJoinLobbyError(null);
      resetJoinLobbyCopyStatus();
      setJoinLobbySheetOpen(true);
    },
    [group?.join_link_base, resetJoinLobbyCopyStatus],
  );

  const closeJoinLobbySheet = useCallback(() => {
    setJoinLobbySheetOpen(false);
    setPendingJoinLobby(null);
    setJoinLobbyDraft("");
    setJoinLobbyError(null);
    resetJoinLobbyCopyStatus();
  }, [resetJoinLobbyCopyStatus]);

  const handleJoinLobbyDraftChange = useCallback((value: string) => {
    setJoinLobbyDraft(value);
    setJoinLobbyError(null);
  }, []);

  const canEditPendingJoinLobby = !!(
    pendingJoinLobby &&
    user?.id &&
    (isHost || pendingJoinLobby.lobby.host_id === user.id)
  );

  const pendingJoinDisplay = pendingJoinLobby
    ? buildLobbyJoinDisplayValue(pendingJoinLobby.lobby.join_code, group?.join_link_base)
    : null;

  const handleCopyJoinLobby = async () => {
    const value = canEditPendingJoinLobby ? joinLobbyDraft.trim() : pendingJoinDisplay?.value;
    await copyJoinLobbyValue(value);
  };

  const handleSaveJoinLobby = async () => {
    if (!pendingJoinLobby || !canEditPendingJoinLobby) return;
    const validationError = validateLobbyJoinInput(joinLobbyDraft, group?.join_link_base ?? null);
    if (validationError) {
      setJoinLobbyError(validationError);
      return;
    }
    try {
      setWorking(true);
      setJoinLobbyError(null);
      const trimmed = joinLobbyDraft.trim();
      await updateLobbyJoinCode(pendingJoinLobby.lobby.id, trimmed === "" ? null : trimmed);
      await loadGroup();
      closeJoinLobbySheet();
      setToast("Lobby join info saved.");
    } catch (err) {
      setJoinLobbyError(extractApiError(err, "Could not save lobby join info."));
    } finally {
      setWorking(false);
    }
  };

  return {
    joinLobbySheetOpen,
    pendingJoinLobby,
    joinLobbyDraft,
    joinLobbyError,
    joinLobbyCopyStatus,
    openJoinLobbySheet,
    closeJoinLobbySheet,
    handleJoinLobbyDraftChange,
    canEditPendingJoinLobby,
    pendingJoinDisplay,
    handleCopyJoinLobby,
    handleSaveJoinLobby,
  };
}
