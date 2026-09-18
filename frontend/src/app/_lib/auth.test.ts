// Tests for the module-singleton access token store and the dedupe-in-flight token refresh.
// `refreshAccessToken` posts via a bare `axios` call (not the shared `api` instance — see the
// comment in auth.ts on why it can't accept an AbortSignal), so it's exercised here against the
// real MSW-mocked `/auth/refresh` endpoint rather than a mocked axios module.
import { http, HttpResponse } from "msw";
import { describe, expect, it } from "vitest";
import { server } from "@/test/msw/server";
import { TEST_API_URL } from "@/test/constants";
import { getAccessToken, refreshAccessToken, setAccessToken } from "./auth";

describe("getAccessToken / setAccessToken", () => {
  it("starts null (reset by the shared test setup's afterEach)", () => {
    expect(getAccessToken()).toBeNull();
  });

  it("returns whatever was last set", () => {
    setAccessToken("abc123");
    expect(getAccessToken()).toBe("abc123");
    setAccessToken(null);
    expect(getAccessToken()).toBeNull();
  });
});

describe("refreshAccessToken", () => {
  it("stores and returns the new access token on success", async () => {
    server.use(
      http.post(`${TEST_API_URL}/auth/refresh`, () => HttpResponse.json({ access_token: "fresh-token" })),
    );

    const token = await refreshAccessToken();

    expect(token).toBe("fresh-token");
    expect(getAccessToken()).toBe("fresh-token");
  });

  it("propagates a rejection when the refresh request fails", async () => {
    server.use(http.post(`${TEST_API_URL}/auth/refresh`, () => new HttpResponse(null, { status: 401 })));

    await expect(refreshAccessToken()).rejects.toBeTruthy();
  });

  it("dedupes concurrent calls into a single request", async () => {
    let requestCount = 0;
    server.use(
      http.post(`${TEST_API_URL}/auth/refresh`, () => {
        requestCount += 1;
        return HttpResponse.json({ access_token: `token-${requestCount}` });
      }),
    );

    const [first, second] = await Promise.all([refreshAccessToken(), refreshAccessToken()]);

    expect(requestCount).toBe(1);
    expect(first).toBe(second);
  });

  it("allows a fresh request after a prior refresh has settled", async () => {
    let requestCount = 0;
    server.use(
      http.post(`${TEST_API_URL}/auth/refresh`, () => {
        requestCount += 1;
        return HttpResponse.json({ access_token: `token-${requestCount}` });
      }),
    );

    await refreshAccessToken();
    await refreshAccessToken();

    expect(requestCount).toBe(2);
  });
});
