"use client";

// Post-lock-in teams view for a single game: per-lobby team rosters, substitutes pool, and
// (host-only) unplaced players, plus the fairness-warning banner and "Join Lobby" link.
import { EMPTY_VALUE, NO_SUBSTITUTES_MESSAGE } from "@/app/_lib/constants";
import { sequentialTeamNumber } from "@/app/_lib/discordPings";
import { formatUserDisplayLabel } from "@/app/_lib/userDisplayName";
import { EventGroupEvent, EventLobby, EventRegistration, GameRank } from "@/app/_types/types";
import { formatPlayerCount } from "../_lib/formatters";
import { lobbyFairnessWarningMessage } from "../_lib/lobbyFairness";
import { lobbyHostName, lobbyPlayerAsRegistration } from "../_lib/placements";
import { teamAverageRankLabel } from "../_lib/ranks";
import { PlayerPlacement } from "../_types";
import { PlayerCard } from "./PlayerCard";

export function TeamsPanel({
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
  onMoveToUnplaced,
  onMoveToSubs,
  onMakeLobbyHost,
  onJoinLobby,
  showJoinLobby,
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
  onMoveToUnplaced?: (placement: PlayerPlacement) => void;
  onMoveToSubs?: (placement: PlayerPlacement) => void;
  onMakeLobbyHost?: (placement: PlayerPlacement) => void;
  onJoinLobby: (lobby: EventLobby, lobbyIndex: number, gameNumber: number, startTime: string) => void;
  showJoinLobby: boolean;
}) {
  const lobbies = event.lobbies ?? [];
  if (lobbies.length === 0) {
    return null;
  }

  const substituteEntries = lobbies.flatMap((lobby, lobbyIndex) =>
    lobby.subs.map((player) => ({ lobby, lobbyIndex, player })),
  );

  return (
    <div className="flex flex-col gap-4">
      {lobbies.map((lobby, lobbyIndex) => (
        <div key={lobby.id} className="flex flex-col gap-3">
          {lobby.fairness_warning && (
            <div className="rounded-lg border border-amber-400/30 bg-amber-400/10 px-3 py-2 text-xs text-amber-200">
              {lobbyFairnessWarningMessage(lobby)}
            </div>
          )}
          <div className="flex flex-wrap items-center gap-x-2 gap-y-0.5 text-[0.9375rem] font-semibold text-[var(--color-text)]">
            <span className="min-w-0 break-words">
              Lobby {lobbyIndex + 1}
              {lobbyHostName(lobby, event) ? ` · Host: ${lobbyHostName(lobby, event)}` : ""}
              {showJoinLobby && (
                <>
                  {" · "}
                  <button
                    type="button"
                    onClick={() => onJoinLobby(lobby, lobbyIndex, gameNumber, event.start_time)}
                    className="font-medium text-[var(--color-accent-blue)] hover:underline"
                  >
                    Join Lobby
                  </button>
                </>
              )}
            </span>
            {lobby.fairness_warning && <span className="text-amber-400" aria-label="Unfair lobby">⚠</span>}
          </div>
          <div className="grid gap-3 sm:grid-cols-2">
            {lobby.teams.map((team) => {
              const averageRank = isHostView ? teamAverageRankLabel(team.players, gameRanks) : EMPTY_VALUE;
              const displayTeamNumber = sequentialTeamNumber(lobbyIndex, team.team_number);
              return (
              <div key={team.team_number} className="flex flex-col gap-2">
                <p className="text-xs font-medium uppercase tracking-wide text-[var(--color-text-muted)]">
                  Team {displayTeamNumber}
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
                    allowRegistrationDelete={false}
                    onShowDetails={onShowDetails}
                    onDeleteRegistrationForGame={onDeleteRegistrationForGame}
                    onDeleteAllFromUser={onDeleteAllFromUser}
                    showDuoRequest
                    placement={{
                      eventId: event.id,
                      userId: player.user_id,
                      discordName: formatUserDisplayLabel(player.display_name, player.discord_name),
                      lobbyId: lobby.id,
                      sourceLobbyIndex: lobbyIndex,
                      teamNumber: team.team_number,
                    }}
                    lobbyHostId={lobby.host_id}
                    onSwap={onSwapPlayer}
                    onMoveToUnplaced={onMoveToUnplaced}
                    onMoveToSubs={onMoveToSubs}
                    onMakeLobbyHost={onMakeLobbyHost}
                  />
                ))}
              </div>
            );
            })}
          </div>
        </div>
      ))}
      <div className="flex flex-col gap-2">
        <p className="text-xs font-medium uppercase tracking-wide text-[var(--color-text-muted)]">
          Substitutes
          {substituteEntries.length > 0
            ? ` · ${formatPlayerCount(substituteEntries.length)}`
            : ""}
        </p>
        {substituteEntries.length === 0 ? (
          <div className="rounded-xl border border-dashed border-white/[0.08] py-8 text-center text-sm text-[var(--color-text-muted)]">
            {NO_SUBSTITUTES_MESSAGE}
          </div>
        ) : (
          substituteEntries.map(({ lobby, lobbyIndex, player }) => (
            <PlayerCard
              key={player.user_id}
              registration={lobbyPlayerAsRegistration(player, event.id)}
              gameNumber={gameNumber}
              eventRegion={eventRegion}
              currentUserRegion={currentUserRegion}
              isHostView={isHostView}
              currentUserId={currentUserId}
              canEditRegistration={false}
              allowRegistrationDelete={false}
              onShowDetails={onShowDetails}
              onDeleteRegistrationForGame={onDeleteRegistrationForGame}
              onDeleteAllFromUser={onDeleteAllFromUser}
              showDuoRequest
              placement={{
                eventId: event.id,
                userId: player.user_id,
                discordName: formatUserDisplayLabel(player.display_name, player.discord_name),
                lobbyId: lobby.id,
                sourceLobbyIndex: lobbyIndex,
                teamNumber: null,
              }}
              onSwap={onSwapPlayer}
              onMoveToUnplaced={onMoveToUnplaced}
              onMoveToSubs={onMoveToSubs}
            />
          ))
        )}
      </div>
      {isHostView && (event.unplaced ?? []).length > 0 && (
        <div className="flex flex-col gap-2">
          <p className="text-xs font-medium uppercase tracking-wide text-[var(--color-text-muted)]">
            Unplaced · {formatPlayerCount((event.unplaced ?? []).length)}
          </p>
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
              allowRegistrationDelete={false}
              onShowDetails={onShowDetails}
              onDeleteRegistrationForGame={onDeleteRegistrationForGame}
              onDeleteAllFromUser={onDeleteAllFromUser}
              showDuoRequest
              placement={{
                eventId: event.id,
                userId: registration.user_id,
                discordName: formatUserDisplayLabel(registration.display_name, registration.discord_name),
                lobbyId: null,
                sourceLobbyIndex: null,
                teamNumber: undefined,
              }}
              onSwap={onSwapPlayer}
              onMoveToUnplaced={onMoveToUnplaced}
              onMoveToSubs={onMoveToSubs}
            />
          ))}
        </div>
      )}
    </div>
  );
}
