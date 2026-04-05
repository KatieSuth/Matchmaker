"use client"

import { useEffect } from "react";
import { useRouter } from "next/navigation"
import { setAccessToken } from "@/app/_lib/auth";
import { api } from "@/app/_lib/api";

interface CompleteAuthResponse {
    access_token: string;
}

export default function CallbackPage() {
    const router = useRouter();

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
            router.replace(isNewUser ? "/my_account" : "/events");
        })
    }, []);

    return (
        <div>Redirecting...</div>
    );
}