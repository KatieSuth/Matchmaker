// Client calls for event groups, registration, and host actions (uses axios + auth interceptors).
import api from "@/app/_lib/axios";
import {
  EventGroupDetail,
  EventsPage,
  EventSortLogic,
  UpsertGroupRegistrationRequest,
} from "@/app/_types/types";

export interface CreateEventRequest {
  game_mode_id: string;
  region: string;
  start_time: string;
  sub_min: number;
  games_to_run: number;
  registration_open: boolean;
  sort_logic: EventSortLogic;
  name?: string;
}

export interface CreateEventResponse {
  group_id: string;
}

export async function createEvent(payload: CreateEventRequest): Promise<CreateEventResponse> {
  const res = await api.post<CreateEventResponse>("/events", payload);
  return res.data;
}

export interface PatchGroupEventScheduleItem {
  event_id: string;
  start_time: string;
  game_mode_id: string;
}

export interface UpdateEventGroupRequest {
  region: string;
  sub_min: number;
  sort_logic: EventSortLogic;
  registration_open: boolean;
  name?: string;
  events: PatchGroupEventScheduleItem[];
}

export interface UpdateRegistrationRequest {
  can_substitute: boolean;
  can_lobby_host: boolean;
  duo_request: string;
}

export async function fetchEventGroup(groupId: string, signal?: AbortSignal): Promise<EventGroupDetail> {
  const res = await api.get<EventGroupDetail>(`/events/${groupId}`, signal ? { signal } : undefined);
  return res.data;
}

export async function updateEventGroup(groupId: string, payload: UpdateEventGroupRequest): Promise<void> {
  await api.patch(`/events/${groupId}`, payload);
}

export async function setEventGroupRegistrationOpen(groupId: string, registrationOpen: boolean): Promise<void> {
  await api.patch(`/events/${groupId}/registration`, { registration_open: registrationOpen });
}

export async function createTeams(groupId: string): Promise<void> {
  await api.post(`/events/${groupId}/teams`);
}

export async function deleteTeams(groupId: string): Promise<void> {
  await api.delete(`/events/${groupId}/teams`);
}

/** Host-only. Exchanges two players' placements for one locked-in game. */
export async function swapPlayers(eventId: string, userIdA: string, userIdB: string): Promise<void> {
  await api.post(`/registrations/${eventId}/player-swap`, { user_id_a: userIdA, user_id_b: userIdB });
}

/** Host-only. Assigns a team player as the lobby host for their lobby. */
export async function setLobbyHost(eventId: string, userId: string): Promise<void> {
  await api.post(`/registrations/${eventId}/lobby-host`, { user_id: userId });
}

/** Host-only. Removes a substitute from their lobby sub pool. */
export async function moveSubToUnplaced(eventId: string, userId: string): Promise<void> {
  await api.post(`/registrations/${eventId}/sub-to-unplaced`, { user_id: userId });
}

/** Host-only. Adds an unplaced substitute-eligible player to a lobby sub pool. */
export async function moveUnplacedToSubs(eventId: string, userId: string, lobbyId: string): Promise<void> {
  await api.post(`/registrations/${eventId}/unplaced-to-subs`, { user_id: userId, lobby_id: lobbyId });
}

export async function deleteEventGroup(groupId: string): Promise<void> {
  await api.delete(`/events/${groupId}`);
}

export async function upsertMyRegistration(eventId: string, payload: UpdateRegistrationRequest): Promise<void> {
  await api.put(`/registrations/${eventId}/me`, payload);
}

export async function upsertMyGroupRegistrations(groupId: string, payload: UpsertGroupRegistrationRequest): Promise<void> {
  await api.put(`/registrations/group/${groupId}/me`, payload);
}

export async function deleteRegistration(eventId: string, userId?: string): Promise<void> {
  if (userId) {
    await api.delete(`/registrations/${eventId}/${userId}`);
    return;
  }
  await api.delete(`/registrations/${eventId}/me`);
}

export async function fetchMyEvents(params: Record<string, string | undefined>, signal?: AbortSignal): Promise<EventsPage> {
  const query = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== "") {
      query.set(key, value);
    }
  }

  const res = await api.get<EventsPage>(`/users/me/events?${query.toString()}`, signal ? { signal } : undefined);
  return res.data;
}
