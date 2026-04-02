"use client"

import { useEffect } from "react";
import { useRouter } from "next/navigation"
import { setAccessToken } from "@/app/_lib/auth";

export default function CallbackPage() {
    const router = useRouter();

    useEffect(() => {
        const params = new URLSearchParams(window.location.search);
        const token = params.get("access_token");
        const isNewUser = params.get("new_user") === "true";        
        
        if (token) {
            setAccessToken(token);
            router.replace(isNewUser ? "/my_account" : "/events");
        } else {
            router.replace("/");
        }
    })

    return (
        <div>Redirecting...</div>
    );
}