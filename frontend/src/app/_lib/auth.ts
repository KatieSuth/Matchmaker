import axios from "axios";

// Stored in module scope — survives re-renders, lost on page refresh
let accessToken: string | null = null;

// Ensures only one in-flight /auth/refresh at a time. Parallel 401s from multiple
// API calls must not each trigger a separate refresh while the server rotates
// refresh tokens, or the "losing" refresh can fail and the app may bounce
// the user to "/" even though a valid session exists.
let refreshInFlight: Promise<string> | null = null;

export function getAccessToken(): string | null {
    return accessToken;
}

export function setAccessToken(token: string | null): void {
    accessToken = token;
}

export async function refreshAccessToken(_signal?: AbortSignal): Promise<string> {
    if (refreshInFlight) {
        return refreshInFlight;
    }

    refreshInFlight = (async () => {
        // withCredentials sends the HttpOnly refresh_token cookie automatically.
        // Do not pass AbortSignal here: AuthProvider's Strict Mode effect cleanup aborts
        // the shared controller while this POST is in flight. That cancels the singleton
        // refresh used by every 401 retry — the server may still rotate and revoke the old
        // token, then the next refresh sees "token not in DB" / 401.
        const response = await axios.post(
            `${process.env.NEXT_PUBLIC_API_URL}/auth/refresh`,
            {},
            { withCredentials: true }
        );

        const newToken = response.data.access_token as string;
        setAccessToken(newToken);
        return newToken;
    })().finally(() => {
        refreshInFlight = null;
    });

    return refreshInFlight;
}