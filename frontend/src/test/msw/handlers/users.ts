// MSW handlers mirroring `@/app/_services/users.ts`.
import { http, HttpResponse } from "msw";
import { buildUser, buildUserGame } from "@/test/fixtures";
import { TEST_API_URL } from "@/test/constants";

export const usersHandlers = [
  http.get(`${TEST_API_URL}/users/me`, () => HttpResponse.json(buildUser())),
  http.get(`${TEST_API_URL}/users/me/games`, () => HttpResponse.json([buildUserGame()])),
  http.put(`${TEST_API_URL}/users/me`, () => new HttpResponse(null, { status: 204 })),
  http.put(`${TEST_API_URL}/users/me/games/:gameId`, () => new HttpResponse(null, { status: 204 })),
  http.delete(`${TEST_API_URL}/users/me/games/:gameId`, () => new HttpResponse(null, { status: 204 })),
  http.get(`${TEST_API_URL}/users/me/discord/guilds`, () => HttpResponse.json({ guilds: [] })),
];
