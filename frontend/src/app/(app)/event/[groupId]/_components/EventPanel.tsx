"use client";

// Pre-lock-in registration list for a single game (flat list, no teams/lobbies yet).
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
            allowRegistrationDelete={event.lobbies_count === 0}
            onShowDetails={onShowDetails}
            onDeleteRegistrationForGame={onDeleteRegistrationForGame}
            onDeleteAllFromUser={onDeleteAllFromUser}
          />
        ))
      )}
    </div>
  );
}
