// Tests for the auth_session cookie helpers and the dedupe-on-concurrent-calls session bootstrap.
import { beforeEach, describe, expect, it, vi } from "vitest";
import { refreshAccessToken } from "@/app/_lib/auth";
import { fetchCurrentUser } from "@/app/_services/users";
import { buildUser } from "@/test/fixtures";
import { bootstrapSession, clearAuthSessionFlag, setAuthSessionFlag } from "./session";

vi.mock("@/app/_lib/auth", () => ({ refreshAccessToken: vi.fn() }));
vi.mock("@/app/_services/users", () => ({ fetchCurrentUser: vi.fn() }));

describe("setAuthSessionFlag / clearAuthSessionFlag", () => {
  it("sets a domain-scoped, max-age cookie when both env vars are configured", () => {
    vi.stubEnv("NEXT_PUBLIC_FRONTEND_DOMAIN", "example.com");
    vi.stubEnv("NEXT_PUBLIC_COOKIE_AUTH_EXPIRE_LIMIT", "3600");
    const setCookie = vi.spyOn(document, "cookie", "set");

    setAuthSessionFlag();

    expect(setCookie).toHaveBeenCalledWith("auth_session=1; domain=example.com; path=/; secure; max-age=3600");
  });

  it("is a no-op when the domain env var is missing", () => {
    vi.stubEnv("NEXT_PUBLIC_FRONTEND_DOMAIN", "");
    vi.stubEnv("NEXT_PUBLIC_COOKIE_AUTH_EXPIRE_LIMIT", "3600");
    const setCookie = vi.spyOn(document, "cookie", "set");

    setAuthSessionFlag();

    expect(setCookie).not.toHaveBeenCalled();
  });

  it("clears both the domain-scoped and host-only cookies when domain is configured", () => {
    vi.stubEnv("NEXT_PUBLIC_FRONTEND_DOMAIN", "example.com");
    const setCookie = vi.spyOn(document, "cookie", "set");

    clearAuthSessionFlag();

    expect(setCookie).toHaveBeenCalledWith("auth_session=; Max-Age=0; domain=example.com; path=/; secure");
    expect(setCookie).toHaveBeenCalledWith("auth_session=; Max-Age=0; path=/; secure");
  });

  it("clearAuthSessionFlag is a no-op when the domain env var is missing", () => {
    vi.stubEnv("NEXT_PUBLIC_FRONTEND_DOMAIN", "");
    const setCookie = vi.spyOn(document, "cookie", "set");

    clearAuthSessionFlag();

    expect(setCookie).not.toHaveBeenCalled();
  });
});

describe("bootstrapSession", () => {
  beforeEach(() => {
    vi.mocked(refreshAccessToken).mockReset();
    vi.mocked(fetchCurrentUser).mockReset();
  });

  it("returns the resolved user after a successful refresh + /users/me", async () => {
    const user = buildUser();
    vi.mocked(refreshAccessToken).mockResolvedValue("token");
    vi.mocked(fetchCurrentUser).mockResolvedValue(user);

    await expect(bootstrapSession()).resolves.toEqual(user);
  });

  it("returns null (not a rejection) when refresh fails", async () => {
    vi.mocked(refreshAccessToken).mockRejectedValue(new Error("no refresh cookie"));

    await expect(bootstrapSession()).resolves.toBeNull();
    expect(fetchCurrentUser).not.toHaveBeenCalled();
  });

  it("returns null when /users/me fails after a successful refresh", async () => {
    vi.mocked(refreshAccessToken).mockResolvedValue("token");
    vi.mocked(fetchCurrentUser).mockRejectedValue(new Error("network error"));

    await expect(bootstrapSession()).resolves.toBeNull();
  });

  it("dedupes concurrent calls into a single refresh + /users/me round trip", async () => {
    const user = buildUser();
    vi.mocked(refreshAccessToken).mockResolvedValue("token");
    vi.mocked(fetchCurrentUser).mockResolvedValue(user);

    const [first, second] = await Promise.all([bootstrapSession(), bootstrapSession()]);

    expect(first).toEqual(user);
    expect(second).toEqual(user);
    expect(refreshAccessToken).toHaveBeenCalledTimes(1);
    expect(fetchCurrentUser).toHaveBeenCalledTimes(1);
  });

  it("allows a fresh bootstrap after a prior call has settled", async () => {
    vi.mocked(refreshAccessToken).mockResolvedValue("token");
    vi.mocked(fetchCurrentUser).mockResolvedValue(buildUser());

    await bootstrapSession();
    await bootstrapSession();

    expect(refreshAccessToken).toHaveBeenCalledTimes(2);
  });
});
