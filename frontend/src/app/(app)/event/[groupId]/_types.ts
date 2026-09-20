// Shared local types for the event group detail page and its subcomponents/hooks.
import { EventLobby } from "@/app/_types/types";

/** Per-event registration options drafted in the register/edit-registration sheet. */
export interface EventRegistrationDraft {
  can_substitute: boolean;
  can_lobby_host: boolean;
}

/** Fresh per-event settings when the user is newly registering (or newly selecting a game). */
export const DEFAULT_EVENT_REGISTRATION_DRAFT: EventRegistrationDraft = {
  can_substitute: true,
  can_lobby_host: true,
};

export interface RegistrationDraft {
  selected_event_ids: string[];
  per_event: Record<string, EventRegistrationDraft>;
  duo_request: string;
}

export interface PendingDeleteAction {
  mode: "single" | "all";
  userId: string;
  userName: string;
  eventId: string;
  gameNumber: number;
  registrationsInGroup: number;
}

/** Identifies a roster/sub/unplaced player within a locked-in game for host swap actions. */
export interface PlayerPlacement {
  eventId: string;
  userId: string;
  discordName: string;
  lobbyId: string | null;
  sourceLobbyIndex: number | null;
  teamNumber: number | null | undefined;
}

export interface PendingJoinLobby {
  lobby: EventLobby;
  lobbyIndex: number;
  gameNumber: number;
  startTime: string;
}

export interface LobbyHostVolunteer {
  userId: string;
  discordName: string;
  teamNumber: number;
  isCurrentHost: boolean;
}

export interface PendingLobbyHostChange {
  placement: PlayerPlacement;
  volunteerOptions: LobbyHostVolunteer[];
}
