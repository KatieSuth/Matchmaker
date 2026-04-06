import axios from "axios";
import { refreshAccessToken, getAccessToken, setAccessToken } from "./auth";

export const api = axios.create({
    baseURL: process.env.NEXT_PUBLIC_API_URL,
    withCredentials: true, // sends the HttpOnly refresh_token cookie automatically
});

// Attach access token to every request
api.interceptors.request.use((config) => {
    const token = getAccessToken();
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});

// On 401, try to refresh then retry the original request
api.interceptors.response.use(
    (response) => response,
    async (error) => {
        const original = error.config;

        if (error.response?.status === 401 && !original._retry) {
            original._retry = true; // prevent infinite retry loop

            try {
                const newToken = await refreshAccessToken();
                original.headers.Authorization = `Bearer ${newToken}`;
                return api(original); // retry original request
            } catch {
                // Refresh failed — clear token and redirect to login
                setAccessToken(null);
                window.location.href = "/";
            }
        }

        return Promise.reject(error);
    }
);

export default api;