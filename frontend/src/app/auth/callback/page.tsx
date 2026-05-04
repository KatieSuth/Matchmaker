"use client"

// OAuth return path: exchanges the one-time `otc` for tokens, then routes to the app shell.
import { useEffect } from "react";
import { useRouter } from "next/navigation"
import { setAccessToken } from "@/app/_lib/auth";
import { consumePostLoginRedirect } from "@/app/_lib/postLoginRedirect";
import { useAuth } from "@/app/_context/AuthContext";
import { completeAuth } from "@/app/_services/auth";
import { fetchCurrentUser } from "@/app/_services/users";

export default function CallbackPage() {
    const router = useRouter();
    const { setUser, setIsAuthenticated } = useAuth()

    useEffect(() => {
        const params = new URLSearchParams(window.location.search);
        const otc = params.get("otc");
        const isNewUser = params.get("new_user") === "true";

        if (!otc) {
            router.replace("/");
            return;
        }

        const ac = new AbortController();
        const { signal } = ac;

        void (async () => {
            try {
                const response = await completeAuth(otc, signal);
                if (signal.aborted) return;
                if (!response.access_token) {
                    router.replace("/");
                    return;
                }

                const domain = process.env.NEXT_PUBLIC_FRONTEND_DOMAIN;
                const expire = process.env.NEXT_PUBLIC_COOKIE_AUTH_EXPIRE_LIMIT;

                const accessToken = response.access_token;
                setAccessToken(accessToken);
                document.cookie = `auth_session=1; domain=${domain}; path=/; secure; max-age=${expire}`;

                const resolvedUser = await fetchCurrentUser(signal);
                if (signal.aborted) return;
                if (!resolvedUser) {
                    if (isNewUser) {
                        router.replace("/my_account");
                    } else {
                        const next = consumePostLoginRedirect();
                        router.replace(next ?? "/my_events");
                    }
                    return;
                }
                setUser(resolvedUser);
                setIsAuthenticated(true);
                if (isNewUser) {
                    router.replace("/my_account");
                } else {
                    const next = consumePostLoginRedirect();
                    router.replace(next ?? "/my_events");
                }
            } catch {
                if (signal.aborted) return;
                router.replace("/");
            }
        })();

        return () => {
            ac.abort();
        };
    }, [router, setUser, setIsAuthenticated]);

    return (
        <div>Redirecting...</div>
    );
}