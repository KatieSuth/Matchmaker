// Tests for the route-guard proxy: origin-verify enforcement, auth-gated redirects, and the
// logged-in "/" -> /my_events (or safe ?next=) redirect. Uses real `NextRequest` instances rather
// than mocking `next/server`, since NextRequest/NextResponse work fine outside the full Next.js
// runtime for this kind of pure request-in/response-out unit.
import { NextRequest } from "next/server";
import { afterEach, describe, expect, it, vi } from "vitest";
import { proxy } from "./proxy";

function makeRequest(url: string, options?: { authSession?: boolean; originVerifyHeader?: string }): NextRequest {
  const headers = new Headers();
  if (options?.originVerifyHeader !== undefined) {
    headers.set("x-origin-verify", options.originVerifyHeader);
  }
  const request = new NextRequest(new URL(url, "https://app.example.com"), { headers });
  if (options?.authSession) {
    request.cookies.set("auth_session", "1");
  }
  return request;
}

describe("proxy — origin verification", () => {
  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it("returns 403 when ORIGIN_VERIFY_SECRET is set and the header is missing", () => {
    vi.stubEnv("ORIGIN_VERIFY_SECRET", "shh");
    const response = proxy(makeRequest("/my_events", { authSession: true }));
    expect(response.status).toBe(403);
  });

  it("returns 403 when the header doesn't match the secret", () => {
    vi.stubEnv("ORIGIN_VERIFY_SECRET", "shh");
    const response = proxy(
      makeRequest("/my_events", { authSession: true, originVerifyHeader: "wrong" }),
    );
    expect(response.status).toBe(403);
  });

  it("passes through when the header matches the secret", () => {
    vi.stubEnv("ORIGIN_VERIFY_SECRET", "shh");
    const response = proxy(
      makeRequest("/my_events", { authSession: true, originVerifyHeader: "shh" }),
    );
    expect(response.status).not.toBe(403);
  });

  it("skips origin verification for /health regardless of the header", () => {
    vi.stubEnv("ORIGIN_VERIFY_SECRET", "shh");
    const response = proxy(makeRequest("/health"));
    expect(response.status).not.toBe(403);
  });

  it("doesn't check origin verification when no secret is configured", () => {
    vi.stubEnv("ORIGIN_VERIFY_SECRET", "");
    const response = proxy(makeRequest("/my_events"));
    expect(response.status).not.toBe(403);
  });
});

describe("proxy — auth gating", () => {
  it("redirects an unauthenticated request on a protected path to '/' with ?next=", () => {
    const response = proxy(makeRequest("/event/abc123?tab=teams"));
    expect(response.status).toBe(307);
    const location = new URL(response.headers.get("location")!);
    expect(location.pathname).toBe("/");
    expect(location.searchParams.get("next")).toBe("/event/abc123?tab=teams");
  });

  it("does not redirect an unauthenticated request to '/' itself", () => {
    const response = proxy(makeRequest("/"));
    expect(response.status).not.toBe(307);
  });

  it("does not redirect an unauthenticated request to /auth/callback (mid-OAuth)", () => {
    const response = proxy(makeRequest("/auth/callback"));
    expect(response.status).not.toBe(307);
  });

  it("does not redirect an unauthenticated request to /about (public page)", () => {
    const response = proxy(makeRequest("/about"));
    expect(response.status).not.toBe(307);
  });

  it("does not redirect an unauthenticated request to /health", () => {
    const response = proxy(makeRequest("/health"));
    expect(response.status).not.toBe(307);
  });

  it("passes through an authenticated request to a protected path", () => {
    const response = proxy(makeRequest("/event/abc123", { authSession: true }));
    expect(response.status).not.toBe(307);
  });
});

describe("proxy — logged-in landing redirect", () => {
  it("redirects an authenticated request on '/' to /my_events by default", () => {
    const response = proxy(makeRequest("/", { authSession: true }));
    expect(response.status).toBe(307);
    const location = new URL(response.headers.get("location")!);
    expect(location.pathname).toBe("/my_events");
  });

  it("redirects to a safe ?next= path when present", () => {
    const response = proxy(makeRequest("/?next=%2Fevent%2Fabc123", { authSession: true }));
    const location = new URL(response.headers.get("location")!);
    expect(location.pathname).toBe("/event/abc123");
  });

  it("falls back to /my_events when ?next= is an unsafe path", () => {
    const response = proxy(makeRequest("/?next=%2F%2Fevil.com", { authSession: true }));
    const location = new URL(response.headers.get("location")!);
    expect(location.pathname).toBe("/my_events");
  });
});
