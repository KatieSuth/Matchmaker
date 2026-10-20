"use client";

// Pre-lock-in registration list for a single game (flat list, no teams/lobbies yet). Absent marks
// survive team deletion, so they still show here after an unlock.
import { EventGroupEvent, EventRegistration } from "@/app/_types/types";
import { PlayerCard } from "./PlayerCard";

export function EventPanel({
  event,
  gameNumber,
  eventRegion,
  currentUserRegion,
  isHostView,
  currentUserId,
  onShowDetails,
  onDeleteRegistrationForGame,
  onDeleteAllFromUser,
  excessiveNoShows,
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
  /** Host-only: user ID → recent no-show count for players at/above the deprioritization limit. */
  excessiveNoShows?: Record<string, number>;
}) {
  const absentUserIds = new Set(event.absent_user_ids ?? []);
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
            allowRegistrationDelete={event.lobbies_count === 0}
            onShowDetails={onShowDetails}
            onDeleteRegistrationForGame={onDeleteRegistrationForGame}
            onDeleteAllFromUser={onDeleteAllFromUser}
            isAbsent={absentUserIds.has(registration.user_id)}
            recentNoShowCount={excessiveNoShows?.[registration.user_id] ?? 0}
          />
        ))
      )}
    </div>
  );
}
