// MSW handlers mirroring `@/app/_services/events.ts`.
import { http, HttpResponse } from "msw";
import { buildCreateTeamsResponse, buildEventGroupDetail, buildEventsPage } from "@/test/fixtures";
import { TEST_API_URL } from "@/test/constants";

export const eventsHandlers = [
  http.post(`${TEST_API_URL}/events`, () => HttpResponse.json({ group_id: "group-1" })),
  http.get(`${TEST_API_URL}/events/:groupId`, () => HttpResponse.json(buildEventGroupDetail())),
  http.get(`${TEST_API_URL}/events/:groupId/access`, () => new HttpResponse(null, { status: 204 })),
  http.patch(`${TEST_API_URL}/events/:groupId`, () => new HttpResponse(null, { status: 204 })),
  http.patch(`${TEST_API_URL}/events/:groupId/registration`, () => new HttpResponse(null, { status: 204 })),
  http.post(`${TEST_API_URL}/events/:groupId/teams`, () => HttpResponse.json(buildCreateTeamsResponse())),
  http.delete(`${TEST_API_URL}/events/:groupId/teams`, () => new HttpResponse(null, { status: 204 })),
  http.delete(`${TEST_API_URL}/events/:groupId`, () => new HttpResponse(null, { status: 204 })),
  http.post(`${TEST_API_URL}/registrations/:eventId/player-swap`, () => new HttpResponse(null, { status: 204 })),
  http.post(`${TEST_API_URL}/registrations/:eventId/lobby-host`, () => new HttpResponse(null, { status: 204 })),
  http.post(`${TEST_API_URL}/registrations/:eventId/sub-to-unplaced`, () => new HttpResponse(null, { status: 204 })),
  http.post(`${TEST_API_URL}/registrations/:eventId/unplaced-to-subs`, () => new HttpResponse(null, { status: 204 })),
  http.patch(`${TEST_API_URL}/lobbies/:lobbyId/join-code`, () => new HttpResponse(null, { status: 204 })),
  http.put(`${TEST_API_URL}/registrations/:eventId/me`, () => new HttpResponse(null, { status: 204 })),
  http.put(`${TEST_API_URL}/registrations/group/:groupId/me`, () => new HttpResponse(null, { status: 204 })),
  http.delete(`${TEST_API_URL}/registrations/:eventId/me`, () => new HttpResponse(null, { status: 204 })),
  http.delete(`${TEST_API_URL}/registrations/:eventId/:userId`, () => new HttpResponse(null, { status: 204 })),
  http.get(`${TEST_API_URL}/users/me/events`, () => HttpResponse.json(buildEventsPage())),
];
