import api from "@/app/_lib/axios";
import { CompleteAuthResponse } from "@/app/_types/types";

export async function completeAuth(otc: string, signal?: AbortSignal): Promise<CompleteAuthResponse> {
  const response = await api.post<CompleteAuthResponse>(
    "/auth/complete",
    { otc },
    signal ? { signal } : undefined
  );
  return response.data;
}

export async function logoutUser(): Promise<void> {
  await api.post("/auth/logout");
}
