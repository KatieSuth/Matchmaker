import { LobbyHostResponsibilitiesList } from "@/app/_components/LobbyHostResponsibilities";

export interface LobbyHostAssignment {
  gameNumber: number;
  lobbyNumber: number;
}

export function LobbyHostAssignmentBanner({ assignments }: { assignments: LobbyHostAssignment[] }) {
  if (assignments.length === 0) {
    return null;
  }

  return (
    <div className="rounded-lg border border-[var(--color-accent-blue)]/30 bg-[var(--color-accent-blue)]/10 px-4 py-3">
      <p className="text-sm font-medium text-[var(--color-accent-blue)]">
        {assignments.length === 1
          ? `You are the lobby host for Lobby ${assignments[0].lobbyNumber} in Game ${assignments[0].gameNumber}.`
          : "You are the lobby host for the following:"}
      </p>
      {assignments.length > 1 && (
        <ul className="mt-2 list-disc space-y-1 pl-5 text-sm text-[var(--color-text-soft)]">
          {assignments.map((assignment) => (
            <li key={`game-${assignment.gameNumber}-lobby-${assignment.lobbyNumber}`}>
              Lobby {assignment.lobbyNumber} in Game {assignment.gameNumber}
            </li>
          ))}
        </ul>
      )}
      <p className="mt-3 text-sm font-medium text-[var(--color-accent-blue)]">Your responsibilities:</p>
      <LobbyHostResponsibilitiesList className="mt-1.5 text-sm text-[var(--color-text-soft)]" />
    </div>
  );
}
