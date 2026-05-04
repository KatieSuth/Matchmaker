// Preserve deep-link intent across Discord OAuth: middleware adds ?next=, sessionStorage survives the round-trip.

export const POST_LOGIN_REDIRECT_STORAGE_KEY = "postLoginRedirect";

const STORAGE_KEY = POST_LOGIN_REDIRECT_STORAGE_KEY;

/** Shared validation for middleware + client (event group pages only). */
export function isAllowedPostLoginPath(path: string): boolean {
  if (typeof path !== "string") return false;
  if (!path.startsWith("/") || path.startsWith("//")) return false;
  const pathname = path.split("?")[0];
  if (!pathname.startsWith("/event/")) return false;
  const id = pathname.slice("/event/".length);
  if (!id || id.includes("/") || id.includes("..")) return false;
  return /^[a-zA-Z0-9_-]+$/.test(id);
}

/** Landing page: persist ?next= before user clicks Discord login. */
export function capturePostLoginRedirectFromWindowSearch(): void {
  if (typeof window === "undefined") return;
  const next = new URLSearchParams(window.location.search).get("next");
  if (next && isAllowedPostLoginPath(next)) {
    sessionStorage.setItem(STORAGE_KEY, next);
  }
}

/** Event page when logged out: remember this URL for after login. */
export function persistPostLoginRedirect(path: string): void {
  if (typeof window === "undefined") return;
  if (isAllowedPostLoginPath(path)) {
    sessionStorage.setItem(STORAGE_KEY, path);
  }
}

export function peekPostLoginRedirect(): string | null {
  if (typeof window === "undefined") return null;
  const raw = sessionStorage.getItem(STORAGE_KEY);
  return raw && isAllowedPostLoginPath(raw) ? raw : null;
}

/** Read validated path and remove from storage (call after successful login routing). */
export function consumePostLoginRedirect(): string | null {
  const path = peekPostLoginRedirect();
  if (path) sessionStorage.removeItem(STORAGE_KEY);
  return path;
}
