"use client";

// Delete-registration confirmation sheet: single-game vs. all-games-in-series deletion.
import { useState } from "react";
import { formatUserDisplayLabel } from "@/app/_lib/userDisplayName";
import { deleteRegistration } from "@/app/_services/events";
import { EventGroupDetail, EventRegistration, User } from "@/app/_types/types";
import { PendingDeleteAction } from "../_types";

export function useDeleteRegistration(
  group: EventGroupDetail | null,
  user: User | null,
  loadGroup: () => Promise<void>,
  setWorking: (working: boolean) => void,
  setPageError: (message: string | null) => void,
) {
  const [deleteWarningSheetOpen, setDeleteWarningSheetOpen] = useState(false);
  const [pendingDeleteAction, setPendingDeleteAction] = useState<PendingDeleteAction | null>(null);

  const deleteAllRegistrationsForUserInGroup = async (targetUserId: string) => {
    if (!group) return;
    const eventIds = group.events
      .filter((event) => event.registrations.some((item) => item.user_id === targetUserId))
      .map((event) => event.id);
    await Promise.all(eventIds.map((eventId) => deleteRegistration(eventId, targetUserId)));
  };

  const openDeleteConfirmation = (
    registration: EventRegistration,
    gameNumber: number,
    mode: "single" | "all",
  ) => {
    if (!group) return;
    const registrationsInGroup = group.events.reduce(
      (count, event) => count + (event.registrations.some((item) => item.user_id === registration.user_id) ? 1 : 0),
      0,
    );
    setPendingDeleteAction({
      mode,
      userId: registration.user_id,
      userName: formatUserDisplayLabel(registration.display_name, registration.discord_name),
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

  /** Used when saving a registration with zero games selected while already registered — treated as "delete all". */
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

  return {
    deleteWarningSheetOpen,
    pendingDeleteAction,
    openDeleteConfirmation,
    closeDeleteWarningSheet,
    openDeleteAllForCurrentUserConfirmation,
    handleDeleteRegistration,
    deletingSelf,
  };
}
