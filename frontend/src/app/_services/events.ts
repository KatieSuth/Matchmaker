import api from "@/app/_lib/axios";

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
