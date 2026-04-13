import api from "@/app/_lib/axios";
import { Game } from "@/app/_types/types";

export async function fetchGames(): Promise<Game[]> {
  const res = await api.get<Game[]>("/games");
  return res.data;
}

// Extracts the `message` field from a Gin error response body, falling back
// to a generic message. Use this in catch blocks for any mutating API call.
export function extractApiError(err: unknown, fallback = "Something went wrong. Please try again."): string {
  return (
    (err as { response?: { data?: { message?: string } } })?.response?.data?.message ?? fallback
  );
}
