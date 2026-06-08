"use client";

// Global session: restores JWT via silent refresh, hydrates /users/me, and exposes logout.
import { createContext, useContext, useEffect, useState } from "react";
import { setAccessToken } from "@/app/_lib/auth";
import { bootstrapSession, clearAuthSessionFlag, setAuthSessionFlag } from "@/app/_lib/session";
import type { User } from "@/app/_types/types";
import { logoutUser } from "@/app/_services/auth";

interface AuthContextType {
    isAuthenticated: boolean;
    isLoading: boolean;
    user: User | null;
    setUser: (user: User) => void;
    setIsAuthenticated: (value: boolean) => void;
    logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
    const [isAuthenticated, setIsAuthenticated] = useState(false);
    const [user, setUser] = useState<User | null>(null);
    
    const [isLoading, setIsLoading] = useState(() => {
        if (typeof window !== "undefined") {
            return window.location.pathname !== "/auth/callback";
        }
        return true;
    })

    // On mount, try a silent refresh to restore session after page reload
    // The refresh_token HttpOnly cookie will be sent automatically if it exists
    useEffect(() => {
        if (window.location.pathname === "/auth/callback") {
            return;
        }

        const ac = new AbortController();
        const { signal } = ac;

        void (async () => {
            const resolvedUser = await bootstrapSession();
            if (signal.aborted) return;
            if (resolvedUser) {
                setIsAuthenticated(true);
                setUser(resolvedUser);
                setAuthSessionFlag();
            } else {
                setIsAuthenticated(false);
                setUser(null);
                setAccessToken(null);
                clearAuthSessionFlag();
            }
            if (!signal.aborted) {
                setIsLoading(false);
            }
        })();

        return () => {
            ac.abort();
        };
    }, []);

    const logout = async () => {
        await logoutUser(); // clears the HttpOnly cookie server-side
        setAccessToken(null);
        setIsAuthenticated(false);
    };

    return (
        <AuthContext.Provider value={{ isAuthenticated, setIsAuthenticated, isLoading, user, setUser, logout }}>
            {children}
        </AuthContext.Provider>
    );
}

export function useAuth() {
    const ctx = useContext(AuthContext);
    if (!ctx) throw new Error("useAuth must be used within AuthProvider");
    return ctx;
}