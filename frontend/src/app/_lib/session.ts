import { refreshAccessToken } from "@/app/_lib/auth";
import { fetchCurrentUser } from "@/app/_services/users";
import type { User } from "@/app/_types/types";

let sessionBootstrapInFlight: Promise<User | null> | null = null;

/** Clears the client-visible auth_session flag for both domain-scoped and host-only cookies. */
export function clearAuthSessionFlag() {
  const domain = process.env.NEXT_PUBLIC_FRONTEND_DOMAIN;
  if (!domain) return;
  document.cookie = `auth_session=; Max-Age=0; domain=${domain}; path=/; secure`;
  document.cookie = "auth_session=; Max-Age=0; path=/; secure";
}

/** Sets the client-visible auth_session flag used by Next.js route guards after a successful login. */
export function setAuthSessionFlag() {
  const domain = process.env.NEXT_PUBLIC_FRONTEND_DOMAIN;
  const expire = process.env.NEXT_PUBLIC_COOKIE_AUTH_EXPIRE_LIMIT;
  if (!domain || !expire) return;
  document.cookie = `auth_session=1; domain=${domain}; path=/; secure; max-age=${expire}`;
}

/**
 * Restores the session after reload: silent refresh, then /users/me.
 * Dedupes concurrent calls (e.g. React Strict Mode double-mount).
 */
export async function bootstrapSession(): Promise<User | null> {
  if (sessionBootstrapInFlight) {
    return sessionBootstrapInFlight;
  }

  sessionBootstrapInFlight = (async () => {
    try {
      await refreshAccessToken();
      return await fetchCurrentUser();
    } catch {
      return null;
    } finally {
      sessionBootstrapInFlight = null;
    }
  })();

  return sessionBootstrapInFlight;
}
