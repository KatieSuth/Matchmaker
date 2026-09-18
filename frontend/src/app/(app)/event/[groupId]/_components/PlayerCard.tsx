"use client";

// Registration card shown in the registration list and in each roster/subs/unplaced section of
// TeamsPanel. Owns the per-player ellipsis menu (show details, swap, move, make host, delete).
import { EllipsisMenu, EllipsisMenuOption } from "@/app/_components/EllipsisMenu";
import { EMPTY_VALUE } from "@/app/_lib/constants";
import { formatUserDisplayLabel } from "@/app/_lib/userDisplayName";
import { EventRegistration } from "@/app/_types/types";
import { isSubPlacement, isTeamAssignedPlacement, isUnplacedPlacement } from "../_lib/placements";
import { PlayerPlacement } from "../_types";

export function PlayerCard({
  registration,
  gameNumber,
  eventRegion,
  currentUserRegion,
  isHostView,
  currentUserId,
  canEditRegistration,
  allowRegistrationDelete = true,
  onShowDetails,
  onDeleteRegistrationForGame,
  onDeleteAllFromUser,
  placement,
  onSwap,
  onMoveToUnplaced,
  onMoveToSubs,
  onMakeLobbyHost,
  lobbyHostId,
  showDuoRequest = false,
}: {
  registration: EventRegistration;
  gameNumber: number;
  eventRegion: string;
  currentUserRegion?: string | null;
  isHostView: boolean;
  currentUserId?: string;
  canEditRegistration: boolean;
  allowRegistrationDelete?: boolean;
  onShowDetails: (registration: EventRegistration) => void;
  onDeleteRegistrationForGame: (registration: EventRegistration, gameNumber: number) => void;
  onDeleteAllFromUser: (registration: EventRegistration, gameNumber: number) => void;
  placement?: PlayerPlacement;
  onSwap?: (placement: PlayerPlacement) => void;
  onMoveToUnplaced?: (placement: PlayerPlacement) => void;
  onMoveToSubs?: (placement: PlayerPlacement) => void;
  onMakeLobbyHost?: (placement: PlayerPlacement) => void;
  lobbyHostId?: string | null;
  showDuoRequest?: boolean;
}) {
  const canOpenMenu = isHostView || canEditRegistration;
  const canDelete =
    allowRegistrationDelete && (isHostView || registration.user_id === currentUserId);
  const regionMismatch =
    canEditRegistration &&
    !!currentUserRegion &&
    currentUserRegion.trim().toUpperCase() !== eventRegion.trim().toUpperCase();
  const menuOptions: EllipsisMenuOption[] = [];
  menuOptions.push({
    label: "Show More Details",
    onSelect: () => onShowDetails(registration),
  });
  if (
    isHostView &&
    placement &&
    (isTeamAssignedPlacement(placement) || isSubPlacement(placement) || isUnplacedPlacement(placement)) &&
    onSwap
  ) {
    menuOptions.push({
      label: "Swap",
      onSelect: () => onSwap(placement),
    });
  }
  if (isHostView && placement && isSubPlacement(placement) && onMoveToUnplaced) {
    menuOptions.push({
      label: "Move to Unplaced",
      onSelect: () => onMoveToUnplaced(placement),
    });
  }
  if (
    isHostView &&
    placement &&
    isUnplacedPlacement(placement) &&
    registration.can_substitute &&
    onMoveToSubs
  ) {
    menuOptions.push({
      label: "Move to Substitutes",
      onSelect: () => onMoveToSubs(placement),
    });
  }
  if (
    isHostView &&
    isTeamAssignedPlacement(placement) &&
    onMakeLobbyHost &&
    placement.userId !== lobbyHostId
  ) {
    menuOptions.push({
      label: "Make Lobby Host",
      onSelect: () => onMakeLobbyHost(placement),
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
            {formatUserDisplayLabel(registration.display_name, registration.discord_name)}
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
          <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">In-game name</p>
          <p className="text-xs text-[var(--color-text-soft)] truncate" title={registration.in_game_name || undefined}>
            {registration.in_game_name?.trim() || EMPTY_VALUE}
          </p>
        </div>
        <div>
          <p className="text-[0.65rem] uppercase tracking-wide text-[var(--color-text-faint)]">Average Rank</p>
          <p className="text-xs text-[var(--color-text-soft)]">{registration.avg_rank_name || EMPTY_VALUE}</p>
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
