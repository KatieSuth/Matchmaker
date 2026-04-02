import axios from "axios";

// Stored in module scope — survives re-renders, lost on page refresh
let accessToken: string | null = null;

export function getAccessToken(): string | null {
    return accessToken;
}

export function setAccessToken(token: string | null): void {
    accessToken = token;
}

export async function refreshAccessToken(): Promise<string> {
// withCredentials sends the HttpOnly cookie automatically
const response = await axios.post(
    `${process.env.NEXT_PUBLIC_API_URL}/auth/refresh`,
    {},
    { withCredentials: true }
);

const newToken = response.data.access_token;
    setAccessToken(newToken);
    return newToken;
}