/** Client-side lobby join code / invite URL helpers. Backend remains source of truth on write. */

const PLAIN_CODE_MAX = 64;
const LINK_PATH_MAX = 512;
const PLAIN_CODE_RE = /^[A-Za-z0-9-]+$/;
// Single path segment + query only (e.g. /LOL?… or /VAL?…). Extra "/" after the
// leading slash is rejected for now; nested paths are a future consideration.
const LINK_PATH_RE = /^\/[A-Za-z0-9?&\-._=%]*$/;

/** True when join_code is a stored invite path (leading /), not a plain lobby code. */
export function isLobbyJoinLinkPath(joinCode: string | null | undefined): boolean {
  return typeof joinCode === "string" && joinCode.startsWith("/");
}

/** Parses join_link_base as an https origin with no path (aside from optional /). */
function parseHttpsBase(joinLinkBase: string): URL | null {
  try {
    const parsed = new URL(joinLinkBase.trim());
    if (parsed.protocol !== "https:") return null;
    if (parsed.username || parsed.password) return null;
    const path = parsed.pathname;
    if (path && path !== "/") return null;
    return parsed;
  } catch {
    return null;
  }
}

/** Charset/length checks mirroring backend validateStoredPath before building an href. */
function isSafeStoredPath(path: string): boolean {
  if (path.length > LINK_PATH_MAX) return false;
  if (!path.startsWith("/") || path.startsWith("//")) return false;
  if (path.includes(":") || path.includes("@") || path.includes("\\")) return false;
  if (/\s/.test(path)) return false;
  return LINK_PATH_RE.test(path);
}

/** Rebuild a display/copy value. Returns a full https URL only when safe; otherwise the raw code or null. */
export function buildLobbyJoinDisplayValue(
  joinCode: string | null | undefined,
  joinLinkBase: string | null | undefined,
): { kind: "link" | "code"; value: string } | null {
  if (!joinCode) return null;
  if (isLobbyJoinLinkPath(joinCode) && joinLinkBase) {
    if (!isSafeStoredPath(joinCode)) return null;
    const base = parseHttpsBase(joinLinkBase);
    if (!base) return null;
    try {
      const rebuilt = new URL(joinLinkBase.replace(/\/$/, "") + joinCode);
      if (rebuilt.protocol !== "https:") return null;
      if (rebuilt.hostname.toLowerCase() !== base.hostname.toLowerCase()) return null;
      if (rebuilt.port !== base.port) return null;
      if (rebuilt.username || rebuilt.password) return null;
      return { kind: "link", value: rebuilt.toString() };
    } catch {
      return null;
    }
  }
  return { kind: "code", value: joinCode };
}

/**
 * Detects URL-shaped input including schemeless hosts (e.g. gg.badactor.net/x)
 * so those values are validated as links rather than plain codes.
 */
function isLinkShaped(s: string): boolean {
  const lower = s.toLowerCase();
  if (lower.startsWith("http://") || lower.startsWith("https://")) return true;
  let cut = -1;
  const slash = s.indexOf("/");
  const qmark = s.indexOf("?");
  if (slash >= 0 && qmark >= 0) cut = Math.min(slash, qmark);
  else if (slash >= 0) cut = slash;
  else if (qmark >= 0) cut = qmark;
  if (cut <= 0) return false;
  const host = s.slice(0, cut);
  return host.includes(".") && !/[\s]/.test(host);
}

/** UX-only validation mirroring backend normalize rules. */
export function validateLobbyJoinInput(
  raw: string,
  joinLinkBase: string | null | undefined,
): string | null {
  const trimmed = raw.trim();
  if (trimmed === "") return null;

  if (isLinkShaped(trimmed)) {
    if (!joinLinkBase?.trim()) {
      return "Links are not supported for this game; enter a lobby code instead";
    }
    const base = parseHttpsBase(joinLinkBase);
    if (!base) return "Game join link base is misconfigured";

    let toParse = trimmed;
    const lower = trimmed.toLowerCase();
    if (!lower.startsWith("http://") && !lower.startsWith("https://")) {
      toParse = "https://" + trimmed;
    }
    let parsed: URL;
    try {
      parsed = new URL(toParse);
    } catch {
      return "Invalid lobby join link";
    }
    if (parsed.protocol !== "https:") return "Lobby join links must use https";
    if (parsed.username || parsed.password) return "Invalid lobby join link";
    if (parsed.hostname.toLowerCase() !== base.hostname.toLowerCase()) {
      return "Lobby join link must use the official game join host";
    }
    if (parsed.port !== base.port) {
      return "Lobby join link must use the official game join host";
    }
    const path = parsed.pathname || "";
    if (!path || path === "/") {
      if (!parsed.search) return "Lobby join link is missing a path";
    }
    const stored = path + (parsed.search || "");
    if (!isSafeStoredPath(stored.startsWith("/") ? stored : `/${stored}`)) {
      return "Invalid lobby join link path";
    }
    return null;
  }

  if (trimmed.startsWith("/") || /[?/]/.test(trimmed)) {
    if (!joinLinkBase?.trim()) {
      return "Links are not supported for this game; enter a lobby code instead";
    }
    const path = trimmed.startsWith("/") ? trimmed : `/${trimmed}`;
    if (!isSafeStoredPath(path)) return "Invalid lobby join link path";
    return null;
  }

  if (trimmed.length > PLAIN_CODE_MAX) return "Lobby code is too long";
  if (!PLAIN_CODE_RE.test(trimmed)) {
    return "Lobby code must contain only letters, digits, and hyphens";
  }
  return null;
}
