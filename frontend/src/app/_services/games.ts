// Game and game-mode fetches, plus a small helper to surface Gin JSON error `message` fields.
import api from "@/app/_lib/axios";
import { Game, GameMode, GameRank } from "@/app/_types/types";

export async function fetchGames(signal?: AbortSignal): Promise<Game[]> {
  const res = await api.get<Game[]>("/games", signal ? { signal } : undefined);
  return res.data;
}

export async function fetchGamesForUser(ownerId: string, signal?: AbortSignal): Promise<Game[]> {
  const res = await api.get<Game[]>(`/games/users/${ownerId}`, signal ? { signal } : undefined);
  return res.data;
}

export async function fetchGameModes(gameId: string, signal?: AbortSignal): Promise<GameMode[]> {
  const res = await api.get<GameMode[]>(`/games/${gameId}/modes`, signal ? { signal } : undefined);
  return res.data;
}

export async function fetchGameRanks(gameId: string, signal?: AbortSignal): Promise<GameRank[]> {
  const res = await api.get<GameRank[]>(`/games/${gameId}/ranks`, signal ? { signal } : undefined);
  return res.data.sort((a, b) => a.order - b.order);
}

// Extracts the `message` field from a Gin error response body, falling back
// to a generic message. Use this in catch blocks for any mutating API call.
export function extractApiError(err: unknown, fallback = "Something went wrong. Please try again."): string {
  return (
    (err as { response?: { data?: { message?: string } } })?.response?.data?.message ?? fallback
  );
}
