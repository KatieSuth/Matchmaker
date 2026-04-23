// Client calls for event groups, registration, and host actions (uses axios + auth interceptors).
import api from "@/app/_lib/axios";
import { EventGroupDetail } from "@/app/_types/types";

export interface CreateEventRequest {
  game_mode_id: string;
  region: string;
  start_time: string;
  sub_min: number;
  games_to_run: number;
  registration_open: boolean;
}

export interface CreateEventResponse {
  group_id: string;
}

export async function createEvent(payload: CreateEventRequest): Promise<CreateEventResponse> {
  const res = await api.post<CreateEventResponse>("/events", payload);
  return res.data;
}

export interface UpdateEventGroupRequest {
  region: string;
  sub_min: number;
}

export interface UpdateRegistrationRequest {
  can_substitute: boolean;
  can_lobby_host: boolean;
  duo_request: string;
}

export async function fetchEventGroup(groupId: string): Promise<EventGroupDetail> {
  const res = await api.get<EventGroupDetail>(`/events/${groupId}`);
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

export async function deleteTeamsAndOpenRegistration(groupId: string): Promise<void> {
  await api.delete(`/events/${groupId}/teams`);
}

export async function deleteEventGroup(groupId: string): Promise<void> {
  await api.delete(`/events/${groupId}`);
}

export async function upsertMyRegistration(eventId: string, payload: UpdateRegistrationRequest): Promise<void> {
  await api.put(`/registrations/${eventId}/me`, payload);
}

export async function deleteRegistration(eventId: string, userId?: string): Promise<void> {
  if (userId) {
    await api.delete(`/registrations/${eventId}/${userId}`);
    return;
  }
  await api.delete(`/registrations/${eventId}/me`);
}
