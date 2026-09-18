// MSW handlers mirroring `@/app/_services/auth.ts`.
import { http, HttpResponse } from "msw";
import { buildCompleteAuthResponse } from "@/test/fixtures";
import { TEST_API_URL } from "@/test/constants";

export const authHandlers = [
  http.post(`${TEST_API_URL}/auth/complete`, () => HttpResponse.json(buildCompleteAuthResponse())),
  http.post(`${TEST_API_URL}/auth/logout`, () => new HttpResponse(null, { status: 204 })),
  // Called directly via a bare `axios.post` (not the shared `api` instance) by
  // `_lib/auth.ts#refreshAccessToken` — see that file for why it can't use an AbortSignal.
  http.post(`${TEST_API_URL}/auth/refresh`, () => HttpResponse.json(buildCompleteAuthResponse())),
];
