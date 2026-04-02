"use client";

import { createContext, useContext, useEffect, useState } from "react";
import { setAccessToken, refreshAccessToken } from "@/app/_lib/auth";
import api from "@/app/_lib/axios"

interface AuthContextType {
    isAuthenticated: boolean;
    isLoading: boolean;
    logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
    const [isAuthenticated, setIsAuthenticated] = useState(false);
    const [isLoading, setIsLoading] = useState(true);

    // On mount, try a silent refresh to restore session after page reload
    // The refresh_token HttpOnly cookie will be sent automatically if it exists
    useEffect(() => {
        refreshAccessToken()
            .then(() => setIsAuthenticated(true))
            .catch(() => setIsAuthenticated(false))
            .finally(() => setIsLoading(false));
    }, []);

    const logout = async () => {
        await api.post("/auth/logout"); // clears the HttpOnly cookie server-side
        setAccessToken(null);
        setIsAuthenticated(false);
    };

    return (
        <AuthContext.Provider value={{ isAuthenticated, isLoading, logout }}>
            {children}
        </AuthContext.Provider>
    );
}

export function useAuth() {
    const ctx = useContext(AuthContext);
    if (!ctx) throw new Error("useAuth must be used within AuthProvider");
    return ctx;
}