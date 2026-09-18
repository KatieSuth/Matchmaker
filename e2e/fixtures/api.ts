// HTTP helpers for Node-side calls against the published API port (not Caddy).

import { e2eApiUrl, e2eBypassToken, testAuthBypassHeader } from "./env";

async function readError(res: Response): Promise<string> {
  const text = await res.text();
  return text || res.statusText;
}

async function assertOk(res: Response, action: string): Promise<Response> {
  if (!res.ok) {
    throw new Error(`${action} failed: ${res.status} ${await readError(res)}`);
  }
  return res;
}

export interface TestLoginResult {
  otc: string;
  new_user: boolean;
}

export interface TestUser {
  discordId: string;
  username: string;
  globalName?: string;
}

/** POSTs the test-only Discord bypass and returns a one-time code for /auth/callback. */
export async function issueTestLogin(user: TestUser): Promise<TestLoginResult> {
  const res = await fetch(`${e2eApiUrl}/auth/test_login`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      [testAuthBypassHeader]: e2eBypassToken,
    },
    body: JSON.stringify({
      discord_id: user.discordId,
      username: user.username,
      global_name: user.globalName,
    }),
  });
  await assertOk(res, "test_login");
  return (await res.json()) as TestLoginResult;
}

/** Exchanges an OTC for a JWT. Cookies from this Node fetch are unused; the browser uses /auth/callback. */
export async function completeAuthApi(otc: string): Promise<string> {
  const res = await fetch(`${e2eApiUrl}/auth/complete`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ otc }),
  });
  await assertOk(res, "auth complete");
  const body = (await res.json()) as { access_token?: string };
  if (!body.access_token) {
    throw new Error("auth complete returned no access_token");
  }
  return body.access_token;
}

function authHeaders(accessToken: string): HeadersInit {
  return {
    Authorization: `Bearer ${accessToken}`,
    "Content-Type": "application/json",
  };
}

/** Marks the account as no longer new_user so later logins land on /my_events. */
export async function saveProfileApi(accessToken: string, displayName: string): Promise<void> {
  const res = await fetch(`${e2eApiUrl}/users/me`, {
    method: "PUT",
    headers: authHeaders(accessToken),
    body: JSON.stringify({
      display_name: displayName,
      pronouns: "",
      show_pronouns: false,
      region: "AMER",
      games: [],
    }),
  });
  await assertOk(res, "save profile");
}

export interface CatalogGame {
  id: string;
  name: string;
}

export interface CatalogRank {
  id: string;
  name: string;
}

/** Loads the Valorant catalog row (seeded by goose) plus a usable rank UUID. */
export async function fetchValorantCatalog(accessToken: string): Promise<{ game: CatalogGame; rank: CatalogRank }> {
  const gamesRes = await fetch(`${e2eApiUrl}/games`, { headers: authHeaders(accessToken) });
  await assertOk(gamesRes, "list games");
  const games = (await gamesRes.json()) as CatalogGame[];
  const game = games.find((g) => g.name === "Valorant");
  if (!game) {
    throw new Error("Valorant is not in the system games catalog");
  }

  const ranksRes = await fetch(`${e2eApiUrl}/games/${game.id}/ranks`, { headers: authHeaders(accessToken) });
  await assertOk(ranksRes, "list ranks");
  const ranks = (await ranksRes.json()) as CatalogRank[];
  const rank = ranks.find((r) => r.name === "Iron 1") ?? ranks[0];
  if (!rank) {
    throw new Error("Valorant has no ranks");
  }
  return { game, rank };
}

/** Persists a Valorant profile so the user can register for events. */
export async function upsertValorantGameApi(
  accessToken: string,
  gameId: string,
  rankId: string,
  inGameName: string,
): Promise<void> {
  const res = await fetch(`${e2eApiUrl}/users/me/games/${gameId}`, {
    method: "PUT",
    headers: authHeaders(accessToken),
    body: JSON.stringify({
      in_game_name: inGameName,
      current_rank: rankId,
      peak_rank: rankId,
      show_rank: false,
    }),
  });
  await assertOk(res, "upsert user game");
}

export interface EventGroupDetail {
  id: string;
  game_id: string;
  events: { id: string }[];
}

export async function fetchEventGroupApi(accessToken: string, groupId: string): Promise<EventGroupDetail> {
  const res = await fetch(`${e2eApiUrl}/events/${groupId}`, { headers: authHeaders(accessToken) });
  await assertOk(res, "get event group");
  return (await res.json()) as EventGroupDetail;
}

/** Registers the current user for every game in the group (API filler players). */
export async function registerForGroupApi(accessToken: string, group: EventGroupDetail): Promise<void> {
  const res = await fetch(`${e2eApiUrl}/registrations/group/${group.id}/me`, {
    method: "PUT",
    headers: authHeaders(accessToken),
    body: JSON.stringify({
      duo_request: "",
      events: group.events.map((event) => ({
        event_id: event.id,
        can_substitute: false,
        can_lobby_host: false,
      })),
    }),
  });
  await assertOk(res, "register for group");
}

/**
 * Creates a fully onboarded user (profile + Valorant game) without a browser session.
 * Used for lock-in roster fillers; 5v5 needs 10 registrations.
 */
export async function bootstrapPlayerWithValorant(user: TestUser): Promise<string> {
  const { otc } = await issueTestLogin(user);
  const accessToken = await completeAuthApi(otc);
  await saveProfileApi(accessToken, user.globalName ?? user.username);
  const { game, rank } = await fetchValorantCatalog(accessToken);
  await upsertValorantGameApi(accessToken, game.id, rank.id, `${user.username}#E2E`);
  return accessToken;
}

/** Clears new_user (and optional Valorant profile) so a later browser login skips /my_account. */
export async function bootstrapExistingUser(user: TestUser, withValorant: boolean): Promise<void> {
  const { otc } = await issueTestLogin(user);
  const accessToken = await completeAuthApi(otc);
  await saveProfileApi(accessToken, user.globalName ?? user.username);
  if (withValorant) {
    const { game, rank } = await fetchValorantCatalog(accessToken);
    await upsertValorantGameApi(accessToken, game.id, rank.id, `${user.username}#E2E`);
  }
}
