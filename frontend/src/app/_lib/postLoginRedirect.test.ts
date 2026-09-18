// Tests for the Discord OAuth deep-link redirect helpers (sessionStorage-backed, client-only).
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { POST_LOGIN_REDIRECT_STORAGE_KEY } from "@/app/_lib/constants";
import {
  capturePostLoginRedirectFromWindowSearch,
  consumePostLoginRedirect,
  isAllowedPostLoginPath,
  peekPostLoginRedirect,
  persistPostLoginRedirect,
} from "./postLoginRedirect";

describe("isAllowedPostLoginPath", () => {
  it("allows a plain event group path", () => {
    expect(isAllowedPostLoginPath("/event/abc123")).toBe(true);
  });

  it("allows an event group path with a query string", () => {
    expect(isAllowedPostLoginPath("/event/abc123?tab=teams")).toBe(true);
  });

  it("rejects a protocol-relative path (open-redirect vector)", () => {
    expect(isAllowedPostLoginPath("//evil.com/event/abc")).toBe(false);
  });

  it("rejects a path that doesn't start with /event/", () => {
    expect(isAllowedPostLoginPath("/my_events")).toBe(false);
    expect(isAllowedPostLoginPath("/")).toBe(false);
  });

  it("rejects an id containing a nested path segment", () => {
    expect(isAllowedPostLoginPath("/event/abc/def")).toBe(false);
  });

  it("rejects an id containing '..'", () => {
    expect(isAllowedPostLoginPath("/event/..")).toBe(false);
  });

  it("rejects an empty id", () => {
    expect(isAllowedPostLoginPath("/event/")).toBe(false);
  });

  it("rejects an id with disallowed characters", () => {
    expect(isAllowedPostLoginPath("/event/abc def")).toBe(false);
    expect(isAllowedPostLoginPath("/event/abc<script>")).toBe(false);
  });

  it("rejects non-string input", () => {
    expect(isAllowedPostLoginPath(null as unknown as string)).toBe(false);
    expect(isAllowedPostLoginPath(undefined as unknown as string)).toBe(false);
  });
});

describe("sessionStorage-backed redirect persistence", () => {
  beforeEach(() => {
    sessionStorage.clear();
  });

  afterEach(() => {
    sessionStorage.clear();
  });

  it("persistPostLoginRedirect stores an allowed path", () => {
    persistPostLoginRedirect("/event/abc123");
    expect(sessionStorage.getItem(POST_LOGIN_REDIRECT_STORAGE_KEY)).toBe("/event/abc123");
  });

  it("persistPostLoginRedirect ignores a disallowed path", () => {
    persistPostLoginRedirect("/my_events");
    expect(sessionStorage.getItem(POST_LOGIN_REDIRECT_STORAGE_KEY)).toBeNull();
  });

  it("peekPostLoginRedirect returns the stored path without clearing it", () => {
    sessionStorage.setItem(POST_LOGIN_REDIRECT_STORAGE_KEY, "/event/abc123");
    expect(peekPostLoginRedirect()).toBe("/event/abc123");
    expect(sessionStorage.getItem(POST_LOGIN_REDIRECT_STORAGE_KEY)).toBe("/event/abc123");
  });

  it("peekPostLoginRedirect returns null for a stored-but-now-invalid path", () => {
    sessionStorage.setItem(POST_LOGIN_REDIRECT_STORAGE_KEY, "//evil.com");
    expect(peekPostLoginRedirect()).toBeNull();
  });

  it("peekPostLoginRedirect returns null when nothing is stored", () => {
    expect(peekPostLoginRedirect()).toBeNull();
  });

  it("consumePostLoginRedirect returns and clears the stored path", () => {
    sessionStorage.setItem(POST_LOGIN_REDIRECT_STORAGE_KEY, "/event/abc123");
    expect(consumePostLoginRedirect()).toBe("/event/abc123");
    expect(sessionStorage.getItem(POST_LOGIN_REDIRECT_STORAGE_KEY)).toBeNull();
  });

  it("consumePostLoginRedirect returns null and doesn't throw when nothing is stored", () => {
    expect(consumePostLoginRedirect()).toBeNull();
  });

  it("capturePostLoginRedirectFromWindowSearch stores ?next= from the current URL when allowed", () => {
    window.history.pushState({}, "", "/?next=%2Fevent%2Fabc123");
    capturePostLoginRedirectFromWindowSearch();
    expect(sessionStorage.getItem(POST_LOGIN_REDIRECT_STORAGE_KEY)).toBe("/event/abc123");
  });

  it("capturePostLoginRedirectFromWindowSearch ignores a disallowed ?next=", () => {
    window.history.pushState({}, "", "/?next=%2F%2Fevil.com");
    capturePostLoginRedirectFromWindowSearch();
    expect(sessionStorage.getItem(POST_LOGIN_REDIRECT_STORAGE_KEY)).toBeNull();
  });
});
