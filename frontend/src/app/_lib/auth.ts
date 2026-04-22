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

export async function refreshAccessToken(): Promise<string> {
    if (refreshInFlight) {
        return refreshInFlight;
    }

    refreshInFlight = (async () => {
        // withCredentials sends the HttpOnly refresh_token cookie automatically
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