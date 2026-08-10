import api from "@/app/_lib/axios";
import { User, UserGame } from "@/app/_types/types";

export interface UpdateUserPreferencesPayload {
  display_name: string | null;
  pronouns: string | null;
  show_pronouns: boolean;
  region: string | null;
  games?: {
    game_id: string;
    in_game_name: string;
    current_rank: string | null;
    peak_rank: string | null;
    show_rank: boolean;
  }[];
}

export interface UpsertUserGamePayload {
  in_game_name: string;
  current_rank: string | null;
  peak_rank: string | null;
  show_rank: boolean;
}

export async function fetchCurrentUser(signal?: AbortSignal): Promise<User> {
  const res = await api.get<User>("/users/me", signal ? { signal } : undefined);
  return res.data;
}

export async function fetchCurrentUserGames(signal?: AbortSignal): Promise<UserGame[]> {
  const res = await api.get<UserGame[]>("/users/me/games", signal ? { signal } : undefined);
  return res.data;
}

export async function updateCurrentUserPreferences(payload: UpdateUserPreferencesPayload): Promise<void> {
  await api.put("/users/me", payload);
}

export async function upsertCurrentUserGame(gameId: string, payload: UpsertUserGamePayload): Promise<void> {
  await api.put(`/users/me/games/${gameId}`, payload);
}
