// MSW handlers mirroring `@/app/_services/games.ts`.
import { http, HttpResponse } from "msw";
import { buildGame, buildGameMode, buildGameRank } from "@/test/fixtures";
import { TEST_API_URL } from "@/test/constants";

export const gamesHandlers = [
  http.get(`${TEST_API_URL}/games`, () => HttpResponse.json([buildGame()])),
  http.get(`${TEST_API_URL}/games/users/:ownerId`, () => HttpResponse.json([buildGame()])),
  http.get(`${TEST_API_URL}/games/:gameId/modes`, () => HttpResponse.json([buildGameMode()])),
  http.get(`${TEST_API_URL}/games/:gameId/ranks`, () =>
    HttpResponse.json([
      buildGameRank({ name: "Bronze", order: 1 }),
      buildGameRank({ name: "Silver", order: 2 }),
      buildGameRank({ name: "Gold", order: 3 }),
    ]),
  ),
];
