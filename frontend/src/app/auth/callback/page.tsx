"use client"

import { useEffect } from "react";
import { useRouter } from "next/navigation"
import { setAccessToken } from "@/app/_lib/auth";
import { api } from "@/app/_lib/axios";
import { useAuth } from "@/app/_context/AuthContext";
import type { User } from "@/app/_types/user";

interface CompleteAuthResponse {
    access_token: string;
}

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
        
        api.post<CompleteAuthResponse>(`/auth/complete`, { otc })
        .then((response) => {
            if (!response.data?.access_token) {
                router.replace("/");
                return;
            }

            const accessToken = response.data.access_token;
            setAccessToken(accessToken);
            document.cookie = `auth_session=1; domain=matchmaker.localhost; path=/; secure; max-age${7 * 24 * 60 * 60}`;
            
            return api.get<User>("/users/me").then((res) => {
                if (!res.data) {
                    router.replace(isNewUser ? "/my_account" : "/events");
                    return;
                }
                setUser(res.data);
                setIsAuthenticated(true);
                router.replace(isNewUser ? "/my_account" : "/events");
            })
        })
    }, []);

    return (
        <div>Redirecting...</div>
    );
}