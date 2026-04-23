// Game and game-mode fetches, plus a small helper to surface Gin JSON error `message` fields.
import api from "@/app/_lib/axios";
import { Game, GameMode } from "@/app/_types/types";

export async function fetchGames(): Promise<Game[]> {
  const res = await api.get<Game[]>("/games");
  return res.data;
}

export async function fetchGamesForUser(ownerId: string): Promise<Game[]> {
  const res = await api.get<Game[]>(`/games/users/${ownerId}`);
  return res.data;
}

export async function fetchGameModes(gameId: string): Promise<GameMode[]> {
  const res = await api.get<GameMode[]>(`/games/${gameId}/modes`);
  return res.data;
}

// Extracts the `message` field from a Gin error response body, falling back
// to a generic message. Use this in catch blocks for any mutating API call.
export function extractApiError(err: unknown, fallback = "Something went wrong. Please try again."): string {
  return (
    (err as { response?: { data?: { message?: string } } })?.response?.data?.message ?? fallback
  );
}
