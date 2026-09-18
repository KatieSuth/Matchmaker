// Tests for client-side lobby join code/link validation and display-value rebuilding. These
// mirror backend validation rules (see module comment) and are security-relevant: they gate what
// gets rendered as a clickable link, so malformed/malicious input must be rejected, not just
// "handled gracefully".
import { describe, expect, it } from "vitest";
import { buildLobbyJoinDisplayValue, isLobbyJoinLinkPath, validateLobbyJoinInput } from "./lobbyJoin";

describe("isLobbyJoinLinkPath", () => {
  it("returns true for a stored path (leading slash)", () => {
    expect(isLobbyJoinLinkPath("/LOL?code=abc")).toBe(true);
  });

  it("returns false for a plain lobby code", () => {
    expect(isLobbyJoinLinkPath("ABC-123")).toBe(false);
  });

  it("returns false for null/undefined", () => {
    expect(isLobbyJoinLinkPath(null)).toBe(false);
    expect(isLobbyJoinLinkPath(undefined)).toBe(false);
  });
});

describe("buildLobbyJoinDisplayValue", () => {
  const base = "https://play.example.com";

  it("returns null when there is no join code", () => {
    expect(buildLobbyJoinDisplayValue(null, base)).toBeNull();
    expect(buildLobbyJoinDisplayValue("", base)).toBeNull();
  });

  it("returns a plain code value for a non-path code", () => {
    expect(buildLobbyJoinDisplayValue("ABC-123", base)).toEqual({ kind: "code", value: "ABC-123" });
  });

  it("rebuilds a full https link for a stored path with a valid base", () => {
    expect(buildLobbyJoinDisplayValue("/LOL?code=abc", base)).toEqual({
      kind: "link",
      value: "https://play.example.com/LOL?code=abc",
    });
  });

  it("handles a trailing slash on the base without double-slashing", () => {
    expect(buildLobbyJoinDisplayValue("/LOL", "https://play.example.com/")).toEqual({
      kind: "link",
      value: "https://play.example.com/LOL",
    });
  });

  it("falls back to displaying the raw stored path as a code when there is no join link base to resolve it against", () => {
    expect(buildLobbyJoinDisplayValue("/LOL", null)).toEqual({ kind: "code", value: "/LOL" });
    expect(buildLobbyJoinDisplayValue("/LOL", undefined)).toEqual({ kind: "code", value: "/LOL" });
  });

  it("returns null when the join link base isn't https", () => {
    expect(buildLobbyJoinDisplayValue("/LOL", "http://play.example.com")).toBeNull();
  });

  it("returns null when the join link base has a non-root path", () => {
    expect(buildLobbyJoinDisplayValue("/LOL", "https://play.example.com/base")).toBeNull();
  });

  it("returns null when the join link base has embedded credentials", () => {
    expect(buildLobbyJoinDisplayValue("/LOL", "https://user:pass@play.example.com")).toBeNull();
  });

  it("returns null for a stored path containing '//' (protocol-relative escape attempt)", () => {
    expect(buildLobbyJoinDisplayValue("//evil.com", base)).toBeNull();
  });

  it("returns null for a stored path containing a backslash", () => {
    expect(buildLobbyJoinDisplayValue("/LOL\\evil", base)).toBeNull();
  });

  it("returns null for a stored path containing whitespace", () => {
    expect(buildLobbyJoinDisplayValue("/LOL abc", base)).toBeNull();
  });

  it("returns null for a stored path over the length limit", () => {
    expect(buildLobbyJoinDisplayValue("/" + "a".repeat(600), base)).toBeNull();
  });

  it("preserves a non-default port from the join link base in the rebuilt URL", () => {
    expect(buildLobbyJoinDisplayValue("/LOL", "https://play.example.com:8443")).toEqual({
      kind: "link",
      value: "https://play.example.com:8443/LOL",
    });
  });
});

describe("validateLobbyJoinInput", () => {
  const base = "https://play.example.com";

  it("allows an empty value (clearing the join code)", () => {
    expect(validateLobbyJoinInput("", base)).toBeNull();
    expect(validateLobbyJoinInput("   ", base)).toBeNull();
  });

  it("accepts a plain alphanumeric-and-hyphen code", () => {
    expect(validateLobbyJoinInput("ABC-123", base)).toBeNull();
  });

  it("rejects a plain code that's too long", () => {
    expect(validateLobbyJoinInput("a".repeat(65), base)).toBe("Lobby code is too long");
  });

  it("rejects a plain code with disallowed characters", () => {
    expect(validateLobbyJoinInput("abc 123", base)).toBe("Lobby code must contain only letters, digits, and hyphens");
  });

  it("rejects a full URL when the game has no join link base", () => {
    expect(validateLobbyJoinInput("https://play.example.com/LOL", null)).toBe(
      "Links are not supported for this game; enter a lobby code instead",
    );
  });

  it("accepts a full https URL matching the configured join link base", () => {
    expect(validateLobbyJoinInput("https://play.example.com/LOL?code=abc", base)).toBeNull();
  });

  it("accepts a schemeless host that looks like a link and matches the base", () => {
    expect(validateLobbyJoinInput("play.example.com/LOL", base)).toBeNull();
  });

  it("rejects a URL using http instead of https", () => {
    expect(validateLobbyJoinInput("http://play.example.com/LOL", base)).toBe("Lobby join links must use https");
  });

  it("rejects a URL pointing at a different host (phishing/redirect vector)", () => {
    expect(validateLobbyJoinInput("https://evil.com/LOL", base)).toBe(
      "Lobby join link must use the official game join host",
    );
  });

  it("rejects a URL with embedded credentials", () => {
    expect(validateLobbyJoinInput("https://user:pass@play.example.com/LOL", base)).toBe("Invalid lobby join link");
  });

  it("rejects a URL with no path and no query", () => {
    expect(validateLobbyJoinInput("https://play.example.com", base)).toBe("Lobby join link is missing a path");
  });

  it("accepts a URL with no path but a query string", () => {
    expect(validateLobbyJoinInput("https://play.example.com?code=abc", base)).toBeNull();
  });

  it("rejects an unparsable URL", () => {
    expect(validateLobbyJoinInput("https://", base)).toBe("Invalid lobby join link");
  });

  it("reports a misconfigured join link base separately from a bad user URL", () => {
    expect(validateLobbyJoinInput("https://play.example.com/LOL", "not-a-url")).toBe(
      "Game join link base is misconfigured",
    );
  });

  it("accepts a bare path/query as a link when a join link base is configured", () => {
    expect(validateLobbyJoinInput("/LOL?code=abc", base)).toBeNull();
  });

  it("rejects a bare path when the game has no join link base", () => {
    expect(validateLobbyJoinInput("/LOL", null)).toBe(
      "Links are not supported for this game; enter a lobby code instead",
    );
  });

  it("rejects a bare path containing unsafe characters", () => {
    expect(validateLobbyJoinInput("/LOL\\evil", base)).toBe("Invalid lobby join link path");
  });
});
